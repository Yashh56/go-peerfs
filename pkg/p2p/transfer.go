package p2p

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Yashh56/go-peerfs/pkg/file"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
)

const FileTransferProtocol = "/go-peerfs/transfer/1.0.0"

var fileIndex []file.FileMeta

func SetStreamHandler(h host.Host, files []file.FileMeta) {
	fileIndex = files
	h.SetStreamHandler(FileTransferProtocol, fileStreamHandler)
	fmt.Println("File Transfer stream handler set.")
}

func fileStreamHandler(s network.Stream) {
	fmt.Printf("New incoming stream from %s\n", s.Conn().RemotePeer())
	defer s.Close()

	reader := bufio.NewReader(s)

	requestString, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Error reading from stream: %v\n", err)
		return
	}

	requestString = strings.TrimSpace(requestString)
	parts := strings.Split(requestString, ":")
	if len(parts) != 2 {
		fmt.Println("Invalid request format. Expected 'filehash:chunkIndex'")
		return
	}

	fileHash, chunkIndexStr := parts[0], parts[1]

	chunkIndex, err := strconv.Atoi(chunkIndexStr)

	if err != nil {
		fmt.Printf("Invalid chunk index: %v\n", err)
		return
	}
	fmt.Printf("Peer %s is requesting chunk %d for file %s\n", s.Conn().RemotePeer(), chunkIndex, fileHash)

	var requestedFile *file.FileMeta

	for _, f := range fileIndex {
		if f.FileHash == fileHash {
			requestedFile = &f
			break
		}
	}

	if requestedFile == nil {
		fmt.Printf("File with hash %s not found.\n", fileHash)
		return
	}
	f, err := os.Open(requestedFile.Path)
	if err != nil {
		fmt.Printf("Error opening file %s: %v\n", requestedFile.Path, err)
		return
	}

	defer f.Close()

	offset := int64(chunkIndex) * file.ChunkSize
	f.Seek(offset, 0)

	chunkReader := io.LimitedReader{
		R: f,
		N: file.ChunkSize,
	}
	bytesSent, err := io.Copy(s, &chunkReader)

	if err != nil {
		fmt.Printf("Error Sending Chunk: %v\n", err)
		return
	}

	fmt.Printf("Finished sending chunk %d. Sent %d bytes.\n", chunkIndex, bytesSent)

}

func RequestFile(ctx context.Context, h host.Host, peerID peer.ID, meta file.FileMeta, savePath string) error {
	fmt.Printf("Opening Stream to %s for file %s\n", peerID, meta.FileHash)

	s, err := h.NewStream(ctx, peerID, FileTransferProtocol)
	if err != nil {
		return fmt.Errorf("Failed to Open Stream: %w", err)
	}
	defer s.Close()

	writer := bufio.NewWriter(s)

	_, err = writer.WriteString(meta.FileHash + "\n")
	if err != nil {
		return fmt.Errorf("Failed to write request to stream: %w", err)
	}
	writer.Flush()

	f, err := os.Create(savePath)
	if err != nil {
		return fmt.Errorf("Failed to create save file: %w", err)
	}
	defer f.Close()

	fmt.Printf("Downloading File to %s...\n", savePath)

	bytesReceived, err := io.Copy(f, s)
	if err != nil {
		return fmt.Errorf("Error During File Download: %w", err)
	}

	fmt.Printf("File Download Complete! Received %d bytes. \n", bytesReceived)

	return nil
}

func RequestChunk(ctx context.Context, h host.Host, peerID peer.ID, fileHash string, chunkIndex int) ([]byte, error) {
	request := fmt.Sprintf("%s:%d\n", fileHash, chunkIndex)
	streamCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	s, err := h.NewStream(streamCtx, peerID, FileTransferProtocol)
	if err != nil {
		return nil, err
	}

	defer s.Close()

	_, err = s.Write([]byte(request))

	if err != nil {
		return nil, err
	}

	s.CloseWrite()
	chunkData, err := io.ReadAll(s)
	if err != nil {
		return nil, err
	}
	return chunkData, nil

}
