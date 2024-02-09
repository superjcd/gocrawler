package counter

import (
	"sync/atomic"
	"time"

	"github.com/go-redis/redis"
)

// 任务技计数器
type Counter interface {
	Incr(key string, num int64)
	GetCounterPrefix() string
}

// Redis transactions use optimistic locking.
const (
	maxRetries = 1000
)

type RedisTaskCounters struct {
	prefix string
	RCli   *redis.Client
	TTL    time.Duration
}

func NewRedisTaskCounters(r_config redis.Options, ttl time.Duration, counterPrefix string) *RedisTaskCounters {
	// r_config redis.Options,
	rc := &RedisTaskCounters{TTL: ttl, prefix: counterPrefix}
	rc.RCli = redis.NewClient(&r_config)
	return rc
}

func (c *RedisTaskCounters) GetCounterPrefix() string {
	return c.prefix
}

func (c *RedisTaskCounters) Incr(key string, increment int64) {
	// transaction
	txf := func(tx *redis.Tx) error {
		// Get the current value or zero.
		n, err := tx.Get(key).Int64()
		if err != nil && err != redis.Nil {
			return err
		}

		// Actual operation (local in optimistic lock).
		atomic.AddInt64(&n, increment)

		// Operation is commited only if the watched keys remain unchanged.
		_, err = tx.Pipelined(func(pipe redis.Pipeliner) error {
			pipe.Set(key, n, c.TTL) // time
			return nil
		})
		return err
	}

	// Retry if the key has been changed.
	for i := 0; i < maxRetries; i++ {
		err := c.RCli.Watch(txf, key)
		if err == nil {
			// Success.
			return
		}
		if err == redis.TxFailedErr {
			// Optimistic lock lost. Retry.
			continue
		}
		// TODO: igonore any other error.
		return
	}

}
