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
	"fmt"

	"github.com/ellcrys/elld/console"
	"github.com/ellcrys/elld/crypto"
	"github.com/spf13/cobra"
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

		rpcAddress, _ := cmd.Flags().GetString("rpcaddress")
		account, _ := cmd.Flags().GetString("account")
		password, _ := cmd.Flags().GetString("pwd")

		var err error
		var loadedAddress *crypto.Key
		if account != "" {
			loadedAddress, err = loadOrCreateAccount(account, password, 0)
			if err != nil {
				log.Fatal(err.Error())
			}
		}

		cs := console.New(loadedAddress, consoleHistoryFilePath)
		err = cs.DialRPCServer(rpcAddress)
		if err != nil {
			log.Fatal("unable to start RPC server", "Err", err)
		}

		cs.PrepareVM()
		if err != nil {
			log.Fatal("unable to prepare console VM", "Err", err)
		}

		fmt.Println("")
		cs.Run()
	},
}

func init() {
	rootCmd.AddCommand(attachCmd)
	attachCmd.Flags().String("rpcaddress", "127.0.0.1:8999", "Address RPC server will listen on")
	attachCmd.Flags().String("account", "", "Account to load. Default account is used if not provided")
	attachCmd.Flags().String("pwd", "", "Used as password during initial account creation or to unlock an account")
}
