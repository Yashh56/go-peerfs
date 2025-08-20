package p2p

import (
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
)

func Host(ctx context.Context) (host.Host, error) {
	host, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"),
	)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Host Created with Id: %s\n", host.ID())

	fmt.Println("Listen Addresses: ", host.Addrs())

	return host, nil
}
