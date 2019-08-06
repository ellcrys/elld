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
	"os/signal"
	path "path/filepath"
	"sync"

	"github.com/spf13/viper"

	"github.com/ellcrys/elld/accountmgr"

	"github.com/ellcrys/elld/util/logger"

	"github.com/ellcrys/elld/config"
	"github.com/spf13/cobra"
)

var (
	// BuildVersion is the build version
	// set by goreleaser
	BuildVersion = ""

	// BuildCommit is the git hash of
	// the build. It is set by goreleaser
	BuildCommit = ""

	// BuildDate is the date the build
	// was created. Its is set by goreleaser
	BuildDate = ""

	// GoVersion is the version of go
	// used to build the client
	GoVersion = "go1.10.4"
)

var (
	cfg                    *config.EngineConfig
	consoleHistoryFilePath string
	log                    logger.Logger
	devMode                bool
	sig                    chan os.Signal
	accountMgr             *accountmgr.AccountManager
	onTerminate            func()
	interrupt              = make(chan struct{})
	mtx                    = &sync.Mutex{}
)

func setTerminateFunc(f func()) {
	mtx.Lock()
	defer mtx.Unlock()
	onTerminate = f
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "elld",
	Short: "ELLD is the Ellcrys protocol command line interface.",
	Long:  `ELLD is the Ellcrys protocol command line interface.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().String("net", config.DefaultNetVersion, "Set the network version")
	rootCmd.PersistentFlags().String("data-dir", "", "Set configuration directory")
	rootCmd.PersistentFlags().Bool("dev", false, "Run client in development mode")
	rootCmd.PersistentFlags().Bool("debug", false, "Set log level to DEBUG")
	rootCmd.PersistentFlags().Bool("cpu-profile", false, "Start CPU Profiling")
	rootCmd.PersistentFlags().Bool("mem-profile", false, "Start Memory Profiling")
	rootCmd.PersistentFlags().Bool("mutex-profile", false, "Start Mutex Profiling")
	viper.BindPFlag("net.version", rootCmd.PersistentFlags().Lookup("net"))
	cobra.OnInitialize(initConfig)

	sig = make(chan os.Signal, 1)
	signal.Reset()
	signal.Notify(sig, os.Interrupt)
	signalReceived := false

	go func() {
		for {
			select {
			case s := <-sig:
				if signalReceived {
					log.Info("Already received signal. Shutting down...", "Signal", s)
					continue
				}
				log.Info("Received signal. Shutting down...", "Signal", s)
				close(interrupt)
				signalReceived = true
			default:
			}
		}
	}()
}

func initConfig() {
	cfg = config.InitConfig(rootCmd)
	devMode = cfg.Node.Mode == config.ModeDev
	log = cfg.Log

	// Set account manager
	accountMgr = accountmgr.New(path.Join(cfg.DataDir(), "accounts"))

	// Set version information
	cfg.VersionInfo = &config.VersionInfo{}
	cfg.VersionInfo.BuildCommit = BuildCommit
	cfg.VersionInfo.BuildDate = BuildDate
	cfg.VersionInfo.GoVersion = GoVersion
	cfg.VersionInfo.BuildVersion = BuildVersion

	// Set the path where console history will be stored
	consoleHistoryFilePath = path.Join(cfg.NetDataDir(), ".console_history")
}
