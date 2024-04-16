package redis

import (
	"time"

	"github.com/go-redis/redis"
	"github.com/superjcd/gocrawler/v1/visit"
)

type RedisVisiter struct {
	VisitedKeyPrefix string
	RCli             *redis.Client
}

var _ visit.Visit = (*RedisVisiter)(nil)

func NewRedisVisiter(r_config redis.Options, prefixKey string) *RedisVisiter {
	rc := &RedisVisiter{VisitedKeyPrefix: prefixKey}
	rc.RCli = redis.NewClient(&r_config)
	return rc
}

func (rc *RedisVisiter) SetVisitted(key string, ttl time.Duration) error {
	if rc.VisitedKeyPrefix == "" {
		key = "gocrawler:" + key
	} else {
		key = rc.VisitedKeyPrefix + key
	}
	_, err := rc.RCli.Set(key, "", ttl).Result()

	return err
}

func (rc *RedisVisiter) UnsetVisitted(key string) error {
	_, err := rc.RCli.Del(key).Result()

	return err
}

func (rc *RedisVisiter) IsVisited(key string) bool {
	_, err := rc.RCli.Get(key).Result()

	return err == nil
}
