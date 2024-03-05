package fetcher

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/superjcd/gocrawler/cookie"
	"github.com/superjcd/gocrawler/health"
	"github.com/superjcd/gocrawler/proxy"
	"github.com/superjcd/gocrawler/request"
	"github.com/superjcd/gocrawler/ua"
)

// 考虑添加一个health checkerm, 添加一个HeathCheck

type Fetcher interface {
	health.HealthChecker
	Fetch(ctx context.Context, req *request.Request) (*http.Response, error)
}

type fectcher struct {
	Cli          *http.Client
	CookieGetter cookie.CoookieGetter
	ProxyGetter  proxy.ProxyGetter
	UaGetter     ua.UaGetter
}

var _ Fetcher = (*fectcher)(nil)

func NewFectcher(timeOut time.Duration, proxyGetter proxy.ProxyGetter, cookieGetter cookie.CoookieGetter, uaGetter ua.UaGetter) *fectcher {
	tr := http.DefaultTransport.(*http.Transport)
	tr.Proxy = proxyGetter.Get
	tr.DisableKeepAlives = true
	client := &http.Client{Transport: tr, Timeout: timeOut}

	f := &fectcher{
		Cli:          client,
		ProxyGetter:  proxyGetter,
		CookieGetter: cookieGetter,
		UaGetter:     uaGetter,
	}

	return f
}

func (f *fectcher) Fetch(ctx context.Context, r *request.Request) (resp *http.Response, err error) {
	jar, err := f.CookieGetter.Get()
	if err != nil {
		return nil, err
	}
	f.Cli.Jar = jar
	req, err := http.NewRequest(r.Method, r.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("get url failed: %w", err)
	}
	ua, err := f.UaGetter.Get(ctx)

	if err != nil {
		return nil, fmt.Errorf("get ua failed: %w", err)
	}
	req.Header.Set("User-Agent", ua)

	resp, err = f.Cli.Do(req)

	if err != nil {
		return nil, err
	}

	return
}

func (f *fectcher) Health() (bool, map[string]any) {
	// internet health
	health := true
	healthDetails := map[string]any{}
	cookieHealthStatus, cookieHealthDetails := f.CookieGetter.Health()
	health = health && cookieHealthStatus
	healthDetails["cookies"] = cookieHealthDetails
	return health, healthDetails
}
