package node

import (
	"github.com/ellcrys/elld/txpool"

	"github.com/ellcrys/elld/wire"
	net "github.com/libp2p/go-libp2p-net"
)

// GossipProtocol represents a protocol
type GossipProtocol interface {

	// messaging
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

	// transaction session
	AddTxSession(txID string)
	HasTxSession(txID string) bool
	RemoveTxSession(txID string)
	CountTxSession() int

	// Tx Pool
	GetUnSignedTxPool() *txpool.TxPool

	// Tx Relay Queue
	GetUnsignedTxRelayQueue() *txpool.TxQueue
}
