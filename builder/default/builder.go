package default_builder

import (
	"time"

	"github.com/superjcd/gocrawler/v1/fetcher"
	"github.com/superjcd/gocrawler/v1/parser"
	"github.com/superjcd/gocrawler/v1/scheduler/nsq"
	"github.com/superjcd/gocrawler/v1/store/mongo"
	"github.com/superjcd/gocrawler/v1/ua"
	"github.com/superjcd/gocrawler/v1/worker"
	"golang.org/x/time/rate"
)

type DefaultWorkerBuilderConfig struct {
	name                 string
	workers              int
	retries              int
	fetch_timeout        int
	save_request_data    bool
	max_run_time_seconds int
	nsqd_addr            string
	nsqlookup_addr       string
	nsq_topic_name       string
	nsq_channel_name     string
	mongo_uri            string
	mongo_database       string
	mongo_collection     string
	limit_rate           int
	buffer_szie          int
	auto_flush_interval  int
}

// defaults
const (
	WORKERS              = 50
	RETRIES              = 5
	FETCH_TIMEOUT        = 10
	SAVE_REQUEST_DATA    = true
	MAX_RUN_TIME_SECONDS = 360
	NSQD_ADDR            = "localhost:4150"
	NSQLOOKUP_ADDR       = "localhost:4161"
	NSQ_TOPIC_NAME       = "gocralwer"
	NSQ_CHANNEL_NAME     = "default"
	MONGO_URI            = "mongodb://localhost:27017"
	MONGO_DATABASE       = "gocrawler"
	MONGO_COLLECTION     = "default"
	LIMIT_RATE           = 50
	BUFFER_SIZE          = 100
	AUTO_FLUSH_INTERVAL  = 10
)

func (bc *DefaultWorkerBuilderConfig) Name(name string) *DefaultWorkerBuilderConfig {
	bc.name = name
	return bc
}

func (bc *DefaultWorkerBuilderConfig) Workers(n int) *DefaultWorkerBuilderConfig {
	bc.workers = n
	return bc
}

func (bc *DefaultWorkerBuilderConfig) Retries(n int) *DefaultWorkerBuilderConfig {
	bc.retries = n
	return bc
}

func (bc *DefaultWorkerBuilderConfig) TimeOut(seconds int) *DefaultWorkerBuilderConfig {
	bc.fetch_timeout = seconds
	return bc
}

func (bc *DefaultWorkerBuilderConfig) SaveRequestData(save bool) *DefaultWorkerBuilderConfig {
	bc.save_request_data = save
	return bc
}

func (bc *DefaultWorkerBuilderConfig) MaxRunTime(seconds int) *DefaultWorkerBuilderConfig {
	bc.max_run_time_seconds = seconds
	return bc
}

func (bc *DefaultWorkerBuilderConfig) NsqScheduler(nsqd_addr, lookup_addr, topic, channel string) *DefaultWorkerBuilderConfig {
	bc.nsqd_addr = nsqd_addr
	bc.nsqlookup_addr = lookup_addr
	bc.nsq_topic_name = topic
	bc.nsq_channel_name = channel
	return bc
}

func (bc *DefaultWorkerBuilderConfig) MongoDb(uri, database, collection string) *DefaultWorkerBuilderConfig {
	bc.mongo_uri = uri
	bc.mongo_database = database
	bc.mongo_collection = collection
	return bc
}

func (bc *DefaultWorkerBuilderConfig) LimitRate(rate int) *DefaultWorkerBuilderConfig {
	bc.limit_rate = rate
	return bc
}

func (bc *DefaultWorkerBuilderConfig) BufferSize(size int) *DefaultWorkerBuilderConfig {
	bc.buffer_szie = size
	return bc
}

func (bc *DefaultWorkerBuilderConfig) AutoFlushInterval(interval int) *DefaultWorkerBuilderConfig {
	bc.auto_flush_interval = interval
	return bc
}

func (bc *DefaultWorkerBuilderConfig) Build(parser parser.Parser, opts ...worker.Option) worker.Worker {
	if bc.workers == 0 {
		bc.workers = WORKERS
	}
	if bc.retries == 0 {
		bc.retries = RETRIES
	}
	if bc.fetch_timeout == 0 {
		bc.fetch_timeout = FETCH_TIMEOUT
	}

	if bc.max_run_time_seconds == 0 {
		bc.max_run_time_seconds = MAX_RUN_TIME_SECONDS
	}
	if bc.nsqd_addr == "" {
		bc.nsqd_addr = NSQD_ADDR
	}
	if bc.nsqlookup_addr == "" {
		bc.nsqlookup_addr = NSQLOOKUP_ADDR
	}
	if bc.nsq_channel_name == "" {
		bc.nsq_channel_name = NSQ_CHANNEL_NAME
	}
	if bc.nsq_topic_name == "" {
		if bc.name != "" {
			bc.nsq_topic_name = bc.name
		} else {
			bc.nsq_topic_name = NSQ_TOPIC_NAME
		}
	}
	if bc.mongo_uri == "" {
		bc.mongo_uri = MONGO_URI
	}
	if bc.mongo_database == "" {
		if bc.name != "" {
			bc.mongo_database = bc.name
		} else {
			bc.mongo_database = MONGO_DATABASE
		}
	}
	if bc.mongo_collection == "" {
		bc.mongo_collection = MONGO_COLLECTION
	}
	if bc.limit_rate == 0 {
		bc.limit_rate = LIMIT_RATE
	}
	if bc.buffer_szie == 0 {
		bc.buffer_szie = BUFFER_SIZE
	}
	if bc.auto_flush_interval == 0 {
		bc.auto_flush_interval = AUTO_FLUSH_INTERVAL
	}
	fetcher := fetcher.NewFectcher(time.Duration(bc.fetch_timeout)*time.Second, fetcher.WithUaGetter(ua.NewDefaultUAGetter()))
	scheduler := nsq.NewNsqScheduler(bc.nsq_topic_name, bc.nsq_channel_name, bc.nsqd_addr, bc.nsqlookup_addr)

	limiter := rate.NewLimiter(rate.Limit(bc.limit_rate), 1)
	storage := mongo.NewBufferedMongoStorage(bc.mongo_uri,
		bc.mongo_database,
		bc.mongo_collection,
		bc.buffer_szie,
		time.Duration(bc.auto_flush_interval)*time.Second)

	worker := worker.NewWorker(bc.name,
		bc.workers,
		bc.retries,
		bc.save_request_data,
		time.Second*time.Duration(bc.max_run_time_seconds),
		worker.WithScheduler(scheduler),
		worker.WithFetcher(fetcher),
		worker.WithLimiter(limiter),
		worker.WithParser(parser),
		worker.WithStore(storage),
	)

	for _, opt := range opts {
		opt(&worker.Options)
	}
	return worker
}
