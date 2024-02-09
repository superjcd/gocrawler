package worker

import (
	"time"

	"github.com/superjcd/gocrawler/fetcher"
	"github.com/superjcd/gocrawler/parser"
	"github.com/superjcd/gocrawler/request"
	"github.com/superjcd/gocrawler/scheduler"
	"github.com/superjcd/gocrawler/store"
	"github.com/superjcd/gocrawler/visit"
	"golang.org/x/time/rate"
)

type RequsetModifier func(*request.Request) error

type options struct {
	Scheduler         scheduler.Scheduler
	Limiter           *rate.Limiter
	UseVist           bool
	Vister            visit.Visit
	VisterTTL         time.Duration
	Fetcher           fetcher.Fetcher
	Parser            parser.Parser
	Store             store.Storage
	RequsetModifier   RequsetModifier
	Duration          time.Duration
	AddtionalHashKeys []string
}

type Option func(opts *options)

func WithScheduler(s scheduler.Scheduler) Option {
	return func(opts *options) {
		opts.Scheduler = s
	}
}

func WithStore(store store.Storage) Option {
	return func(opts *options) {
		opts.Store = store
	}
}

func WithFetcher(fetcher fetcher.Fetcher) Option {
	return func(opts *options) {
		opts.Fetcher = fetcher
	}
}

func WithRequestModifer(m RequsetModifier) Option {
	return func(opts *options) {
		opts.RequsetModifier = m
	}
}

func WithLimiter(limiter *rate.Limiter) Option {
	return func(opts *options) {
		opts.Limiter = limiter
	}
}

func WithVisiter(v visit.Visit, ttl time.Duration) Option {
	return func(opts *options) {
		opts.Vister = v
		opts.UseVist = true
		opts.VisterTTL = ttl

	}
}

func WithDuration(duration time.Duration) Option {
	return func(opts *options) {
		opts.Duration = duration
	}
}

func WithParser(p parser.Parser) Option {
	return func(opts *options) {
		opts.Parser = p
	}
}

func WithAddtionalHashKeys(keys []string) Option {
	return func(opts *options) {
		opts.AddtionalHashKeys = keys
	}
}
