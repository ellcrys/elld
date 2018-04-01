package peer

import (
	"bufio"
	"context"
	"fmt"

	pb "github.com/ellcrys/gcoin/modules/pb"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	protocol "github.com/libp2p/go-libp2p-protocol"
	pc "github.com/multiformats/go-multicodec/protobuf"
)

// HandshakeVersion is the current handshake protocol supported
var HandshakeVersion = "/inception/handshake/0.0.1"

// SendHandshake sends an introduction message to a peer
func SendHandshake(remotePeer *Peer) error {
	remotePeer.localPeer.Peerstore().AddAddr(remotePeer.ID(), remotePeer.GetIP4Addr(), pstore.PermanentAddrTTL)
	s, err := remotePeer.localPeer.host.NewStream(context.Background(), remotePeer.ID(), protocol.ID(HandshakeVersion))
	if err != nil {
		return fmt.Errorf("handshake failed. failed to connect to peer (%s) -> %s", remotePeer.IDPretty(), err)
	}
	defer s.Close()

	w := bufio.NewWriter(s)
	msg := &pb.Handshake{Address: remotePeer.localPeer.GetMultiAddr()}
	if err := pc.Multicodec(nil).Encoder(w).Encode(msg); err != nil {
		return fmt.Errorf("handshake failed. failed to write to stream -> %s", err)
	}

	w.Flush()
	return nil
}
