package testutil

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	mrand "math/rand"
	"os"
	"path"
	"time"

	"github.com/ellcrys/elld/blockchain/common"
	elldCrypto "github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/util"
	crypto "github.com/libp2p/go-libp2p-crypto"

	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/wire"
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

// GenerateKeyPair generates private and public keys
func GenerateKeyPair(r io.Reader) (crypto.PrivKey, crypto.PubKey, error) {
	return crypto.GenerateEd25519Key(r)
}

// RandomHost creates a host with random identity
func RandomHost(seed int64, port int) (host.Host, error) {

	priv, _, err := GenerateKeyPair(mrand.New(mrand.NewSource(seed)))
	if seed == 0 {
		priv, _, _ = GenerateKeyPair(rand.Reader)
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

// RandString is like RandBytes but returns string
func RandString(n int) string {
	return string(RandBytes(n))
}

// RandBytes gets random string of fixed length
func RandBytes(n int) []byte {
	b := make([]byte, n)
	for i, cache, remain := n-1, mrand.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = mrand.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return b
}

// SetTestCfg prepare a config directory for tests
func SetTestCfg() (*config.EngineConfig, error) {
	var err error
	dir, _ := homedir.Dir()
	cfgDir := path.Join(dir, ".ellcrys_test")
	os.MkdirAll(cfgDir, 0700)
	cfg, err := config.LoadCfg(cfgDir)
	cfg.Node.Dev = true
	cfg.Node.MaxAddrsExpected = 5
	cfg.Node.Test = true
	return cfg, err
}

// RemoveTestCfgDir removes test config directory
func RemoveTestCfgDir() error {
	dir, _ := homedir.Dir()
	err := os.RemoveAll(path.Join(dir, ".ellcrys_test"))
	return err
}

// GenerateBlock generates a block and returns the json equivalent.
// Specifically meant for create valid blocks for testing.
func GenerateBlock(to *elldCrypto.Key, from *elldCrypto.Key, value string) string {

	now := time.Now()
	block := &wire.Block{}
	block.Transactions = append(block.Transactions, &wire.Transaction{
		Type:         wire.TxTypeBalance,
		Nonce:        0,
		To:           to.Addr(),
		From:         from.Addr(),
		SenderPubKey: from.PubKey().Base58(),
		Value:        value,
		Timestamp:    now.Unix(),
		Fee:          "0.1",
	})
	block.Transactions[0].Hash = block.Transactions[0].ComputeHash2()
	sig, _ := wire.TxSign(block.Transactions[0], from.PrivKey().Base58())
	block.Transactions[0].Sig = util.ToHex(sig)

	block.Header = &wire.Header{
		ParentHash:       "",
		CreatorPubKey:    to.PubKey().Base58(),
		Number:           1,
		TransactionsRoot: util.ToHex(common.ComputeTxsRoot(block.Transactions)),
		Nonce:            1030,
		Difficulty:       "2747646837",
		Timestamp:        now.Unix(),
		MixHash:          "",
	}

	block.Hash = block.ComputeHash()
	sig, _ = wire.BlockSign(block, to.PrivKey().Base58())
	block.Sig = util.ToHex(sig)

	bs, _ := json.MarshalIndent(block, "", "\t")
	return string(bs)
}
