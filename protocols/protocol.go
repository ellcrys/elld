package protocols

import net "github.com/libp2p/go-libp2p-net"

// Protocol represents a protocol
type Protocol interface {
	GetVersion() string
	GetCodeName() string
	Handle(net.Stream)
}
