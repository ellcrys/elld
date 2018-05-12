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

	"github.com/fatih/color"

	"github.com/ellcrys/druid/crypto"
	"github.com/spf13/cobra"
)

// addressCmd represents the address command
var addressCmd = &cobra.Command{
	Use:   "address",
	Short: "Create an address",
	Long: `Description:
  Create an address. The address, peer ID, public and private keys are
  displayed. 
  
  Use '--seed' set your own Int64 random number to be used as the seed.
  It '--seed' is not provided or set to -1, a random seed is used.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// addressNewCmd represents the address command
var addressNewCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new address",
	Long:  `Create a new address`,
	Run: func(cmd *cobra.Command, args []string) {

		var seed *int64 = nil
		_seed, _ := cmd.Flags().GetInt64("seed")
		if _seed != -1 {
			seed = &_seed
		}

		addr, _ := crypto.NewAddress(seed)
		newAddr := addr.Addr()

		fmt.Println(fmt.Sprintf("Address:     %s", color.HiCyanString(newAddr)))
		fmt.Println(fmt.Sprintf("Public Key:  %s", addr.PubKey().Base58()))
		fmt.Println(fmt.Sprintf("Private Key: %s", addr.PrivKey().Base58()))
		fmt.Println(fmt.Sprintf("Peer ID:     %s", addr.PeerID()))
	},
}

func init() {
	addressCmd.AddCommand(addressNewCmd)
	rootCmd.AddCommand(addressCmd)
	addressNewCmd.Flags().Int64P("seed", "s", -1, "Set a random seed")
}
