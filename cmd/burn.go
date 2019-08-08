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

	"github.com/ellcrys/elld/burner"

	"github.com/thoas/go-funk"

	"github.com/ellcrys/elld/util"

	"github.com/ellcrys/elld/accountmgr"
	"github.com/ellcrys/elld/config"
	"github.com/spf13/cobra"
	validate "gopkg.in/asaskevich/govalidator.v4"
)

// burnCmd represents the burn command
var burnCmd = &cobra.Command{
	Use:   "burn",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		burnerAddress, _ := cmd.Flags().GetString("account")
		coinbaseAddress, _ := cmd.Flags().GetString("coinbase")
		amount, _ := cmd.Flags().GetString("amount")
		am := accountmgr.New(path.Join(cfg.DataDir(), config.AccountDirName))

		// Ensure burner account is not empty
		if validate.IsNull(burnerAddress) {
			util.PrintCLIError("Burner account is required.")
			return
		}

		// Ensure coinbase account is not empty
		if validate.IsNull(coinbaseAddress) {
			util.PrintCLIError("Coinbase account is required.")
			return
		}

		// Ensure amount is not empty and it is numeric
		if validate.IsNull(amount) {
			util.PrintCLIError("Burn amount is required.")
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

		var selectedCoinbaseAccount *accountmgr.StoredAccount

		// If coinbase address is numeric, we are expected to use the
		// burner account at the specified index
		if validate.IsNumeric(coinbaseAddress) {
			index, _ := strconv.Atoi(coinbaseAddress)
			if len(accounts)-1 < index {
				util.PrintCLIError("Coinbase account index is out of range.")
				return
			}
			selectedCoinbaseAccount = accounts[index]
		} else {
			result := funk.Find(accounts, func(account *accountmgr.StoredAccount) bool {
				if account.Address == coinbaseAddress {
					return true
				}
				return false
			})
			if result != nil {
				selectedCoinbaseAccount = result.(*accountmgr.StoredAccount)
			}
		}

		// Return error if the coinbase address specified was not found
		if selectedCoinbaseAccount == nil {
			util.PrintCLIError("Coinbase account (%s) does not exist.", burnerAddress)
			return
		}

		pwd, _ := cmd.Flags().GetString("pwd")
		burner.BurnCmd(selectedCoinbaseAccount, selectedBurnAccount, pwd, amount)
	},
}

func init() {
	rootCmd.AddCommand(burnCmd)
	burnCmd.Flags().String("pwd", "", "Providing the burner account password or path to a file containing it (No interactive mode)")
	burnCmd.Flags().StringP("account", "a", "", "The Litecoin burner account to burn coins from. Accepts account index or address.")
	burnCmd.Flags().StringP("coinbase", "c", "", "The client account that will be granted mining privilege.")
	burnCmd.Flags().String("amount", "", "The amount of coins to burn")
}
