package peer

import (
	"fmt"

	"github.com/ellcrys/gcoin/modules/types"

	"github.com/ellcrys/gcoin/modules"
	"github.com/ellcrys/gcoin/modules/util"
	net "github.com/libp2p/go-libp2p-net"
	"go.uber.org/zap"
)

var (
	log *zap.SugaredLogger
)

func init() {
	log = modules.NewLogger("protocol/inception")
}

// Inception represents the peer protocol
type Inception struct {
	version string
	peer    *Peer
}

// NewInception creates a new instance of this protocol
// with a version it is supposed to handle
func NewInception(p *Peer, version string) *Inception {
	return &Inception{peer: p, version: version}
}

// GetCodeName returns the code name of the protocol
func (protoc *Inception) GetCodeName() string {
	return "inception"
}

// GetVersion returns the version
func (protoc *Inception) GetVersion() string {
	return protoc.version
}

// GetLocalPeer returns the local peer
func (protoc *Inception) GetLocalPeer() *Peer {
	return protoc.peer
}

// Handle handles incoming request
func (protoc *Inception) Handle(s net.Stream) {
	log.Info(fmt.Sprintf("Received new message from peer #{%s}", protoc.version))

	// read message from the stream
	m, err := util.ReadMessageFromStream(s)
	if err != nil {
		s.Reset()
		log.Errorf(err.Error())
		return
	}

	// process message according to operation type
	switch op := m.Op; op {
	case types.OpHandshake:
		protoc.HandleHandshake(m, s)
		s.Write([]byte("Thanks"))
		s.Close()
	}
}
