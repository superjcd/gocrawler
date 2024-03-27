package worker

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/superjcd/gocrawler/health"
	"github.com/superjcd/gocrawler/parser"
	"github.com/superjcd/gocrawler/request"
	"github.com/superjcd/gocrawler/scheduler"
)

type Worker interface {
	health.HealthChecker
	Name() string
	Run()
	BeforeRequest(context.Context, *request.Request) (Signal, error)
	AfterRequest(context.Context, *http.Response) (Signal, error)
	BeforeSave(context.Context, *parser.ParseResult) (Signal, error)
	AfterSave(context.Context, *parser.ParseResult) (Signal, error)
}

type worker struct {
	name            string
	Workers         int
	MaxRetries      int
	SaveRequestData bool
	MaxRunTime      time.Duration
	options
}

var _ Worker = (*worker)(nil)

func NewWorker(name string, workers, retries int, saveRequestData bool, maxRunTime time.Duration, opts ...Option) *worker {
	options := options{}

	for _, opt := range opts {
		opt(&options)
	}
	w := &worker{name: name, Workers: workers, MaxRetries: retries, SaveRequestData: saveRequestData, MaxRunTime: maxRunTime}
	w.options = options

	go w.Scheduler.Schedule()

	return w
}

func (w *worker) BeforeRequest(ctx context.Context, req *request.Request) (Signal, error) {
	var sig Signal
	if w.BeforeRequestHook != nil {
		return w.BeforeRequestHook(ctx, req)
	}
	sig |= DummySignal
	return sig, nil
}

func (w *worker) AfterRequest(ctx context.Context, resp *http.Response) (Signal, error) {
	var sig Signal
	if w.AfterRequestHook != nil {
		return w.AfterRequestHook(ctx, resp)
	}
	sig |= DummySignal
	return sig, nil
}

func (w *worker) BeforeSave(ctx context.Context, par *parser.ParseResult) (Signal, error) {
	var sig Signal
	if w.BeforeSaveHook != nil {
		return w.BeforeSaveHook(ctx, par)
	}
	sig |= DummySignal
	return sig, nil
}

func (w *worker) AfterSave(ctx context.Context, par *parser.ParseResult) (Signal, error) {
	var sig Signal

	if w.AfterSaveHook != nil {
		return w.AfterSaveHook(ctx, par)
	}
	sig |= DummySignal
	return sig, nil
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
	var sig Signal
	var err error

Loop:
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
				continue Loop
			}
		}
		originReq := req

		ctx := context.Background()
		ctx = context.WithValue(ctx, request.RequestDataCtxKey{}, req.Data)

		sig, err = w.BeforeRequest(ctx, req)

		switch w.dealSignal(sig, err, req, originReq) {
		case ContinueLoop:
			continue Loop
		case BreakLoop:
			break Loop
		}

		resp, err := w.Fetcher.Fetch(ctx, req)
		if err != nil {
			log.Printf("fetch failed: %v", err)
			w.retry(req, originReq)
			continue
		}

		if w.AfterRequestHook != nil {
			sig, err = w.AfterRequestHook(ctx, resp)
			switch w.dealSignal(sig, err, req, originReq) {
			case ContinueLoop:
				continue Loop
			case BreakLoop:
				break Loop
			}
		} else {
			if resp.StatusCode != http.StatusOK {
				w.retry(req, originReq)
				continue
			}
		}

		// Parse
		parseResult, err := w.Parser.Parse(ctx, resp)
		if err != nil {
			log.Printf("parse failed for request: %s, error: %v", req.URL, err)
			w.retry(req, originReq)
			continue
		}

		// New Requests
		if parseResult.Requests != nil && len(parseResult.Requests) > 0 {
			for _, req := range parseResult.Requests {
				if !req.IsSecondary {
					w.Scheduler.Push(scheduler.TYP_PUSH_SCHEDULER, req)
					continue
				}
				if secondScheduler := w.Scheduler.SecondScheduler(); secondScheduler != nil {
					secondScheduler.Push(scheduler.TYP_PUSH_SCHEDULER, req)
				}
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

			sig, err = w.BeforeSave(ctx, parseResult)
			switch w.dealSignal(sig, err, req, originReq) {
			case ContinueLoop:
				continue Loop
			case BreakLoop:
				break Loop
			}

			if err := w.Store.Save(ctx, parseResult.Items...); err != nil {
				log.Printf("item saved failed err: %v;items: ", err)
				continue
			}

			sig, err = w.AfterSave(ctx, parseResult)
			switch w.dealSignal(sig, err, req, originReq) {
			case ContinueLoop:
				continue Loop
			case BreakLoop:
				break Loop
			}
		}
		if w.UseVisit {
			w.Visiter.SetVisitted(reqKey, w.VisiterTTL)
		}

	}
}

func (w *worker) Name() string {
	return w.name
}

func (w *worker) Health() (bool, map[string]any) {
	health := true
	healthDetails := map[string]any{}
	fetcherHealthStatus, fetcherHealthDetails := w.Fetcher.Health()
	health = health && fetcherHealthStatus
	healthDetails["fetcher"] = fetcherHealthDetails
	return health, healthDetails
}

func (w *worker) retry(req, originReq *request.Request) {
	if req.Retry < w.MaxRetries {
		originReq.Retry += 1
		w.Scheduler.Push(scheduler.TYP_PUSH_SCHEDULER, originReq)
	} else {
		log.Printf("too many retries for request:%s, exceed max retries: %d", req.URL, w.MaxRetries)
	}
}

func (w *worker) dealSignal(sig Signal, err error, req, originReq *request.Request) int {
	if sig&DummySignal != 0 {
		return KeepGoing
	}
	if sig&ContinueWithRetrySignal != 0 {
		w.retry(req, originReq)
		return ContinueLoop
	}
	if sig&ContinueWithoutRetrySignal != 0 {
		return ContinueLoop
	}
	if sig&BreakWithPanicSignal != 0 {
		panic(err)
	}
	if sig&BreakWithoutPanicSignal != 0 {
		return BreakLoop
	}
	return KeepGoing
}
