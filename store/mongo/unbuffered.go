package mongo

import (
	"context"
	"log"

	"github.com/qiniu/qmgo"
	"github.com/superjcd/gocrawler/parser"
	"github.com/superjcd/gocrawler/store"
)

type mongoStorage struct {
	Cli *qmgo.QmgoClient
}

var _ store.Storage = (*mongoStorage)(nil)

func NewMongoStorage(uri, database, collection string) *mongoStorage {
	ctx := context.Background()
	cli, err := qmgo.Open(ctx, &qmgo.Config{Uri: uri,
		Database: database,
		Coll:     collection}) // counter
	if err != nil {
		panic(err)
	}

	return &mongoStorage{Cli: cli}
}

func (s *mongoStorage) Save(ctx context.Context, items ...parser.ParseItem) error {
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
