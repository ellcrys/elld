package testutil

import (
	"context"
	"crypto/rand"
	"fmt"
	mrand "math/rand"
	"os"
	"path"
	"time"

	crypto "github.com/libp2p/go-libp2p-crypto"

	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/util"
	libp2p "github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-host"
	net "github.com/libp2p/go-libp2p-net"
	homedir "github.com/mitchellh/go-homedir"
)

// NoOpStreamHandler accepts a stream and does nothing
var NoOpStreamHandler = func(s net.Stream) {}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

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
	cfg, err := config.LoadCfg(dataDir)
	cfg.Node.Mode = config.ModeTest
	cfg.Node.MaxAddrsExpected = 5
	return cfg, err
}
