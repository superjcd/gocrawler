package scheduler

import "github.com/superjcd/gocrawler/request"

const (
	TYP_PUSH_CHANNEL = iota
	TYP_PUSH_SCHEDULER
)

type Scheduler interface {
	Pull() *request.Request
	Push(typ int, reqs ...*request.Request)
	Schedule()
	SecondScheduler() Scheduler
}
