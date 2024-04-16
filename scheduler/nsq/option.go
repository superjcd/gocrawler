package nsq

import (
	"github.com/superjcd/gocrawler/v1/scheduler"
)

type options struct {
	secondScheduler scheduler.Scheduler
}

type Option func(opts *options)

func WithSecondScheduler(scheduler *nsqScheduler) Option {
	return func(opts *options) {
		opts.secondScheduler = scheduler
	}
}
