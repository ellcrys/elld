package types

import (
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/wire/messages"
	net "github.com/libp2p/go-libp2p-net"
)

// Gossip defines an interface for a gossip protocol
type Gossip interface {
	SendHandshake(Engine) error
	OnHandshake(net.Stream)
	SendPing([]Engine)
	OnPing(net.Stream)
	SendGetAddr([]Engine) error
	OnGetAddr(net.Stream)
	OnAddr(net.Stream)
	RelayAddr([]*messages.Address) []error
	SelfAdvertise([]Engine) int
	OnTx(net.Stream)
	RelayTx(core.Transaction, []Engine) error
	RelayBlock(core.Block, []Engine) error
	RequestBlock(remotePeer Engine, blockHash util.Hash) error
}
