// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
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
	path "path/filepath"
	"strconv"

	"github.com/thoas/go-funk"

	"github.com/ellcrys/elld/burner"
	"github.com/ellcrys/elld/util"

	"github.com/ellcrys/elld/accountmgr"
	"github.com/ellcrys/elld/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	validate "gopkg.in/asaskevich/govalidator.v4"
)

// burnCmd represents the burn command
var burnCmd = &cobra.Command{
	Use:   "burn <burner account>",
	Short: "Buy block producer ticket by burning litecoins.",
	Long: `Burn litecoins to acquire block producer tickets which will
give the ticket owner a chance to create new blocks. Purchasing block producer
ticket will lead to the purchase amount being burned.

<burner account> is the address or index of a burner account that will pay the cost
of acquiring the block producer ticket. It must have sufficient balance and must 
exist in the burner account key directory on the node.`,
	Run: func(cmd *cobra.Command, args []string) {

		viper.BindPFlag("node.id", cmd.Flags().Lookup("id"))
		viper.BindPFlag("producer.address", cmd.Flags().Lookup("producer"))
		viper.BindPFlag("burner.pass", cmd.Flags().Lookup("pass"))
		viper.BindPFlag("rpc.address", cmd.Flags().Lookup("rpc-address"))
		viper.BindPFlag("burn.fee", cmd.Flags().Lookup("fee"))
		nodeID := viper.GetString("node.id")
		burnerAccountPass := viper.GetString("burner.pass")
		rpcAddress := viper.GetString("rpc.address")
		fee := viper.GetFloat64("burn.fee")

		// The address of the producer account that will gain the block producer privilege
		producerAddress := viper.GetString("producer.address")

		// The amount of litecoins to burn
		amount, _ := cmd.Flags().GetString("amount")

		am := accountmgr.New(path.Join(cfg.DataDir(), config.AccountDirName))

		// Ensure the purchasing account is provided
		if len(args) == 0 {
			log.Fatal("Address of the purchasing burner account is required.")
		}

		// Ensure block producer candidate account is provided
		if validate.IsNull(producerAddress) {
			util.PrintCLIError("Address of the block producer account is required.")
			return
		}

		// Ensure amount is not empty and it is numeric
		if validate.IsNull(amount) {
			util.PrintCLIError("Amount of litecoins to burn is required.")
			return
		} else if !validate.IsFloat(amount) && !validate.IsNumeric(amount) {
			util.PrintCLIError("Burn amount must be numeric.")
			return
		}

		// Fetch burner accounts on this client
		accounts, err := am.ListBurnerAccounts()
		if err != nil {
			util.PrintCLIError("Failed to read burner accounts: %s.", err)
			return
		}

		var selectedBurnAccount *accountmgr.StoredAccount
		burnerAddress := args[0]

		// If burner address is numeric, we are expected to use the
		// burner account at the specified index
		if validate.IsNumeric(burnerAddress) {
			index, _ := strconv.Atoi(burnerAddress)
			if len(accounts)-1 < index {
				util.PrintCLIError("Burner account index is out of range.")
				return
			}
			selectedBurnAccount = accounts[index]
		} else {
			result := funk.Find(accounts, func(account *accountmgr.StoredAccount) bool {
				if account.Address == burnerAddress {
					return true
				}
				return false
			})
			if result != nil {
				selectedBurnAccount = result.(*accountmgr.StoredAccount)
			}
		}

		// Return error if the burner address specified was not found
		if selectedBurnAccount == nil {
			util.PrintCLIError("Burner account (%s) does not exist.", burnerAddress)
			return
		}

		// Fetch regular accounts on this client
		accounts, err = am.ListAccounts()
		if err != nil {
			util.PrintCLIError("Failed to read burner accounts: %s.", err)
			return
		}

		var selectedProducerAccount *accountmgr.StoredAccount

		// If block producer address is numeric, we are expected
		// to use the account at the specified index
		if validate.IsNumeric(producerAddress) {
			index, _ := strconv.Atoi(producerAddress)
			if len(accounts)-1 < index {
				util.PrintCLIError("Coinbase account index is out of range.")
				return
			}
			selectedProducerAccount = accounts[index]
		} else {
			result := funk.Find(accounts, func(account *accountmgr.StoredAccount) bool {
				if account.Address == producerAddress {
					return true
				}
				return false
			})
			if result != nil {
				selectedProducerAccount = result.(*accountmgr.StoredAccount)
			}
		}

		// Return error if the producer address specified was not found
		if selectedProducerAccount == nil {
			util.PrintCLIError("Producer account (%s) does not exist.", producerAddress)
			return
		}

		// if node id is not set, use default account
		if nodeID == "" {
			defAct, err := accountMgr.GetDefault()
			if err != nil {
				util.PrintCLIError(`No default account found. Node environment has ` +
					`not been initialized. Run 'elld init' to initialize the node or specify ` +
					`an existing account using '--account' flag.`)
				return
			}
			nodeID = defAct.Address
		}

		// Now that we have all the right information we need,
		// run the burn command business logic.
		burner.BurnCmd(cfg,
			selectedProducerAccount,
			rpcAddress,
			selectedBurnAccount,
			burnerAccountPass,
			producerAddress,
			amount, fee)
	},
}

func init() {
	rootCmd.AddCommand(burnCmd)
	burnCmd.Flags().String("producer", "", "The address that will be granted block producer privilege.")
	burnCmd.Flags().String("amount", "", "The amount of coins to burn.")
	burnCmd.Flags().String("pass", "", "The password used to open the <burner account> (file path also accepted).")
	burnCmd.Flags().String("rpc-address", "127.0.0.1:8999", "The Address node's RPC server will listen on")
	burnCmd.Flags().Float64("fee", 0, "The amount of fee to pay for the burn transaction.")
}
