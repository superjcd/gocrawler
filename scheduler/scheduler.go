package scheduler

import "github.com/superjcd/gocrawler/request"

type Scheduler interface {
	Pull() *request.Request
	Push(typ int, reqs ...*request.Request)
	Schedule()
	NamedSchedulers() map[string]Scheduler
}
