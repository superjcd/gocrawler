package worker

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/superjcd/gocrawler/health"
	"github.com/superjcd/gocrawler/parser"
	"github.com/superjcd/gocrawler/request"
	"github.com/superjcd/gocrawler/scheduler/nsq"
)

type Worker interface {
	health.HealthChecker
	Name() string
	ID() int
	Run()
	BeforeRequest(context.Context, *request.Request)
	BeforeSave(context.Context, *parser.ParseResult)
}

type worker struct {
	name            string
	id              int
	Workers         int
	MaxRetries      int
	SaveRequestData bool
	MaxRunTime      time.Duration
	options
}

var _ Worker = (*worker)(nil)

func NewWorker(name string, id, workers, retries int, saveRequestData bool, maxRunTime time.Duration, opts ...Option) *worker {
	options := options{}

	for _, opt := range opts {
		opt(&options)
	}
	w := &worker{name: name, id: id, Workers: workers, MaxRetries: retries, SaveRequestData: saveRequestData, MaxRunTime: maxRunTime}
	w.options = options

	go w.Scheduler.Schedule()

	return w
}

func (w *worker) BeforeRequest(ctx context.Context, req *request.Request) {
	if w.BeforeRequestHook != nil {
		err := w.BeforeRequestHook(ctx, req)
		if err != nil {
			panic(err)
		}
	}
}

func (w *worker) BeforeSave(ctx context.Context, par *parser.ParseResult) {
	if w.BeforeSaveHook != nil {
		err := w.BeforeSaveHook(ctx, par)
		if err != nil {
			panic(err)
		}
	}

}

func (w *worker) AfterSave(ctx context.Context, par *parser.ParseResult) {
	if w.AfterSaveHook != nil {
		err := w.AfterSaveHook(ctx, par)
		if err != nil {
			panic(err)
		}
	}

}

func (w *worker) Run() {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		w.MaxRunTime,
	)
	defer cancel()
	for i := 0; i < w.Workers; i++ {
		go singleRun(w)
	}

	<-ctx.Done()
}

func singleRun(w *worker) {
	for {
		w.Limiter.Wait(context.TODO())
		req := w.Scheduler.Pull()
		if req == nil {
			continue
		}
		var reqKey string
		if w.UseVisit {
			if w.AddtionalHashKeys == nil {
				reqKey = req.Hash()
			} else {
				reqKey = req.Hash(w.AddtionalHashKeys...)
			}

			if w.Visiter.IsVisited(reqKey) {
				continue
			}
		}
		originReq := req

		// Fetch
		w.BeforeRequest(context.Background(), req)
		resp, err := w.Fetcher.Fetch(req)

		if err != nil {
			log.Printf("request failed: %v", err)
			if req.Retry < w.MaxRetries {
				originReq.Retry += 1
				w.Scheduler.Push(nsq.NSQ_PUSH, originReq)
			} else {
				log.Printf("too many fetch failures for request:%s, exceed max retries: %d", req.URL, w.MaxRetries)
			}
			continue
		}

		if resp.StatusCode != http.StatusOK {
			originReq.Retry += 1
			w.Scheduler.Push(nsq.NSQ_PUSH, originReq)
			continue
		}

		// Parse
		parseResult, err := w.Parser.Parse(resp)
		if err != nil {
			log.Printf("parse failed for request: %s, error: %v", req.URL, err)
			originReq.Retry += 1
			w.Scheduler.Push(nsq.NSQ_PUSH, originReq)
			continue
		}

		// New Requests
		if parseResult.Requests != nil && len(parseResult.Requests) > 0 {
			for _, req := range parseResult.Requests {
				w.Scheduler.Push(nsq.NSQ_PUSH, req)
			}
		}

		// Save
		if parseResult.Items != nil && len(parseResult.Items) > 0 {
			if w.SaveRequestData {
				for _, p_item := range parseResult.Items {
					for dk, dv := range req.Data {
						p_item[dk] = dv
					}
				}
			}

			w.BeforeSave(context.Background(), parseResult)
			if err := w.Store.Save(parseResult.Items...); err != nil {
				log.Printf("item saved failed err: %v;items: ", err)
				continue
			}
			w.AfterSave(context.Background(), parseResult)
		}
		if w.UseVisit {
			w.Visiter.SetVisitted(reqKey, w.VisiterTTL)
		}

	}
}

func (w *worker) Name() string {
	return w.name
}
func (w *worker) ID() int {
	return w.id
}
func (w *worker) Health() (bool, map[string]any) {
	health := true
	healthDetails := map[string]any{}
	fetcherHealthStatus, fetcherHealthDetails := w.Fetcher.Health()
	health = health && fetcherHealthStatus
	healthDetails["fetcher"] = fetcherHealthDetails
	return health, healthDetails
}
