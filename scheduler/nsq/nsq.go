package nsq

import (
	"encoding/json"
	"log"

	"github.com/nsqio/go-nsq"
	"github.com/superjcd/gocrawler/request"
	"github.com/superjcd/gocrawler/scheduler"
)

type nsqScheduler struct {
	workerCh       chan *request.Request
	nsqLookupdAddr string
	topicName      string
	channelName    string
	nsqConsumer    *nsq.Consumer
	nsqProducer    *nsq.Producer
	options
}

type nsqMessageHandler struct {
	s *nsqScheduler
}

func (h *nsqMessageHandler) HandleMessage(m *nsq.Message) error {
	var err error
	if len(m.Body) == 0 {
		return nil
	}

	processMessage := func(mb []byte) error {
		var req request.Request
		if err = json.Unmarshal(mb, &req); err != nil {
			return err

		}
		h.s.Push(scheduler.TYP_PUSH_CHANNEL, &req)
		return nil
	}

	err = processMessage(m.Body)

	return err

}

var _ scheduler.Scheduler = (*nsqScheduler)(nil)

func NewNsqScheduler(topicName, channelName, nsqAddr, nsqLookupdAddr string, opts ...Option) *nsqScheduler {
	options := options{}

	for _, opt := range opts {
		opt(&options)
	}

	nsqConfig := nsq.NewConfig()

	nsqConsumer, err := nsq.NewConsumer(topicName, channelName, nsqConfig)

	if err != nil {
		log.Fatal(err)
	}

	nsqProducer, err := nsq.NewProducer(nsqAddr, nsqConfig)

	if err != nil {
		log.Fatal(err)
	}

	workerCh := make(chan *request.Request)

	return &nsqScheduler{workerCh: workerCh,
		topicName:      topicName,
		channelName:    channelName,
		nsqLookupdAddr: nsqLookupdAddr,
		nsqConsumer:    nsqConsumer,
		nsqProducer:    nsqProducer,
		options:        options,
	}
}

func (s *nsqScheduler) Pull() *request.Request {
	req := <-s.workerCh
	return req
}

func (s *nsqScheduler) Push(typ int, reqs ...*request.Request) {
	switch typ {
	case scheduler.TYP_PUSH_CHANNEL:
		for _, req := range reqs {
			s.workerCh <- req
		}
	case scheduler.TYP_PUSH_SCHEDULER:
		for _, req := range reqs {
			msg, err := json.Marshal(req)
			if err != nil {
				log.Printf("push msg to nsq failed")
			}
			s.nsqProducer.Publish(s.topicName, msg)

		}
	default:
		log.Fatal("wrong push type")
	}

}

func (s *nsqScheduler) Schedule() {
	s.nsqConsumer.AddHandler(&nsqMessageHandler{s: s})
	if err := s.nsqConsumer.ConnectToNSQLookupd(s.nsqLookupdAddr); err != nil {
		log.Fatal(err)
	}

}

func (s *nsqScheduler) SecondScheduler() scheduler.Scheduler {
	return s.secondScheduler
}

func (s *nsqScheduler) Stop() {
	s.nsqConsumer.Stop()
}
