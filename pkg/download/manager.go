package download

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"

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

	f, err := os.Create(savePath)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Printf("Starting sequential download of %d chunks...\n", numChunks)

	for i := 0; i < numChunks; i++ {
		provider := providers[i%len(providers)]
		chunkIndex := i
		var chunkData []byte
		var err error

		if provider == dm.Host.ID() {
			fmt.Printf("Reading chunk %d from local disk...\n", chunkIndex)
			chunkData, err = dm.readLocalChunk(meta.FileHash, chunkIndex)
		} else {
			fmt.Printf("Requesting chunk %d from remote peer %s...\n", chunkIndex, provider)
			chunkData, err = p2p.RequestChunk(ctx, dm.Host, provider, meta.FileHash, chunkIndex)
		}

		if err != nil {
			return fmt.Errorf("failed to get chunk %d: %w", chunkIndex, err)
		}

		expectedHash := meta.ChunkHash[chunkIndex]
		hasher := sha256.New()
		hasher.Write(chunkData)
		receivedHash := hex.EncodeToString(hasher.Sum(nil))

		if receivedHash != expectedHash {
			return fmt.Errorf("chunk %d verification failed! Corrupted data", chunkIndex)
		}

		if _, err := f.Write(chunkData); err != nil {
			return fmt.Errorf("failed to write chunk %d to file: %w", chunkIndex, err)
		}
		fmt.Printf("Successfully downloaded and wrote chunk %d\n", chunkIndex)
	}

	fmt.Println("File download complete!")
	return nil
}

func (dm *DownloadManager) readLocalChunk(fileHash string, chunkIndex int) ([]byte, error) {
	var localMeta *file.FileMeta
	for _, f := range dm.LocalFiles {
		if f.FileHash == fileHash {
			localMeta = &f
			break
		}
	}
	if localMeta == nil {
		return nil, fmt.Errorf("could not find local file metadata for hash %s", fileHash)
	}

	f, err := os.Open(localMeta.Path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	offset := int64(chunkIndex) * file.ChunkSize
	if _, err := f.Seek(offset, 0); err != nil {
		return nil, err
	}

	buf := make([]byte, file.ChunkSize)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return nil, err
	}

	return buf[:n], nil
}
