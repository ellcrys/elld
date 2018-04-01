package types

import (
	net "github.com/libp2p/go-libp2p-net"
)

// Protocol represents a protocol
type Protocol interface {
	HandleHandshake(net.Stream)
}
