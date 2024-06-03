package redis

import (
	"time"

	"github.com/go-redis/redis"
	"github.com/superjcd/gocrawler/visit"
)

type RedisVisit struct {
	VisitedKeyPrefix string
	RCli             *redis.Client
}

var _ visit.Visit = (*RedisVisit)(nil)

func NewRedisVisit(r_config redis.Options, prefixKey string) *RedisVisit {
	rc := &RedisVisit{VisitedKeyPrefix: prefixKey}
	rc.RCli = redis.NewClient(&r_config)
	return rc
}

func (rc *RedisVisit) SetVisitted(key string, ttl time.Duration) error {
	if rc.VisitedKeyPrefix == "" {
		key = "gocrawler:" + key
	} else {
		key = rc.VisitedKeyPrefix + key
	}
	_, err := rc.RCli.Set(key, "", ttl).Result()

	return err
}

func (rc *RedisVisit) UnsetVisitted(key string) error {
	_, err := rc.RCli.Del(key).Result()

	return err
}

func (rc *RedisVisit) IsVisited(key string) bool {
	_, err := rc.RCli.Get(key).Result()

	return err == nil
}
