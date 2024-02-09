package store

import (
	"github.com/superjcd/gocrawler/parser"
)

type Storage interface {
	Save(datas ...parser.ParseItem) error
}
