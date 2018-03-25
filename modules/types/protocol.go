package types

import (
	net "github.com/libp2p/go-libp2p-net"
	protocol "github.com/libp2p/go-libp2p-protocol"
)

// StreamProtocol represents a protocol
type StreamProtocol interface {
	GetVersion() string
	GetCodeName() string
	Handle(net.Stream)
	HandleHandshake(*Message, protocol.ID, net.Conn)
}
