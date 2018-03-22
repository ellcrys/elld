package inception

import (
	"fmt"

	"github.com/ellcrys/garagecoin/components"
	net "github.com/libp2p/go-libp2p-net"
	"go.uber.org/zap"
)

var (
	log *zap.SugaredLogger
)

func init() {
	log = components.NewLogger("protocol/inception")
}

// Inception represents the first Garagecoin protocol
type Inception struct {
	version string
}

// NewInception creates a new instance of this protocol
// with a version it is supposed to handle
func NewInception(version string) *Inception {
	return &Inception{version}
}

// GetCodeName returns the code name of the protocol
func (p *Inception) GetCodeName() string {
	return "inception"
}

// GetVersion returns the version
func (p *Inception) GetVersion() string {
	return p.version
}

// Handle handles incoming request
func (p *Inception) Handle(s net.Stream) {
	log.Info(fmt.Sprintf("Handling new request in #{%s}", p.version))
}
