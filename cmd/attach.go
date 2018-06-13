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
	"github.com/spf13/cobra"
)

// attachCmd represents the attach command
var attachCmd = &cobra.Command{
	Use:   "attach",
	Short: "Attach to a remote node",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {

		rpcAddress, _ := cmd.Flags().GetString("rpcaddress")

		cs := console.New(nil)
		err := cs.DialRPCServer(rpcAddress)
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
}
