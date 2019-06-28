package testutil

import (
	"context"
	"crypto/rand"
	"fmt"
	mrand "math/rand"
	"os"
	"path"
	"time"

	"github.com/spf13/viper"

	"github.com/spf13/cobra"

	crypto "github.com/libp2p/go-libp2p-crypto"

	"github.com/ellcrys/mother/config"
	"github.com/ellcrys/mother/util"
	libp2p "github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-host"
	net "github.com/libp2p/go-libp2p-net"
	homedir "github.com/mitchellh/go-homedir"
)

// NoOpStreamHandler accepts a stream and does nothing
var NoOpStreamHandler = func(s net.Stream) {}

func init() {
	mrand.Seed(time.Now().UnixNano())
}

// RandomHost creates a host with random identity
func RandomHost(seed int64, port int) (host.Host, error) {

	priv, _, err := crypto.GenerateEd25519Key(mrand.New(mrand.NewSource(seed)))
	if seed == 0 {
		priv, _, _ = crypto.GenerateEd25519Key(rand.Reader)
	}

	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", port)),
		libp2p.Identity(priv),
	}

	host, err := libp2p.New(context.Background(), opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create host")
	}

	return host, nil
}

// SetTestCfg prepare a config directory for tests
func SetTestCfg() (*config.EngineConfig, error) {
	var err error
	dir, _ := homedir.Dir()
	dataDir := path.Join(dir, util.RandString(5))
	os.MkdirAll(dataDir, 0700)

	// Create test root command and
	// set required flags and values
	rootCmd := &cobra.Command{}
	rootCmd.PersistentFlags().String("net", config.DefaultNetVersion, "Set the network version")
	rootCmd.PersistentFlags().String("datadir", "", "Set configuration directory")
	rootCmd.PersistentFlags().Set("datadir", dataDir)
	rootCmd.PersistentFlags().Set("net", dataDir)
	viper.Set("net.version", "test")

	// Initialize the config using the test root command
	cfg := config.InitConfig(rootCmd)
	cfg.Node.Mode = config.ModeTest
	cfg.Node.MaxAddrsExpected = 5
	return cfg, err
}
