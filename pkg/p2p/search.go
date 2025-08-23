package p2p

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/Yashh56/go-peerfs/pkg/file"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
)

const SearchProtocol = "/go-peerfs/search/1.0.0"

func SetSearchHandler(h host.Host, files []file.FileMeta) {
	fileIndex = files
	h.SetStreamHandler(SearchProtocol, searchStreamHandler)
	fmt.Println("Search Stream Handler set.")
}

func searchStreamHandler(s network.Stream) {
	defer s.Close()

	reader := bufio.NewReader(s)
	query, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading an search query : ", err)
		return
	}
	query = query[:len(query)-1]
	fmt.Printf("Received Search Query '%s' from %s\n", err)

	results := file.SearchLocal(fileIndex, query)

	encoder := json.NewEncoder(s)

	if err := encoder.Encode(results); err != nil {
		fmt.Printf("Error Encoding Search Results: %v\n", err)
		return
	}
	fmt.Printf("sent %d search results to %s\n", len(results), s.Conn().RemotePeer())
}

func RequestSearch(ctx context.Context, h host.Host, peerID peer.ID, query string) ([]file.FileMeta, error) {
	fmt.Printf("Opening Search Stream to %s for query '%s'\n", peerID, query)
	s, err := h.NewStream(ctx, peerID, SearchProtocol)
	if err != nil {
		return nil, err
	}
	defer s.Close()

	writer := bufio.NewWriter(s)
	_, err = writer.WriteString(query + "\n")
	if err != nil {
		return nil, err
	}
	writer.Flush()

	bytes, err := io.ReadAll(s)
	if err != nil {
		return nil, err
	}
	var results []file.FileMeta

	if err := json.Unmarshal(bytes, &results); err != nil {
		return nil, fmt.Errorf("Failed to unmarshal search results: %w", err)
	}
	fmt.Printf("Received %d results from %s\n", len(results), peerID)
	return results, nil

}
