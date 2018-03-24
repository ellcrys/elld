package protocol

import (
	"fmt"

	"github.com/ellcrys/garagecoin/modules/types"

	"github.com/ellcrys/garagecoin/modules"
	"github.com/ellcrys/garagecoin/modules/peer"
	"github.com/ellcrys/garagecoin/modules/util"
	net "github.com/libp2p/go-libp2p-net"
	"go.uber.org/zap"
)

var (
	log *zap.SugaredLogger
)

func init() {
	log = modules.NewLogger("protocol/inception")
}

// Inception represents the first Garagecoin protocol
type Inception struct {
	version string
	peer    *peer.Peer
}

// NewInception creates a new instance of this protocol
// with a version it is supposed to handle
func NewInception(p *peer.Peer, version string) *Inception {
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
		protoc.HandleHandshake(m, s.Protocol(), s.Conn())
	}
}
