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

type BufferedMongoStorage struct {
	L         *sync.Mutex
	Cli       *qmgo.QmgoClient
	buf       []parser.ParseItem
	bufSize   int
	counter   counter.Counter
	taskField string
}

type BufferedMongoStorageOption func(s *BufferedMongoStorage)

func WithRedisCounter(r_config redis.Options, ttl time.Duration, counterPrefix, keyField string) BufferedMongoStorageOption {
	return func(s *BufferedMongoStorage) {
		redisCounter := counter.NewRedisTaskCounters(r_config, ttl, counterPrefix, keyField)
		s.counter = redisCounter
	}
}

type MongoStorage struct {
	Cli *qmgo.QmgoClient
}

var _ store.Storage = (*MongoStorage)(nil)

const DEFAULT_BUFFER_SIZE = 100

func NewMongoStorage(uri, database, collection string) *MongoStorage {
	ctx := context.Background()
	cli, err := qmgo.Open(ctx, &qmgo.Config{Uri: uri,
		Database: database,
		Coll:     collection}) // counter
	if err != nil {
		panic(err)
	}

	return &MongoStorage{Cli: cli}
}

func (s *MongoStorage) Save(items ...parser.ParseItem) error {
	var result *qmgo.InsertOneResult
	var err error
	for _, item := range items {

		result, err = s.Cli.Collection.InsertOne(context.Background(), item)
		if err == nil {
			log.Println("[store]insert one ok")
		}
	}
	if err != nil {
		return err
	}
	_ = result
	return nil
}

func NewBufferedMongoStorage(uri, database, collection string, bufferSize int, autoFlushInterval time.Duration, opts ...BufferedMongoStorageOption) *BufferedMongoStorage {
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

	store := &BufferedMongoStorage{Cli: cli, L: &l, bufSize: bufferSize, buf: buf}
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

func (s *BufferedMongoStorage) Save(items ...parser.ParseItem) error {
	s.L.Lock()
	defer s.L.Unlock()

	if len(items) > (s.bufSize - len(s.buf)) {
		if err := s.flush(); err != nil {
			return err
		}

	} else {
		s.buf = append(s.buf, items...)
	}
	return nil
}

func (s *BufferedMongoStorage) flush() error {
	//

	if len(s.buf) == 0 {
		return nil
	}
	err := s.insertManyTOMongo(s.buf...) //  我需要拿到{taskid: 数量}
	if err != nil {
		return err
	}
	tc := collectTaskCounts(s.buf)
	s.afterFlush(tc)
	s.buf = make([]parser.ParseItem, 0, s.bufSize)
	log.Printf("Flushed")
	return nil
}

func collectTaskCounts(buf []parser.ParseItem) (tc map[string]int64) {
	tc = make(map[string]int64, 128)
	for _, item := range buf {
		if taskId, ok := item["taskId"]; !ok {
			panic(fmt.Errorf("`taskId` not fount in Parseitem, if u want to use the task counter, then taskId embeded must be stuffed in the ParsedItem"))
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

func (s *BufferedMongoStorage) afterFlush(tc map[string]int64) {
	for k, v := range tc {
		s.counter.Incr(k, v)
	}
}

func (s *BufferedMongoStorage) insertManyTOMongo(items ...parser.ParseItem) error {
	// ...  {taskid: +1}
	if result, err := s.Cli.Collection.InsertMany(context.Background(), items); err != nil {
		return err
	} else {
		log.Printf("%d objects saved", len(result.InsertedIDs))
		return nil
	}
}

func (s *BufferedMongoStorage) Close() error {
	s.flush()
	return s.Cli.Close(context.Background())
}
