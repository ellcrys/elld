package node

import (
	"github.com/ellcrys/druid/wire"
	net "github.com/libp2p/go-libp2p-net"
)

// OnTx handles incoming transaction message
func (pt *Inception) OnTx(s net.Stream) {

}

// RelayTx relays transactions to peers
func (pt *Inception) RelayTx(tx *wire.Transaction) error {
	return nil
}
