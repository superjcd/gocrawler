package parser

import (
	"net/http"

	"github.com/superjcd/gocrawler/request"
)

type ParseItem map[string]interface{}

type ParseResult struct {
	Items    []ParseItem
	Requests []*request.Request
}

type Parser interface {
	Parse(resp *http.Response) (*ParseResult, error)
}
