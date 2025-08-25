package download

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/Yashh56/go-peerfs/pkg/file"
	"github.com/Yashh56/go-peerfs/pkg/p2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
)

type DownloadManager struct {
	Host host.Host
}

func NewDownloadManager(h host.Host) *DownloadManager {
	return &DownloadManager{Host: h}
}

func (dm *DownloadManager) DownloadFile(ctx context.Context, meta file.FileMeta, providers []peer.ID, savePath string) error {
	numChunks := len(meta.ChunkHash)
	downloadedChunks := make([][]byte, numChunks)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i := 0; i < numChunks; i++ {
		wg.Add(1)
		provider := providers[i%len(providers)]
		chunkIndex := i

		go func(p peer.ID, index int) {
			defer wg.Done()
			fmt.Printf("Requesting Chunk %d from Peer %s\n", index, p)

			chunkData, err := p2p.RequestChunk(ctx, dm.Host, p, meta.FileHash, index)

			if err != nil {
				fmt.Printf("Error Downloading Chunk %d from Peer %s: %v\n", index, p, err)
				return
			}
			mu.Lock()
			downloadedChunks[index] = chunkData
			mu.Unlock()
			fmt.Printf("Successfully Downloaded Chunk %d\n", index)
		}(provider, chunkIndex)
	}
	wg.Wait()
	fmt.Println("All Chunks Downloaded. Reassembling File...")

	f, err := os.Create(savePath)
	if err != nil {
		return err
	}
	defer f.Close()

	for i := 0; i < numChunks; i++ {
		if downloadedChunks[i] == nil {
			return fmt.Errorf("Downloaded Failed: Missing Chunks %d", i)
		}
		_, err = f.Write(downloadedChunks[i])
		if err != nil {
			return err
		}
	}
	fmt.Println("File Reassembled Successfully!")
	return nil
}
