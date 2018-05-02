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
	"os"
	"os/signal"
	"path"
	"syscall"

	"github.com/ellcrys/druid/crypto"
	homedir "github.com/mitchellh/go-homedir"

	"github.com/ellcrys/druid/util"
	"github.com/ellcrys/druid/util/logger"

	"github.com/ellcrys/druid/configdir"
	"github.com/spf13/cobra"
)

var (
	cfg         *configdir.Config
	log         logger.Logger
	seed        int64
	devMode     bool
	sigs        chan os.Signal
	done        chan bool
	onTerminate func()
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "druid",
	Short: "druid is a cryptocurrency based on the original bitcoin protocol",
	Long:  `druid is a cryptocurrency based on the original bitcoin protocol`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	sigs = make(chan os.Signal, 1)
	done = make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		if onTerminate != nil {
			onTerminate()
		}
		done <- true
	}()

	rootCmd.PersistentFlags().String("cfgdir", "", "Set configuration directory")
	rootCmd.PersistentFlags().Int64P("seed", "s", 0, "Random seed to use for identity creation")
	rootCmd.PersistentFlags().Bool("dev", false, "Run client in development mode")
	cobra.OnInitialize(initConfig)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {

	var err error

	log = logger.NewLogrus()

	devMode, _ = rootCmd.Flags().GetBool("dev")
	seed, _ = rootCmd.Flags().GetInt64("seed")
	cfgDirPath, _ := rootCmd.Root().PersistentFlags().GetString("cfgdir")

	if devMode && cfgDirPath == "" {
		addr, _ := crypto.NewAddress(&seed)
		cfgDirPath, _ = homedir.Expand(path.Join("~", "ellcry_dev_"+addr.PeerID()[42:]))
		os.MkdirAll(cfgDirPath, 0700)
	}

	cfg, err = util.LoadCfg(cfgDirPath)
	if err != nil {
		log.Fatal(err.Error())
	}

	cfg.Node.Test = false
}
