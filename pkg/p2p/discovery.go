package p2p

import (
	"context"
	"fmt"
	"sync"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
)

func DiscoveryService(ctx context.Context, h host.Host) error {
	kadDHT, err := dht.New(ctx, h)
	if err != nil {
		return fmt.Errorf("failed to create DHT: %w", err)
	}
	fmt.Println("BootStrapping the DHT...")
	if err = kadDHT.Bootstrap(ctx); err != nil {
		return fmt.Errorf("failed to bootstrap DHT: %w", err)
	}

	rendezvousString := "go-peerfs-rendezvous"

	routingDiscovery := drouting.NewRoutingDiscovery(kadDHT)

	fmt.Println("Successfully Announced!")

	fmt.Println("Searching for others Peers...")

	peerChan, err := routingDiscovery.FindPeers(ctx, rendezvousString)
	if err != nil {
		return fmt.Errorf("failed to find peers: %w", err)
	}
	var wg sync.WaitGroup

	for p := range peerChan {
		if p.ID == h.ID() {
			continue
		}
		wg.Add(1)
		go func(peerInfo peer.AddrInfo) {
			defer wg.Done()
			fmt.Printf("Found Peer: %s\n", peerInfo.ID.String())
			if err := h.Connect(ctx, peerInfo); err != nil {
				fmt.Printf("Failed to connect to peer %s: %s\n", peerInfo.ID.String(), err)
			} else {
				fmt.Printf("Connected to Peer: %s\n", peerInfo.ID.String())
			}
		}(p)
	}
	wg.Wait()
	fmt.Println("Peer Discovery Finished.")

	return nil
}
