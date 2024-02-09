package proxy

import (
	"net/http"
	"net/url"
)

type ProxyGetter interface {
	Get(*http.Request) (*url.URL, error)
}

type randomFixedProxyGetter struct {
	Urls []string
}

func NewRandomFixedProxyGetter(urls ...string) *randomFixedProxyGetter {
	return &randomFixedProxyGetter{Urls: urls}
}

func (p *randomFixedProxyGetter) Get(*http.Request) (*url.URL, error) {
	return nil, nil
}
