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
	"github.com/ellcrys/elld/config"
	"github.com/pkg/profile"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// consoleCmd represents the console command
var consoleCmd = &cobra.Command{
	Use:   "console",
	Short: "Starts the node and an interactive Javascript console",
	Long: `Description:
  Starts the node and an interactive Javascript console.
  
  Set the listening address on the node using '--address' flag. 
	
  Use '--addnode' to provide a comma separated list of initial addresses of peers
  to connect to. Addresses must be valid ipfs multiaddress. An account must be 
  provided and unlocked to be used for signing transactions and blocks. Use '--account'
  flag to provide the account. If account is not provided, the default account account 
  (oldest account) in <CONFIGDIR>/` + config.AccountDirName + ` is used instead.
	
  If no account was found, an interactive session to create an account is started.   
	
  Account password will be interactively requested during account creation and unlock
  operations. Use '--pwd' flag to provide the account password non-interactively. '--pwd'
  can also accept a path to a file containing the password.`,
	Run: func(cmd *cobra.Command, args []string) {

		profilePath := profile.ProfilePath(cfg.NetDataDir())
		viper.BindPFlag("debug.cpuProfile", cmd.Flags().Lookup("cpu-profile"))
		viper.BindPFlag("debug.memProfile", cmd.Flags().Lookup("mem-profile"))
		viper.BindPFlag("debug.mutexProfile", cmd.Flags().Lookup("mutex-profile"))
		cpuProfile := viper.GetBool("debug.cpuProfile")
		memProfile := viper.GetBool("debug.memProfile")
		mtxProfile := viper.GetBool("debug.mutexProfile")

		if cpuProfile {
			defer profile.Start(profile.CPUProfile, profilePath).Stop()
		}

		if memProfile {
			defer profile.Start(profile.MemProfile, profilePath).Stop()
		}

		if mtxProfile {
			defer profile.Start(profile.MutexProfile, profilePath).Stop()
		}

		node, rpcServer, cs, miner := start(cmd, args, true)
		cs.OnStop(func() {
			if miner != nil {
				miner.Stop()
			}
			rpcServer.Stop()
			node.Stop()
		})

		node.Wait()
	},
}

func init() {
	rootCmd.AddCommand(consoleCmd)
	consoleCmd.Flags().StringSliceP("add-node", "j", nil, "IP of a node to connect to")
	consoleCmd.Flags().StringP("address", "a", "127.0.0.1:9000", "Address to listen on")
	consoleCmd.Flags().Bool("rpc", false, "Launch RPC server")
	consoleCmd.Flags().String("rpc-address", "127.0.0.1:8999", "Address RPC server will listen on")
	consoleCmd.Flags().Bool("rpc-disable-auth", false, "Disable RPC authentication (not recommended)")
	consoleCmd.Flags().String("account", "", "Coinbase account to load. An ephemeral account is used as default.")
	consoleCmd.Flags().Int64P("seed", "s", 0, "Provide a strong seed for network account creation (not recommended)")
	consoleCmd.Flags().String("pwd", "", "Used as password during initial account creation or loading an account")
	consoleCmd.Flags().Bool("mine", false, "Start Blake2 CPU mining")
	consoleCmd.Flags().Int("miners", 0, "The number of miner threads to use. (Default: Number of CPU)")
	consoleCmd.Flags().Bool("no-net", false, "Closes the network host and prevents (in/out) connections")
}
