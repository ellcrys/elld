package peer

import (
	"context"
	"fmt"

	"github.com/ellcrys/garagecoin/modules/types"
	"github.com/ellcrys/garagecoin/modules/util"
	"github.com/kr/pretty"
	net "github.com/libp2p/go-libp2p-net"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	protocol "github.com/libp2p/go-libp2p-protocol"
)

// SendHandshake sends an introduction message to a peer
func SendHandshake(p *Peer) error {

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
		s.Reset()
		return fmt.Errorf("handshake failed. failed to write to stream -> %s", err)
	}

	// wait for response
	util.WaitThenReadStream(util.WaitTimeBeforeRead, s, func(err error, bs []byte) {
		pretty.Println(err, string(bs))
	})

	return nil
}

// HandleHandshake processes handshake message from a remote peer
func (protoc *Inception) HandleHandshake(m *types.Message, protocol protocol.ID, conn net.Conn) {
	var opMsg types.HandshakeMsg
	m.Scan(&opMsg)
	pretty.Println(opMsg)
}
