package p2p

import (
	"context"
	"fmt"
	"sync"
	"time"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	dutil "github.com/libp2p/go-libp2p/p2p/discovery/util"
)

const rendezvousString = "go-peerfs-rendezvous"

type notifee struct {
	h host.Host
}

func (n *notifee) HandlePeerFound(pi peer.AddrInfo) {
	if pi.ID == n.h.ID() {
		return
	}
	fmt.Printf("Found Peer via mDNS: %s\n", pi.ID.String())
	if err := n.h.Connect(context.Background(), pi); err != nil {
		fmt.Printf("Failed to connect to mDNS peer %s: %s\n", pi.ID.String())
	} else {
		fmt.Printf("Connected to mDNS peer: %s\n", pi.ID.String())
	}
}

func DiscoveryService(ctx context.Context, h host.Host) error {
	fmt.Println("Starting mDNS for local discovery...")
	mdnsService := mdns.NewMdnsService(h, rendezvousString, &notifee{h: h})
	if err := mdnsService.Start(); err != nil {
		return fmt.Errorf("failed to start mDNS: %w", err)
	}
	fmt.Println("mDNS started successfully.")

	kadDHT, err := dht.New(ctx, h, dht.Mode(dht.ModeServer))
	if err != nil {
		return fmt.Errorf("failed to create DHT: %w", err)
	}

	fmt.Println("Bootstrapping the DHT...")
	if err = kadDHT.Bootstrap(ctx); err != nil {
		return fmt.Errorf("failed to bootstrap DHT: %w", err)
	}

	routingDiscovery := drouting.NewRoutingDiscovery(kadDHT)

	fmt.Println("Announcing our presence via DHT...")
	dutil.Advertise(ctx, routingDiscovery, rendezvousString)
	fmt.Println("Successfully announced!")

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			fmt.Println("Searching for other peers via DHT...")
			peerChan, err := routingDiscovery.FindPeers(ctx, rendezvousString)
			if err != nil {
				fmt.Printf("Failed to find DHT peers: %v\n", err)
				continue
			}

			var wg sync.WaitGroup
			for p := range peerChan {
				if p.ID == h.ID() || h.Network().Connectedness(p.ID) == network.Connected {
					continue
				}
				wg.Add(1)
				go func(peerInfo peer.AddrInfo) {
					defer wg.Done()
					fmt.Printf("Found DHT peer: %s\n", peerInfo.ID.String())
					if err := h.Connect(ctx, peerInfo); err != nil {
						fmt.Printf("Failed to connect to DHT peer %s: %s\n", peerInfo.ID.String(), err)
					} else {
						fmt.Printf("Connected to DHT peer: %s\n", peerInfo.ID.String())
					}
				}(p)
			}
			wg.Wait()
			fmt.Println("DHT peer discovery round finished.")
		}
	}
}
