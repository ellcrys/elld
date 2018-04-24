package peer

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/ellcrys/druid/util/logger"

	"github.com/ellcrys/druid/wire"
	ic "github.com/libp2p/go-libp2p-crypto"
	net "github.com/libp2p/go-libp2p-net"
)

// Protocol represents a protocol
type Protocol interface {
	SendHandshake(*Peer) error
	OnHandshake(net.Stream)
	SendPing([]*Peer)
	OnPing(net.Stream)
	SendGetAddr([]*Peer) error
	OnGetAddr(net.Stream)
	OnAddr(net.Stream)
	RelayAddr([]*wire.Address) error
}

// Inception represents the peer protocol
type Inception struct {
	arm                         *sync.Mutex   // addr relay mutex
	version                     string        // the protocol version
	peer                        *Peer         // the local peer
	log                         logger.Logger // the logger
	lastRelayPeersSelectionTime time.Time     // the time the last addr msg relay peers where selected
	addrRelayPeers              [2]*Peer      // peers to relay addr msgs to
}

// NewInception creates a new instance of the protocol codenamed "Inception"
func NewInception(p *Peer, log logger.Logger) *Inception {
	return &Inception{
		peer: p,
		log:  log,
		arm:  &sync.Mutex{},
	}
}

// LocalPeer returns the local peer
func (protoc *Inception) LocalPeer() *Peer {
	return protoc.peer
}

// PM returns the local peer's peer manager
func (protoc *Inception) PM() *Manager {
	if protoc.peer == nil {
		return nil
	}
	return protoc.peer.PM()
}

// sign takes an object, marshals it to JSON and signs it
func (protoc *Inception) sign(msg interface{}) []byte {
	bs, _ := json.Marshal(msg)
	key := protoc.LocalPeer().PrivKey()
	sig, _ := key.Sign(bs)
	return sig
}

// verify verifies a signature
func (protoc *Inception) verify(msg interface{}, sig []byte, pKey ic.PubKey) error {
	bs, _ := json.Marshal(msg)
	result, err := pKey.Verify(bs, sig)
	if err != nil {
		return fmt.Errorf("failed to verify -> %s", err)
	}
	if !result {
		return fmt.Errorf("invalid signature")
	}
	return nil
}
