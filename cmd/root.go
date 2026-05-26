package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "meshsync",
	Short: "Decentralized local-first sync cluster",
	Long:  "MeshSync is a decentralized peer-to-peer file synchronization system.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
