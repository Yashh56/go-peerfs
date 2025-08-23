package p2p

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"

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
	fmt.Printf("New Incoming Stream from %s\n", s.Conn().RemotePeer())
	defer s.Close()

	reader := bufio.NewReader(s)

	fileHash, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Error reading from stream: %v\n", err)
		return
	}

	fileHash = fileHash[:len(fileHash)-1]

	fmt.Printf("Peer %s is requesting file with hash: %s\n", s.Conn().RemotePeer(), fileHash)

	var requestedFile *file.FileMeta
	for _, f := range fileIndex {
		if f.FileHash == fileHash {
			requestedFile = &f
			break
		}
	}
	if requestedFile == nil {
		fmt.Printf("File with hash %s not found. \n", fileHash)
		return
	}

	f, err := os.Open(requestedFile.Path)
	if err != nil {
		fmt.Printf("Error opening File %s: %v\n", requestedFile.Path, err)
		return
	}
	defer f.Close()

	fmt.Printf("Sending File %s to Peer %s\n", requestedFile.Name, s.Conn().RemotePeer())

	bytesSent, err := io.Copy(s, f)
	if err != nil {
		fmt.Printf("Error Sending File :%v,\n", err)
		return
	}
	fmt.Printf("Finished Sending File. Sent %d bytes.\n", bytesSent)
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
