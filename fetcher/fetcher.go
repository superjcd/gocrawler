package fetcher

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/superjcd/gocrawler/request"
)

type Fetcher interface {
	Fetch(ctx context.Context, req *request.Request) (*http.Response, error)
}

type fectcher struct {
	Cli *http.Client
	options
}

var _ Fetcher = (*fectcher)(nil)

func NewFectcher(timeOut time.Duration, opts ...Option) *fectcher {
	options := options{}
	for _, opt := range opts {
		opt(&options)
	}

	var transport *http.Transport
	if options.transport != nil {
		transport = options.transport
	} else {
		transport = http.DefaultTransport.(*http.Transport)
		transport.DisableKeepAlives = true
	}

	client := &http.Client{Transport: transport, Timeout: timeOut}
	f := &fectcher{Cli: client}

	return f
}

func (f *fectcher) Fetch(ctx context.Context, r *request.Request) (resp *http.Response, err error) {
	if f.cookieGetter != nil {
		jar, err := f.cookieGetter.Get(ctx)
		if err != nil {
			return nil, err
		}
		f.Cli.Jar = jar
	}

	req, err := http.NewRequest(r.Method, r.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("get url failed: %w", err)
	}

	if f.uaGetter != nil {
		ua, err := f.uaGetter.Get(ctx)

		if err != nil {
			return nil, fmt.Errorf("get ua failed: %w", err)
		}
		req.Header.Set("User-Agent", ua)
	}

	resp, err = f.Cli.Do(req)

	if err != nil {
		return nil, err
	}

	return
}

func (f *fectcher) Health() (bool, map[string]any) {
	return true, nil
}
