package types

import (
	net "github.com/libp2p/go-libp2p-net"
	protocol "github.com/libp2p/go-libp2p-protocol"
)

// Protocol represents a protocol
type Protocol interface {
	GetVersion() string
	GetCodeName() string
	Handle(net.Stream)
	HandleHandshake(*Message, protocol.ID, net.Conn)
}
