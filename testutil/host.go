package testutil

import (
	"context"
	"crypto/rand"
	"fmt"
	mrand "math/rand"

	"github.com/ellcrys/druid/peer"
	libp2p "github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-host"
	net "github.com/libp2p/go-libp2p-net"
)

// NoOpStreamHandler accepts a stream and does nothing
var NoOpStreamHandler = func(s net.Stream) {}

// RandomHost creates a host with random identity
func RandomHost(seed int64, port int) (host.Host, error) {

	priv, _, err := peer.GenerateKeyPair(mrand.New(mrand.NewSource(seed)))
	if seed == 0 {
		priv, _, _ = peer.GenerateKeyPair(rand.Reader)
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
