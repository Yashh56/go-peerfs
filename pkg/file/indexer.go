package file

import (
	"encoding/hex"
	"io"
	"os"
	"path/filepath"

	"github.com/minio/sha256-simd"
)

type FileMeta struct {
	Name      string
	Path      string
	Size      int64
	FileHash  string
	ChunkHash []string
}

func IndexDirectory(dir string) ([]FileMeta, error) {
	var files []FileMeta

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		meta, err := indexFile(path, info)
		if err != nil {
			return err
		}
		files = append(files, meta)
		return nil
	})

	return files, err
}

// func indexFile(path string) (FileMeta, error) {
// 	f, err := os.Open(path)

// 	if err != nil {
// 		return FileMeta{}, err
// 	}
// 	defer f.Close()

// 	h := sha256.New()
// 	var chunkHashes []string
// 	chunks, err := Chunk(f)
// 	if err != nil {
// 		return FileMeta{}, err
// 	}

// 	for _, c := range chunks {
// 		ch := sha256.Sum256(c)
// 		chunkHashes = append(chunkHashes, hex.EncodeToString(ch[:]))
// 		h.Write(c)
// 	}

// 	info, _ := f.Stat()
// 	return FileMeta{
// 		Name:      info.Name(),
// 		Path:      path,
// 		Size:      info.Size(),
// 		FileHash:  hex.EncodeToString(h.Sum(nil)),
// 		ChunkHash: chunkHashes,
// 	}, nil
// }

func indexFile(path string, info os.FileInfo) (FileMeta, error) {
	f, err := os.Open(path)
	if err != nil {
		return FileMeta{}, err
	}
	defer f.Close()

	totalFileHash := sha256.New()
	var chunkHashes []string

	buf := make([]byte, chunkSize)

	for {
		n, err := f.Read(buf)
		if err != nil && err != io.EOF {
			return FileMeta{}, err
		}
		if n == 0 {
			break
		}

		chunkData := buf[:n]

		chunkHash := sha256.Sum256(chunkData)
		chunkHashes = append(chunkHashes, hex.EncodeToString(chunkHash[:]))

		totalFileHash.Write(chunkData)
	}

	return FileMeta{
		Name:      info.Name(),
		Path:      path,
		Size:      info.Size(),
		FileHash:  hex.EncodeToString(totalFileHash.Sum(nil)),
		ChunkHash: chunkHashes,
	}, nil
}
