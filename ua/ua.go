package ua

import "math/rand"

type UaGetter interface {
	Get() (string, error)
}

type roundRobinUA struct{}

func (u *roundRobinUA) Get() (string, error) {
	return DEFAULT_UAS[rand.Intn(len(DEFAULT_UAS)-1)], nil
}

func NewDefaultUAGetter() *roundRobinUA {
	return &roundRobinUA{}
}
