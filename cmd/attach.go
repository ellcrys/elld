// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"os"

	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/console"
	"github.com/ellcrys/elld/crypto"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// attachCmd represents the attach command
var attachCmd = &cobra.Command{
	Use:   "attach",
	Short: "Attach to a remote node and start a Javascript console.",
	Long: `Description:
Attach to a remote node and start a Javascript console.
	
To load an initial account, provide the account using '--account' flag. 
An interactive console will be started to unlock the account. To skip the
interactive session, set the unlock password using '--pwd' flag. The provided
account does not exist, the command will fail.`,
	Run: func(cmd *cobra.Command, args []string) {

		viper.BindPFlag("node.account", cmd.Flags().Lookup("account"))
		viper.BindPFlag("node.password", cmd.Flags().Lookup("pwd"))
		viper.BindPFlag("rpc.address", cmd.Flags().Lookup("rpc-address"))
		account := viper.GetString("node.account")
		password := viper.GetString("node.password")
		rpcAddress := viper.GetString("rpc.address")

		var err error
		var coinbase *crypto.Key

		// load the coinbase coinbase account.
		if account != "" {
			coinbase, err = getKey(account, password, 0)
			if err != nil {
				log.Fatal(err.Error())
			}
		}

		// Set up the console in attach mode
		cs := console.NewAttached(coinbase, consoleHistoryFilePath, cfg, log)
		cs.SetVersions(config.GetVersions().Protocol, BuildVersion, GoVersion, BuildCommit)

		// Set the RPC server address to be dialled
		cs.SetRPCServerAddr(rpcAddress, false)

		// Prepare the console and JS context
		if err := cs.Prepare(); err != nil {
			log.Fatal("failed to prepare console VM", "Err", err)
		}

		cs.OnStop(func() {
			os.Exit(0)
		})

		cs.Run()
	},
}

func init() {
	rootCmd.AddCommand(attachCmd)
	attachCmd.Flags().String("rpc-address", "127.0.0.1:8999", "Address RPC server will listen on")
	attachCmd.Flags().String("account", "", "Account to load. Default account is used if not provided")
	attachCmd.Flags().String("pwd", "", "Used as password during initial account creation or to unlock an account")
}
