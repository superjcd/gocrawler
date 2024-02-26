package cookie

import (
	"net/http/cookiejar"

	"github.com/superjcd/gocrawler/health"
)

type CoookieGetter interface {
	health.HealthChecker
	Get() (*cookiejar.Jar, error)
}
