package mongo

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"github.com/qiniu/qmgo"
	"github.com/superjcd/gocrawler/counter"
	"github.com/superjcd/gocrawler/parser"
	"github.com/superjcd/gocrawler/store"
)

type bufferedMongoStorage struct {
	L            *sync.Mutex
	Cli          *qmgo.QmgoClient
	buf          []parser.ParseItem
	bufSize      int
	counter      counter.Counter
	taskKeyField string
}

var _ store.Storage = (*bufferedMongoStorage)(nil)

type BufferedMongoStorageOption func(s *bufferedMongoStorage)

func WithRedisCounter(r_config redis.Options, ttl time.Duration, counterPrefix, keyField string) BufferedMongoStorageOption {
	return func(s *bufferedMongoStorage) {
		redisCounter := counter.NewRedisTaskCounters(r_config, ttl, counterPrefix, keyField)
		s.counter = redisCounter
		s.taskKeyField = keyField
	}
}

const DEFAULT_BUFFER_SIZE = 100

func NewBufferedMongoStorage(uri, database, collection string, bufferSize int, autoFlushInterval time.Duration, opts ...BufferedMongoStorageOption) *bufferedMongoStorage {
	ctx := context.Background()
	cli, err := qmgo.Open(ctx, &qmgo.Config{Uri: uri,
		Database: database,
		Coll:     collection})
	if err != nil {
		panic(err)
	}

	if bufferSize == 0 {
		bufferSize = DEFAULT_BUFFER_SIZE
	}

	buf := make([]parser.ParseItem, 0, bufferSize)

	var l sync.Mutex

	store := &bufferedMongoStorage{Cli: cli, L: &l, bufSize: bufferSize, buf: buf}
	for _, option := range opts {
		option(store)
	}

	ticker := time.NewTicker(autoFlushInterval)

	go func() {
		for t := range ticker.C {
			log.Printf("auto flush triggered at %v", t)
			store.flush()
		}

	}()

	return store

}

func (s *bufferedMongoStorage) Save(items ...parser.ParseItem) error {
	s.L.Lock()
	defer s.L.Unlock()

	for {
		if len(items) > s.bufSize {
			return fmt.Errorf("number of items too large(larger than the max bufSize), either increase storage bufSize or decrease number of items")
		}

		if len(items) > (s.bufSize - len(s.buf)) {
			if err := s.flush(); err != nil {
				return err
			}

		} else {
			s.buf = append(s.buf, items...)
			break
		}
	}

	return nil
}

func (s *bufferedMongoStorage) flush() error {
	if len(s.buf) == 0 {
		return nil
	}
	err := s.insertManyTOMongo(s.buf...)
	if err != nil {
		return err
	}

	if s.counter != nil {
		tc := s.collectTaskCounts(s.buf)
		s.count(tc)
	}
	// update buffer to an empty buffer
	s.buf = make([]parser.ParseItem, 0, s.bufSize)
	log.Printf("Flushed")
	return nil
}

func (s *bufferedMongoStorage) collectTaskCounts(buf []parser.ParseItem) (tc map[string]int64) {
	tc = make(map[string]int64, 128)
	for _, item := range buf {
		if taskId, ok := item[s.taskKeyField]; !ok {
			panic(fmt.Errorf("`%s` not found in Parseitem, if you want to use the task counter, then `%s` embeded must be stuffed in the ParsedItem", s.taskKeyField, s.taskKeyField))
		} else {
			switch v := taskId.(type) {
			case string:
				tc[v] += 1
			default:
				panic("`taskId` must be string")
			}
		}

	}
	return tc
}

func (s *bufferedMongoStorage) count(tc map[string]int64) {
	for k, v := range tc {
		s.counter.Incr(k, v)
	}
}

func (s *bufferedMongoStorage) insertManyTOMongo(items ...parser.ParseItem) error {
	if result, err := s.Cli.Collection.InsertMany(context.Background(), items); err != nil {
		return err
	} else {
		log.Printf("%d objects saved", len(result.InsertedIDs))
		return nil
	}
}

func (s *bufferedMongoStorage) Close() error {
	s.flush()
	return s.Cli.Close(context.Background())
}
