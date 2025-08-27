package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Yashh56/go-peerfs/pkg/file"
	"github.com/spf13/cobra"
)

var downloadCmd = &cobra.Command{
	Use:   "download [file_hash] [peer_id...]",
	Short: "Download a file from one or more peers.",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		fileHash := args[0]
		peerStrings := args[1:]

		meta, err := getFileMeta(fileHash)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		type DownloadRequest struct {
			Meta      file.FileMeta `json:"meta"`
			Providers []string      `json:"providers"`
		}
		payload := DownloadRequest{
			Meta:      meta,
			Providers: peerStrings,
		}
		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			fmt.Printf("Error creating request payload: %v\n", err)
			return
		}

		fmt.Printf("Sending download request to daemon for file '%s'...\n", meta.Name)
		resp, err := http.Post("http://localhost:8000/download", "application/json", bytes.NewBuffer(payloadBytes))
		if err != nil {
			fmt.Println("Error: Could not connect to the go-peerfs daemon.")
			fmt.Println("Please make sure the daemon is running with 'go-peerfs start'.")
			return
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusOK {
			fmt.Printf("Error from daemon: %s - %s\n", resp.Status, string(body))
		} else {
			fmt.Println(string(body))
		}
	},
}

func getFileMeta(hash string) (file.FileMeta, error) {
	var meta file.FileMeta
	fmt.Println(hash)
	resp, err := http.Get("http://localhost:8000/fileMeta?hash=" + hash)
	if err != nil {
		return meta, fmt.Errorf("could not connect to the go-peerfs daemon")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return meta, fmt.Errorf("daemon returned an error: %s", resp.Status)
	}

	if err := json.NewDecoder(resp.Body).Decode(&meta); err != nil {
		return meta, fmt.Errorf("error parsing file metadata")
	}
	return meta, nil
}

func init() {
	rootCmd.AddCommand(downloadCmd)
}
