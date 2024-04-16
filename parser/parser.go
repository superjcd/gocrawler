package parser

import (
	"context"
	"net/http"

	"github.com/superjcd/gocrawler/v1/request"
)

type ParseItem map[string]interface{}

type ParseResult struct {
	Items    []ParseItem
	Requests []*request.Request
}

type Parser interface {
	Parse(ctx context.Context, resp *http.Response) (*ParseResult, error)
}
