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
	path "path/filepath"

	"github.com/ellcrys/druid/accountmgr"
	"github.com/ellcrys/druid/configdir"
	"github.com/spf13/cobra"
)

// accountCmd represents the account command
var accountCmd = &cobra.Command{
	Use:   "account",
	Short: "Create and manage your accounts",
	Long: `NAME:
druid account -

This command provides the ability to create an account, list, import and update 
accounts. Accounts are stored in an encrypted format using a passphrase provided 
by you. Please understand that if you forget the password, it is IMPOSSIBLE to 
unlock your account.

Password will be stored under <CONFIGDIR>/` + configdir.AccountDirName + `. It is safe to transfer the 
directory or individual accounts to another node. 

Always backup your keeps regularly.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// accountCreateCmd represents the account command
var accountCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an account",
	Long: `NAME:
druid account create -

This command creates an account and encrypts it using a passphrase
you provide. Do not forget your your passphrase. You will not be able 
to unlock your account if you do.

Password will be stored under <CONFIGDIR>/` + configdir.AccountDirName + `. 
It is safe to transfer the directory or individual accounts to another node. 

Use --pwd to directly specify a password without going interactive mode. You 
can also provide a path to a file containing a password. If a path is provided,
password is fetched with leading and trailing newline character removed. 

Always backup your keeps regularly.`,
	Run: func(cmd *cobra.Command, args []string) {

		onTerminate = func() {
			os.Exit(-1)
		}

		pwd, _ := cmd.Flags().GetString("pwd")
		am := accountmgr.New(path.Join(cfg.ConfigDir(), configdir.AccountDirName))
		err := am.Create(pwd)
		if err != nil {
			log.Fatal(err.Error())
		}
	},
}

func init() {
	accountCmd.AddCommand(accountCreateCmd)
	accountCreateCmd.Flags().String("pwd", "", "Providing a password or path to a file containing a password (No interactive mode)")
	rootCmd.AddCommand(accountCmd)
}
