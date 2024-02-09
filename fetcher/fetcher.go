package fetcher

import (
	"fmt"
	"net/http"
	"time"

	"github.com/CUCyber/ja3transport"
	"github.com/superjcd/gocrawler/cookie"
	"github.com/superjcd/gocrawler/proxy"
	"github.com/superjcd/gocrawler/request"
	"github.com/superjcd/gocrawler/ua"
)

type Fetcher interface {
	Fetch(req *request.Request) (*http.Response, error)
}

type fectcher struct {
	Cli          *http.Client
	CookieGetter cookie.CoookieGetter
	UaGetter     ua.UaGetter
}

func NewFectcher(timeOut time.Duration, proxyGetter proxy.ProxyGetter, cookieGetter cookie.CoookieGetter, uaGetter ua.UaGetter) *fectcher {
	// tr := http.DefaultTransport.(*http.Transport)
	tr, _ := ja3transport.NewTransport("771,4865-4866-4867-49195-49199-49196-49200-52393-52392-49171-49172-156-157-47-53-10,0-23-65281-10-11-35-16-5-13-18-51-45-43-27-21,29-23-24,0")
	tr.Proxy = proxyGetter.Get
	tr.DisableKeepAlives = true
	client := &http.Client{Transport: tr, Timeout: timeOut}

	f := &fectcher{
		Cli:          client,
		CookieGetter: cookieGetter,
		UaGetter:     uaGetter,
	}

	return f
}

func (f *fectcher) Fetch(r *request.Request) (resp *http.Response, err error) {
	jar, err := f.CookieGetter.Get()
	if err != nil {
		return nil, err
	}
	f.Cli.Jar = jar
	req, err := http.NewRequest(r.Method, r.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("get url failed: %w", err)
	}
	ua, err := f.UaGetter.Get()

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
