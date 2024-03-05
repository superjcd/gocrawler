package cookie

import (
	"context"
	"net/http/cookiejar"

	"github.com/superjcd/gocrawler/health"
)

type CoookieGetter interface {
	health.HealthChecker
	Get(context.Context) (*cookiejar.Jar, error)
}
