package nsq

import "github.com/superjcd/gocrawler/scheduler"

type options struct {
	namedScheduler map[string]scheduler.Scheduler
}

type Option func(opts *options)

func WithNamedSchedulers(name string, scheduler *nsqScheduler) Option {
	return func(opts *options) {
		opts.namedScheduler[name] = scheduler
	}
}
