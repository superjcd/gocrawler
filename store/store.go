package store

import (
	"context"

	"github.com/superjcd/gocrawler/v1/parser"
)

type Storage interface {
	Save(ctx context.Context, datas ...parser.ParseItem) error
}
