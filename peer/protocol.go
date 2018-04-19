package peer

import (
	"encoding/json"
	"fmt"

	ic "github.com/libp2p/go-libp2p-crypto"
	net "github.com/libp2p/go-libp2p-net"
	"go.uber.org/zap"
)

// Protocol represents a protocol
type Protocol interface {
	SendHandshake(*Peer) error
	OnHandshake(net.Stream)
	SendPing([]*Peer)
	OnPing(net.Stream)
	SendGetAddr([]*Peer) error
	OnGetAddr(net.Stream)
}

// Inception represents the peer protocol
type Inception struct {
	version string
	peer    *Peer
	log     *zap.SugaredLogger
}

// NewInception creates a new instance of the protocol codenamed "Inception"
func NewInception(p *Peer, log *zap.SugaredLogger) *Inception {
	return &Inception{
		peer: p,
		log:  log,
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
