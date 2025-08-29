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

var benchmarkCmd = &cobra.Command{
	Use:   "benchmark [file_hash] [peer_id...]",
	Short: "Runa a download benchmark for a specific file",
	Long:  `This command Triggers a download on the running daemon and logs the performance (time and speed) to benchmarks.txt`,
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
		fmt.Printf("Requesting benchmark for file '%s'...\n", meta.Name)

		resp, err := http.Post("http://localhost:8000/benchmark/transfer", "application/json", bytes.NewBuffer(payloadBytes))

		if err != nil {
			fmt.Println("Error: Could not connect to the go-peerfs daemon.")
			return
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		fmt.Println(string(body))

	},
}

func init() {
	rootCmd.AddCommand(benchmarkCmd)
}
