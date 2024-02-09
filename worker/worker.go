package worker

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/superjcd/gocrawler/request"
	"github.com/superjcd/gocrawler/scheduler/nsq"
)

type Worker interface {
	Run()
	ModifyRequest(req *request.Request)
}

type worker struct {
	Workers         int
	MaxRetries      int
	SaveRequestData bool
	options
}

var _ Worker = (*worker)(nil)

func NewWorker(workers, retries int, saveRequestData bool, opts ...Option) *worker {
	options := options{}

	for _, opt := range opts {
		opt(&options)
	}
	w := &worker{Workers: workers, MaxRetries: retries, SaveRequestData: saveRequestData}
	w.options = options

	go w.Scheduler.Schedule()

	return w
}

func (w *worker) ModifyRequest(req *request.Request) {
	if w.RequsetModifier != nil {
		err := w.RequsetModifier(req)
		if err != nil {
			panic(err)
		}
	}
}

func (w *worker) Run() {
	for i := 0; i < w.Workers; i++ {
		go singleRun(w)
	}

	time.Sleep(time.Minute * 5) // TODO
}

func singleRun(w *worker) {
	for {
		w.Limiter.Wait(context.TODO())
		req := w.Scheduler.Pull()
		if req == nil {
			continue
		}
		var reqKey string
		if w.AddtionalHashKeys == nil {
			reqKey = req.Hash()
		} else {
			reqKey = req.Hash(w.AddtionalHashKeys...)
		}

		if w.UseVist && w.Vister.IsVisited(reqKey) {
			continue
		}
		originReq := req

		// Fetch
		w.ModifyRequest(req)
		resp, err := w.Fetcher.Fetch(req)

		if err != nil {
			log.Printf("request failed: %v", err)
			if req.Retry < w.MaxRetries {
				originReq.Retry += 1
				w.Scheduler.Push(nsq.NSQ_PUSH, originReq)
			}
			log.Printf("failure times for request:%s ecceed max retries: %d", req.URL, w.MaxRetries)
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
			log.Printf("parsed failed for request: %s", req.URL)
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
			//
			if err := w.Store.Save(parseResult.Items...); err != nil {
				log.Printf("item saved failed err: %v;items: ", err)
				continue
			}

		}
		if w.UseVist {
			w.Vister.SetVisitted(reqKey, w.VisterTTL)
		}

	}
}
