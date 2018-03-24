package types

import net "github.com/libp2p/go-libp2p-net"

// StreamProtocol represents a protocol
type StreamProtocol interface {
	GetVersion() string
	GetCodeName() string
	Handle(net.Stream)
	HandleHandshakeMsg()
}
