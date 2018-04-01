package peer

import (
	"bufio"

	pb "github.com/ellcrys/gcoin/modules/pb"
	"github.com/kr/pretty"
	net "github.com/libp2p/go-libp2p-net"
	pc "github.com/multiformats/go-multicodec/protobuf"
	"go.uber.org/zap"
)

// Inception represents the peer protocol
type Inception struct {
	version string
	peer    *Peer
	log     *zap.SugaredLogger
}

// NewInception creates a new instance of this protocol
// with a version it is supposed to handle
func NewInception(p *Peer) *Inception {
	return &Inception{
		peer: p,
		log:  peerLog.Named("protocol.inception"),
	}
}

// GetLocalPeer returns the local peer
func (protoc *Inception) GetLocalPeer() *Peer {
	return protoc.peer
}

// HandleHandshake handles incoming handshake request
func (protoc *Inception) HandleHandshake(s net.Stream) {

	msg := &pb.Handshake{}
	if err := pc.Multicodec(nil).Decoder(bufio.NewReader(s)).Decode(msg); err != nil {
		protoc.log.Errorf("failed to read message from %s -> %s", s.Conn().RemotePeer().Pretty(), err)
		return
	}

	pretty.Println(msg)
}
