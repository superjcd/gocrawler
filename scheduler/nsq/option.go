package nsq

type options struct {
	OtherSchedulers map[string]*nsqScheduler
}

type Option func(opts *options)

func WithOtherScheduler(name string, scheduler *nsqScheduler) Option {
	return func(opts *options) {
		opts.OtherSchedulers[name] = scheduler
	}
}
