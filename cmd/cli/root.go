package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "go-peerfs",
	Short: "A P2P File sharing System in Go.",
	Long:  `go-peerfs is a decentralized file sharing application built with libp2p`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
