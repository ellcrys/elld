package peer

import (
	"encoding/json"

	"go.uber.org/zap"
)

// Protocol represents the peer protocol
type Protocol struct {
	version string
	peer    *Peer
	log     *zap.SugaredLogger
}

// NewProtocol creates a new instance of this protocol
// with a version it is supposed to handle
func NewProtocol(p *Peer) *Protocol {
	return &Protocol{
		peer: p,
		log:  protocLog,
	}
}

// LocalPeer returns the local peer
func (protoc *Protocol) LocalPeer() *Peer {
	return protoc.peer
}

// PM returns the local peer's peer manager
func (protoc *Protocol) PM() *Manager {
	if protoc.peer == nil {
		return nil
	}
	return protoc.peer.PM()
}

// sign takes an object, marshals it to JSON and signs it
func (protoc *Protocol) sign(msg interface{}) []byte {
	bs, _ := json.Marshal(msg)
	key := protoc.LocalPeer().PrivKey()
	sig, _ := key.Sign(bs)
	return sig
}
