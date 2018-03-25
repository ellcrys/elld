package peer

import (
	"context"
	"fmt"

	"github.com/ellcrys/garagecoin/modules/types"

	"github.com/ellcrys/garagecoin/modules/util"
	pstore "github.com/libp2p/go-libp2p-peerstore"
)

// Do define rules and behaviours
// that are not provided by a stream protocol
type Do struct {
}

// SendHandShake sends a handshake request to a remote peer
func (b *Do) SendHandShake(p *Peer) error {

	// create a stream to remote peer
	p.localPeer.Peerstore().AddAddr(p.ID(), p.GetIP4Addr(), pstore.PermanentAddrTTL)
	s, err := p.localPeer.host.NewStream(context.Background(), p.ID(), p.localPeer.curProtocolVersion)
	if err != nil {
		return fmt.Errorf("handshake failed. failed to connect to peer (%s) -> %s", p.ID().String(), err)
	}
	defer s.Close()

	// create message object
	msg := types.NewMessage(types.OpHandshake, util.StructToBytes(types.HandshakeMsg{
		ID: "stuff",
	}))

	// write to peer. Message is encoded as hex
	_, err = s.Write(msg.Hex())
	if err != nil {
		return fmt.Errorf("handshake failed. failed to write to stream -> %s", err)
	}

	return nil
}
