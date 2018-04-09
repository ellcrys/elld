package peer

import (
	"encoding/json"

	net "github.com/libp2p/go-libp2p-net"
	"go.uber.org/zap"
)

// Protocol represents a protocol
type Protocol interface {
	DoSendHandshake(*Peer)
	OnHandshake(net.Stream)
}

// Inception represents the peer protocol
type Inception struct {
	version string
	peer    *Peer
	log     *zap.SugaredLogger
}

// NewInception creates a new instance of the protocol codenamed "Inception"
func NewInception(p *Peer) *Inception {
	return &Inception{
		peer: p,
		log:  protocLog,
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
