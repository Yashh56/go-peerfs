package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	"github.com/Yashh56/go-peerfs/pkg/download"
	"github.com/Yashh56/go-peerfs/pkg/file"
	"github.com/Yashh56/go-peerfs/pkg/p2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/spf13/cobra"
)

var apiPort int

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the go-peerfs node and connect to the network.",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		p2pHost, err := p2p.Host(ctx)
		if err != nil {
			log.Fatalf("Failed to create host: %v", err)
		}
		sharedFiles, err := file.IndexDirectory("./shared")
		if err != nil {
			log.Fatalf("Failed to index Directory: %v", err)
		}
		fmt.Printf("Sharing %d files.\n", len(sharedFiles))

		p2p.SetStreamHandler(p2pHost, sharedFiles)
		p2p.SetSearchHandler(p2pHost, sharedFiles)

		go p2p.DiscoveryService(ctx, p2pHost)
		fmt.Printf("NODE ID: %s\n", p2pHost.ID())

		go startAPIServer(p2pHost, sharedFiles)

		fmt.Println("Node is Running. Press Ctrl+C to Exit.")
		select {}
	},
}

func startAPIServer(h host.Host, localFiles []file.FileMeta) {

	handleSearch := func(w http.ResponseWriter, r *http.Request) {
		queryValues := r.URL.Query()
		query := queryValues.Get("q")
		if query == "" {
			http.Error(w, "Missing search query 'q'", http.StatusBadRequest)
			return
		}
		fmt.Printf("API: Received search query '%s'\n", query)

		type SearchResult struct {
			Name     string `json:"name"`
			Size     int64  `json:"size"`
			FileHash string `json:"file_hash"`
			PeerID   string `json:"peer_id"`
		}
		var allResults []SearchResult

		localResults := file.SearchLocal(localFiles, query)
		for _, meta := range localResults {
			allResults = append(allResults, SearchResult{
				Name:     meta.Name,
				Size:     meta.Size,
				FileHash: meta.FileHash,
				PeerID:   h.ID().String(), // Add our own ID
			})
		}
		peers := h.Peerstore().Peers()
		for _, p := range peers {
			if p == h.ID() {
				continue
			}
			results, err := p2p.RequestSearch(r.Context(), h, p, query)

			if err != nil {
				fmt.Printf("Error Searching Peer %s: %v\n", p, err)
				continue
			}
			for _, meta := range results {
				allResults = append(allResults, SearchResult{
					Name:     meta.Name,
					Size:     meta.Size,
					FileHash: meta.FileHash,
					PeerID:   p.String(),
				})
			}
		}
		w.Header().Set("Content-type", "application/json")
		json.NewEncoder(w).Encode(allResults)
	}

	handleFileMeta := func(w http.ResponseWriter, r *http.Request) {
		hash := r.URL.Query().Get("hash")
		if hash == "" {
			http.Error(w, "Missing file hash", http.StatusBadRequest)
			return
		}

		for _, meta := range localFiles {
			if meta.FileHash == hash {
				w.Header().Set("Content-type", "application/json")
				json.NewEncoder(w).Encode(meta)
				return
			}
		}
		http.Error(w, "File metadata not found", http.StatusNotFound)
	}

	handleDownload := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
			return
		}

		type DownloadRequest struct {
			Meta      file.FileMeta `json:"meta"`
			Providers []string      `json:"providers"`
		}

		var req DownloadRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid Request body", http.StatusBadRequest)
			return
		}

		var providerIDs []peer.ID
		for _, pStr := range req.Providers {
			id, err := peer.Decode(pStr)
			if err != nil {
				http.Error(w, "Invalid peer ID", http.StatusBadRequest)
				return
			}
			providerIDs = append(providerIDs, id)
		}

		savePath := filepath.Join("./downloads", req.Meta.Name)
		fmt.Printf("API: Received download request for '%s'\n", req.Meta.Name)

		dlManager := download.NewDownloadManager(h, localFiles)
		err := dlManager.DownloadFile(r.Context(), req.Meta, providerIDs, savePath)
		if err != nil {
			msg := fmt.Sprintf("Download failed: %v", err)
			http.Error(w, msg, http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Download successful! File saved to %s", savePath)
	}

	http.HandleFunc("/search", handleSearch)
	http.HandleFunc("/fileMeta", handleFileMeta)
	http.HandleFunc("/download", handleDownload)

	listenAddr := fmt.Sprintf(":%d", apiPort)
	fmt.Printf("API Server listening on http://localhost%s\n", listenAddr)

	if err := http.ListenAndServe(listenAddr, nil); err != nil {
		log.Fatalf("Failed to start API server: %v", err)
	}
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().IntVarP(&apiPort, "port", "p", 8000, "Port for the API server")

}
