package config

import (
	golog "log"
	"os"
	path "path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/util/logger"

	"github.com/mitchellh/go-homedir"
)

// SeedAddresses includes addresses to nodes that
// the client will attempt to synchronize with.
var SeedAddresses = []string{
	// "ellcrys://12D3KooWKAEhd4DXGPeN71FeSC1ih86Ym2izpoPueaCrME8xu8UM@n1.ellnode.com:9000",
	// "ellcrys://12D3KooWD276x1ieiV9cmtBdZeVLN5LtFrnUS6AT2uAkHHFNADRx@n2.ellnode.com:9000",
	// "ellcrys://12D3KooWDdUZny1FagkUregeNQUb8PB6Vg1LMWcwWquqovm7QADb@n3.ellnode.com:9000",
	// "ellcrys://12D3KooWDWA4g8EXWWBSbWbefSu2RGttNh1QDpQYA7nCDnbVADP1@n4.ellnode.com:9000",
}

// AccountDirName is the name of the directory for storing accounts
var AccountDirName = "accounts"

// setDefaultConfig sets default config values.
// They are used when their values is not provided
// in flag, env or config file.
func setDefaultConfig() {
	viper.SetDefault("net.version", DefaultNetVersion)
	viper.SetDefault("node.getAddrInt", 300)
	viper.SetDefault("node.pingInt", 60)
	viper.SetDefault("node.selfAdvInt", 120)
	viper.SetDefault("node.cleanUpInt", 1200)
	viper.SetDefault("node.maxAddrsExpected", 1000)
	viper.SetDefault("node.maxOutConnections", 10)
	viper.SetDefault("node.maxInConnections", 115)
	viper.SetDefault("node.conEstInt", 10)
	viper.SetDefault("node.messageTimeout", 30)
	viper.SetDefault("txPool.capacity", 10000)
	viper.SetDefault("miner.mode", 0)
	viper.SetDefault("rpc.username", "admin")
	viper.SetDefault("rpc.password", "admin")
	viper.SetDefault("rpc.sessionSecretKey", util.RandString(32))
}

func setDevDefaultConfig() {
	viper.SetDefault("node.getAddrInt", 10)
	viper.SetDefault("node.pingInt", 60)
	viper.SetDefault("node.selfAdvInt", 10)
	viper.SetDefault("node.cleanUpInt", 10)
	viper.SetDefault("node.conEstInt", 10)
	viper.SetDefault("txPool.capacity", 100)
}

// InitConfig reads in config file and ENV variables if set.
func InitConfig(rootCommand *cobra.Command) *EngineConfig {
	var c = EngineConfig{Node: &NodeConfig{Mode: ModeProd}}
	var homeDir, _ = homedir.Dir()
	var dataDir = path.Join(homeDir, ".ellcrys")
	devMode, _ := rootCommand.Flags().GetBool("dev")

	// If data directory path is set in a flag, update the default data directory
	dataDirF, _ := rootCommand.PersistentFlags().GetString("datadir")
	if dataDirF != "" {
		dataDir = dataDirF
	}

	// In development mode, use the development data directory.
	// Attempt to create the directory
	if devMode {
		dataDir, _ = homedir.Expand(path.Join("~", "ellcrys_dev"))
		c.Node.Mode = ModeDev
	}

	// Create the data directory and other sub directories
	os.MkdirAll(dataDir, 0700)
	os.MkdirAll(path.Join(dataDir, AccountDirName), 0700)

	// Set viper configuration
	setDefaultConfig()
	if devMode {
		setDevDefaultConfig()
	}

	viper.SetConfigName("ellcrys")
	viper.AddConfigPath(dataDir)
	viper.AddConfigPath(".")
	viper.SetEnvPrefix("ELLD")
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.AutomaticEnv()

	// Create the config file if it does not exist
	if err := viper.ReadInConfig(); err != nil {
		if strings.Index(err.Error(), "Not Found") != -1 {
			viper.SetConfigType("yaml")
			if err = viper.WriteConfigAs(path.Join(dataDir, "ellcrys.yml")); err != nil {
				golog.Fatalf("Failed to create config file: %s", err)
			}
		} else {
			golog.Fatalf("Failed to read config file: %s", err)
		}
	}

	// Read the loaded config into EngineConfig
	if err := viper.Unmarshal(&c); err != nil {
		golog.Fatalf("Failed to unmarshal configuration file: %s", err)
	}

	// Set network version environment variable
	// if not already set and then reset protocol
	// handlers version.
	SetVersions(viper.GetString("net.version"))

	// Set data and network directories
	c.SetDataDir(dataDir)
	c.SetNetDataDir(path.Join(dataDir, viper.GetString("net.version")))

	// Create network data directory
	os.MkdirAll(c.NetDataDir(), 0700)

	// Create logger with file rotation enabled
	logPath := path.Join(c.NetDataDir(), "logs")
	os.MkdirAll(logPath, 0700)
	logFile := path.Join(logPath, "elld.log")
	c.Log = logger.NewLogrusWithFileRotation(logFile)

	// Set default version information
	c.VersionInfo = &VersionInfo{}
	c.VersionInfo.BuildCommit = ""
	c.VersionInfo.BuildDate = ""
	c.VersionInfo.GoVersion = "go0"
	c.VersionInfo.BuildVersion = ""

	// Initialize miner config
	c.Miner = &MinerConfig{}

	// set connections hard limit
	if c.Node.MaxOutboundConnections > 10 {
		c.Node.MaxOutboundConnections = 10
	}
	if c.Node.MaxInboundConnections > 115 {
		c.Node.MaxInboundConnections = 115
	}

	return &c
}
