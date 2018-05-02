package testutil

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	mrand "math/rand"
	"time"

	libp2p "github.com/libp2p/go-libp2p"
	crypto "github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-host"
	net "github.com/libp2p/go-libp2p-net"
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
