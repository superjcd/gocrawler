package worker

import (
	"context"
	"net/http"
	"time"

	"github.com/superjcd/gocrawler/fetcher"
	"github.com/superjcd/gocrawler/parser"
	"github.com/superjcd/gocrawler/request"
	"github.com/superjcd/gocrawler/scheduler"
	"github.com/superjcd/gocrawler/store"
	"github.com/superjcd/gocrawler/visit"
	"golang.org/x/time/rate"
)

type Options struct {
	Scheduler         scheduler.Scheduler
	Limiter           *rate.Limiter
	UseVisit          bool
	Visiter           visit.Visit
	VisiterTTL        time.Duration
	Fetcher           fetcher.Fetcher
	Parser            parser.Parser
	Store             store.Storage
	Duration          time.Duration
	AddtionalHashKeys []string

	BeforeRequestHook BeforeRequestHook
	AfterRequestHook  AfterRequestHook
	BeforeSaveHook    BeforeSaveHook
	AfterSaveHook     AfterSaveHook
}

// lifecycle hooks

type BeforeRequestHook func(context.Context, *request.Request) (Signal, error)

type AfterRequestHook func(context.Context, *http.Response) (Signal, error)

type BeforeSaveHook func(context.Context, *parser.ParseResult) (Signal, error)

type AfterSaveHook func(context.Context, *parser.ParseResult) (Signal, error)

type Option func(opts *Options)

func WithScheduler(s scheduler.Scheduler) Option {
	return func(opts *Options) {
		opts.Scheduler = s
	}
}

func WithStore(store store.Storage) Option {
	return func(opts *Options) {
		opts.Store = store
	}
}

func WithFetcher(fetcher fetcher.Fetcher) Option {
	return func(opts *Options) {
		opts.Fetcher = fetcher
	}
}

func WithLimiter(limiter *rate.Limiter) Option {
	return func(opts *Options) {
		opts.Limiter = limiter
	}
}

func WithVisiter(v visit.Visit, ttl time.Duration) Option {
	return func(opts *Options) {
		opts.Visiter = v
		opts.UseVisit = true
		opts.VisiterTTL = ttl

	}
}

func WithDuration(duration time.Duration) Option {
	return func(opts *Options) {
		opts.Duration = duration
	}
}

func WithParser(p parser.Parser) Option {
	return func(opts *Options) {
		opts.Parser = p
	}
}

func WithAddtionalHashKeys(keys []string) Option {
	return func(opts *Options) {
		opts.AddtionalHashKeys = keys
	}
}

func WithBeforeRequestHook(h BeforeRequestHook) Option {
	return func(opts *Options) {
		opts.BeforeRequestHook = h
	}
}

func WithAfterRequestHook(h AfterRequestHook) Option {
	return func(opts *Options) {
		opts.AfterRequestHook = h
	}
}

func WithBeforeSaveHook(h BeforeSaveHook) Option {
	return func(opts *Options) {
		opts.BeforeSaveHook = h
	}
}

func WithAfterSaveHook(h AfterSaveHook) Option {
	return func(opts *Options) {
		opts.AfterSaveHook = h
	}
}
