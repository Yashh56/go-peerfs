package main

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/Yashh56/go-peerfs/pkg/download"
	"github.com/Yashh56/go-peerfs/pkg/file"
	"github.com/Yashh56/go-peerfs/pkg/p2p"
	"github.com/libp2p/go-libp2p/core/peer"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	host, err := p2p.Host(ctx)
	if err != nil {
		log.Fatalf("Failed to create host: %v", err)
	}
	defer host.Close()

	sharedFiles, err := file.IndexDirectory("./shared")
	if err != nil {
		log.Fatalf("Failed to index directory: %v", err)
	}
	fmt.Printf("Sharing %d files.\n", len(sharedFiles))

	p2p.SetStreamHandler(host, sharedFiles)
	p2p.SetSearchHandler(host, sharedFiles)

	go p2p.DiscoveryService(ctx, host)

	go func() {
		time.Sleep(20 * time.Second)

		peers := host.Peerstore().Peers()
		if len(peers) <= 1 {
			fmt.Println("No peers found to search.")
			return
		}

		var targetPeer peer.ID
		for _, p := range peers {
			if p != host.ID() {
				targetPeer = p
				break
			}
		}

		if targetPeer == "" {
			fmt.Println("Could not select a target peer.")
			return
		}

		searchQuery := "large-test-file"
		fmt.Printf("\n--- Searching for '%s' on peer %s ---\n", searchQuery, targetPeer)

		results, err := p2p.RequestSearch(ctx, host, targetPeer, searchQuery)
		if err != nil {
			log.Printf("Search request failed: %v", err)
			return
		}

		if len(results) == 0 {
			fmt.Println("Search returned no results.")
			return
		}

		fileToDownload := results[0]

		var providers []peer.ID

		for _, p := range host.Peerstore().Peers() {
			if p != host.ID() {
				providers = append(providers, p)
			}
		}

		if len(providers) == 0 {
			fmt.Println("No Providers Found for The File.")
			return
		}

		fmt.Printf("Found %d Providers for file %s. Starting multi-peer download.\n", len(providers), fileToDownload.Name)

		savePath := filepath.Join("./downloads", fileToDownload.Name)

		dlManager := download.NewDownloadManager(host)
		err = dlManager.DownloadFile(ctx, fileToDownload, providers, savePath)

		if err != nil {
			log.Printf("Multi-peer download failed: %v", err)
		} else {
			fmt.Printf("Multi-Peer Download Succeeded!\n")
		}
	}()

	select {}
}
