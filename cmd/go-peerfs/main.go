package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Yashh56/go-peerfs/pkg/p2p"
)

func main() {
	ctx := context.Background()

	host, err := p2p.Host(ctx)
	if err != nil {
		log.Fatalf("Failed to Create an Host: %v", err)
	}
	defer host.Close()

	fmt.Println("Host Created Successfully. Starting Discovery...")

	if err := p2p.DiscoveryService(ctx, host); err != nil {
		log.Fatalf("Discovery failed: %v", err)
	}

	select {}
}
