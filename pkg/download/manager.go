package download

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/Yashh56/go-peerfs/pkg/file"
	"github.com/Yashh56/go-peerfs/pkg/p2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
)

type DownloadManager struct {
	Host       host.Host
	LocalFiles []file.FileMeta
}

func NewDownloadManager(h host.Host, localFiles []file.FileMeta) *DownloadManager {
	return &DownloadManager{
		Host:       h,
		LocalFiles: localFiles,
	}
}

func (dm *DownloadManager) DownloadFile(ctx context.Context, meta file.FileMeta, providers []peer.ID, savePath string) error {
	numChunks := len(meta.ChunkHash)
	if numChunks == 0 {
		return fmt.Errorf("metadata contains no chunk hashes, cannot download")
	}

	downloadedChunks := make([][]byte, numChunks)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i := 0; i < numChunks; i++ {
		wg.Add(1)
		provider := providers[i%len(providers)]
		chunkIndex := i

		go func(p peer.ID, index int) {
			defer wg.Done()
			var chunkData []byte
			var err error

			if p == dm.Host.ID() {
				fmt.Printf("Reading chunk %d locally\n", index)
				var localMeta *file.FileMeta
				for _, f := range dm.LocalFiles {
					if f.FileHash == meta.FileHash {
						localMeta = &f
						break
					}
				}
				if localMeta == nil {
					err = fmt.Errorf("Could Not Find Local File Metadata for hash %s", meta.FileHash)
				} else {

					chunkData, err = readChunkFromFile(localMeta.Path, index)
				}
			} else {
				fmt.Printf("Requesting chunk %d from remote peer %s\n", index, p)
				chunkData, err = p2p.RequestChunk(ctx, dm.Host, p, meta.FileHash, index)
			}

			if err != nil {
				fmt.Printf("Error getting chunk %d from peer %s: %v\n", index, p, err)
				return
			}

			expectedHash := meta.ChunkHash[index]
			hasher := sha256.New()
			hasher.Write(chunkData)
			receivedHash := hex.EncodeToString(hasher.Sum(nil))

			if receivedHash != expectedHash {
				fmt.Printf("Chunk %d verification failed!\n", index)
				return
			}

			mu.Lock()
			downloadedChunks[index] = chunkData
			mu.Unlock()
			fmt.Printf("Successfully got and verified chunk %d\n", index)
		}(provider, chunkIndex)
	}

	wg.Wait()
	fmt.Println("All chunk operations complete. Reassembling file...")

	f, err := os.Create(savePath)
	if err != nil {
		return err
	}
	defer f.Close()

	for i := 0; i < numChunks; i++ {
		if downloadedChunks[i] == nil {
			return fmt.Errorf("download failed: missing chunk %d", i)
		}
		_, err := f.Write(downloadedChunks[i])
		if err != nil {
			return err
		}
	}
	fmt.Println("File reassembled successfully!")
	return nil
}

func readChunkFromFile(filePath string, chunkIndex int) ([]byte, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	offSet := int64(chunkIndex) * file.ChunkSize
	_, err = f.Seek(offSet, 0)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, file.ChunkSize)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return nil, err
	}
	return buf[:n], err
}
