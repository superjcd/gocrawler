package proxy

import (
	"net/http"
	"net/url"

	"github.com/superjcd/gocrawler/health"
)

type ProxyGetter interface {
	health.HealthChecker
	Get(*http.Request) (*url.URL, error)
}

var _ ProxyGetter = (*randomFixedProxyGetter)(nil)

type randomFixedProxyGetter struct {
	Urls []string
}

func NewRandomFixedProxyGetter(urls ...string) *randomFixedProxyGetter {
	return &randomFixedProxyGetter{Urls: urls}
}

func (p *randomFixedProxyGetter) Get(*http.Request) (*url.URL, error) {
	return nil, nil
}

func (p *randomFixedProxyGetter) Health() (bool, map[string]any) {
	return true, nil
}
