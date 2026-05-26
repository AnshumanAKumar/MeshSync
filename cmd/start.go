package cmd

import (
	"fmt"
	"meshsync/internal/runtime"

	"github.com/spf13/cobra"
)

var (
	bootstrap bool
	join      bool
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start MeshSync node",
	RunE: func(cmd *cobra.Command, args []string) error {

		if bootstrap {

			rt := runtime.New(runtime.BootstrapRole)

			return rt.Start()
		}

		if join {

			rt := runtime.New(runtime.PeerRole)

			return rt.Start()
		}

		return fmt.Errorf("please specify either --bootstrap or --join")
	},
}

func init() {
	startCmd.Flags().BoolVar(
		&bootstrap,
		"bootstrap",
		false,
		"Start node in bootstrap mode",
	)

	startCmd.Flags().BoolVar(
		&join,
		"join",
		false,
		"Start node in peer mode",
	)

	rootCmd.AddCommand(startCmd)
}

func startBootstrapNode() {
	fmt.Println("Starting bootstrap node...")
}

func startPeerNode() {
	fmt.Println("Starting peer node...")
}
