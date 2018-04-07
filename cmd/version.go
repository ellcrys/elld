package cmd

import (
	"fmt"

	"github.com/ellcrys/gcoin/peer"
	"github.com/spf13/cobra"
)

// ClientVersion is the version of this current code base
var ClientVersion = "0.0.1"

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Client, protocol and Go versions",
	Long:  `Client, protocol and Go versions`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(fmt.Sprintf("Client Version: v%s", ClientVersion))
		fmt.Println("Protocol Version: ", peer.ProtocolVersion)
		fmt.Println("Go Version: ", "go1.10")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
