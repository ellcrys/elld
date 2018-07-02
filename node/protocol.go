package node

import (
	"github.com/ellcrys/elld/wire"
	net "github.com/libp2p/go-libp2p-net"
)

// GossipProtocol represents a protocol
type GossipProtocol interface {
	SendHandshake(*Node) error
	OnHandshake(net.Stream)
	SendPing([]*Node)
	OnPing(net.Stream)
	SendGetAddr([]*Node) error
	OnGetAddr(net.Stream)
	OnAddr(net.Stream)
	RelayAddr([]*wire.Address) []error
	SelfAdvertise([]*Node) int
	OnTx(net.Stream)
	RelayTx(*wire.Transaction, []*Node) error
}
