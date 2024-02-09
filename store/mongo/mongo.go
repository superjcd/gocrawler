package mongo

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/qiniu/qmgo"
	"github.com/superjcd/gocrawler/parser"
	"github.com/superjcd/gocrawler/store"
)

type BufferedMongoStorage struct {
	L       *sync.Mutex
	Cli     *qmgo.QmgoClient
	buf     []parser.ParseItem
	bufSize int
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
		Coll:     collection})
	if err != nil {
		panic(err)
	}

	return &MongoStorage{Cli: cli}
}

// TODO: index 不要忘记
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

func NewBufferedMongoStorage(uri, database, collection string, bufferSize int, autoFlushInterval time.Duration) *BufferedMongoStorage {
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

	ticker := time.NewTicker(autoFlushInterval)

	go func() {
		for t := range ticker.C {
			log.Printf("auto flush triggered at %v", t)
			store.Flush()
		}

	}()

	return store

}

func (s *BufferedMongoStorage) Save(items ...parser.ParseItem) error {
	s.L.Lock()
	defer s.L.Unlock()

	if len(items) > (s.bufSize - len(s.buf)) {
		if err := s.Flush(); err != nil {
			return err
		}

	} else {
		s.buf = append(s.buf, items...)
	}
	return nil
}

func (s *BufferedMongoStorage) Flush() error {
	if len(s.buf) == 0 {
		return nil
	}
	err := s.InsertManyTOMongo(s.buf...)
	if err != nil {
		return err
	}
	// re-allocate a new buffer
	s.buf = make([]parser.ParseItem, 0, s.bufSize)
	log.Printf("Flushed")
	return nil
}

func (s *BufferedMongoStorage) InsertManyTOMongo(items ...parser.ParseItem) error {
	if result, err := s.Cli.Collection.InsertMany(context.Background(), items); err != nil {
		return err
	} else {
		log.Printf("%d objects saved", len(result.InsertedIDs))
		return nil
	}
}

func (s *BufferedMongoStorage) Close() error {
	s.Flush()
	return s.Cli.Close(context.Background())
}
