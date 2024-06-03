package ua

import (
	"context"
	"math/rand"
)

type UaGetter interface {
	Get(context.Context) (string, error)
}

type roundRobinUA struct{}

func (u *roundRobinUA) Get(ctx context.Context) (string, error) {
	return DEFAULT_UAS[rand.Intn(len(DEFAULT_UAS)-1)], nil
}

func NewRoundRobinUAGetter() *roundRobinUA {
	return &roundRobinUA{}
}
