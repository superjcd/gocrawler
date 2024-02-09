package worker_service

import (
	"context"
	"errors"
	"fmt"
	"net"
)

type WorkerService struct{}

type GreetArgs struct {
}

type GreetReply struct {
	Data string
}

func (w *WorkerService) Greet(ctx context.Context, args *GreetArgs, reply *GreetReply) error {
	if ipAddress, err := getLocalIP(); err != nil {
		return err
	} else {
		reply.Data = fmt.Sprintf("greet from %s", ipAddress)
		return nil
	}

}

func getLocalIP() (string, error) {
	var (
		addrs []net.Addr
		err   error
	)
	if addrs, err = net.InterfaceAddrs(); err != nil {
		return "", err
	}
	for _, addr := range addrs {
		if ipNet, isIpNet := addr.(*net.IPNet); isIpNet && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String(), nil
			}
		}
	}

	return "", errors.New("no local ip")
}
