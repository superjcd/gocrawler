package fetcher

import (
	"net/http"

	"github.com/superjcd/gocrawler/cookie"
	"github.com/superjcd/gocrawler/proxy"
	"github.com/superjcd/gocrawler/ua"
)

// proxyGetter proxy.ProxyGetter, cookieGetter cookie.CoookieGetter, uaGetter ua.UaGetter
type options struct {
	transport    *http.Transport
	proxyGetter  proxy.ProxyGetter
	cookieGetter cookie.CookieGetter
	uaGetter     ua.UaGetter
	headers      map[string]string
}

type Option func(opts *options)

func WithTransport(transport *http.Transport) Option {
	return func(opts *options) {
		opts.transport = transport
	}
}

func WithProxyGetter(proxyGetter proxy.ProxyGetter) Option {
	return func(opts *options) {
		opts.proxyGetter = proxyGetter
	}
}

func WithCookieGetter(cookieGetter cookie.CookieGetter) Option {
	return func(opts *options) {
		opts.cookieGetter = cookieGetter
	}
}

func WithHeaders(headers map[string]string) Option {
	return func(opts *options) {
		opts.headers = headers
	}
}

func WithUaGetter(uaGetter ua.UaGetter) Option {
	return func(opts *options) {
		opts.uaGetter = uaGetter
	}
}
