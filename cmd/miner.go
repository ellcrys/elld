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
	"github.com/spf13/cobra"
)

// minerCmd represents the miner command
var minerCmd = &cobra.Command{
	Use:   "miner",
	Short: "Mining Algorithm for proof of work",
	Long: `An Ethash proof of work Algorith based on formerly Dagger-Hashimoto algorith
	It uses Dag file to speed up mining process
	go run main.go miner to run this package.`,

	Run: func(cmd *cobra.Command, args []string) {

	},
}

func init() {
	rootCmd.AddCommand(minerCmd)

}
