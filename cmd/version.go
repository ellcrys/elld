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
		fmt.Println(fmt.Sprintf(`Client Version:   %s
Protocol Version: %s
Go Version:       %s
Commit Hash: 	  %s
Build Date: 	  %s`,
			BuildVersion, config.GetVersions().Protocol, GoVersion, BuildCommit, BuildDate))
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
