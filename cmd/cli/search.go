package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

type SearchResult struct {
	Name     string `json:"name"`
	Size     int64  `json:"size"`
	FileHash string `json:"file_hash"`
	PeerID   string `json:"peer_id"`
}

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search for files on the network",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := args[0]
		fmt.Println(query)
		resp, err := http.Get("http://localhost:8000/search?q=" + query)
		if err != nil {
			fmt.Println("Error: Could not connect to the go-peerfs daemon.")
			fmt.Println("Please make sure the daemon is running with 'go-peerfs start'.")
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Error Reading response Body: %v\n", err)
			return
		}
		var results []SearchResult

		if err := json.Unmarshal(body, &results); err != nil {
			fmt.Printf("Error parsing Search results :%v\n", err)
			return
		}
		if len(results) == 0 {
			fmt.Println("No Results found for your query.")
			return
		}

		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 0, 8, 2, '\t', 0)
		fmt.Fprintln(w, "NAME\tSIZE (Bytes)\tHASH\tPEER ID")
		fmt.Fprintln(w, "----\t------------\t----\t-------")
		for _, res := range results {
			fmt.Fprintf(w, "%s\t%d\t%s\t%s\n", res.Name, res.Size, res.FileHash, res.PeerID)
		}
		w.Flush()
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
