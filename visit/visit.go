package visit

import (
	"time"
)

type Visit interface {
	SetVisitted(key string, ttl time.Duration) error
	UnsetVisitted(key string) error
	IsVisited(key string) bool
}
