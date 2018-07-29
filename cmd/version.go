package cmd

import (
	"fmt"

	"github.com/ellcrys/elld/config"
	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Client, protocol and Go versions",
	Long:  `Client, protocol and Go versions`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(fmt.Sprintf("Client Version: v%s", config.ClientVersion))
		fmt.Println("Protocol Version: ", config.ProtocolVersion)
		fmt.Println("Go Version: ", "go1.10")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
