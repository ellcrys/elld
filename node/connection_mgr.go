package node

import (
	"sync"
	"time"

	"github.com/ellcrys/elld/util/logger"

	"github.com/ellcrys/elld/util"
	net "github.com/libp2p/go-libp2p-net"
	ma "github.com/multiformats/go-multiaddr"
)

// ConnectionManager manages the active connections
// ensuring the required number of connections at any given
// time is maintained
type ConnectionManager struct {
	sync.RWMutex
	pm         *Manager
	log        logger.Logger
	tickerDone chan bool
}

// NewConnMrg creates a new connection manager
func NewConnMrg(m *Manager, log logger.Logger) *ConnectionManager {
	return &ConnectionManager{
		pm:  m,
		log: log,
	}
}

// Manage starts connection management
func (m *ConnectionManager) Manage() {
	go m.makeConnections(m.tickerDone)
}

// makeConnections will attempt to send a handshake to
// addresses that have not been connected to as long
// as the max connection limit has not been reached
func (m *ConnectionManager) makeConnections(done chan bool) {
	dur := time.Duration(m.pm.config.Node.ConnEstInterval)
	ticker := time.NewTicker(dur * time.Second)
	for {
		select {
		case <-ticker.C:
			unconnectedPeers := m.pm.GetUnconnectedPeers()
			if !m.pm.RequirePeers() || len(unconnectedPeers) == 0 {
				continue
			}
			m.log.Info("Establishing connection with more peers",
				"UnconnectedPeers", len(unconnectedPeers))
			for _, p := range unconnectedPeers {
				m.pm.ConnectToPeer(p.StringID())
			}
		case <-done:
			ticker.Stop()
			return
		}
	}
}

// Listen is called when hosts starts listening on an address
func (m *ConnectionManager) Listen(net.Network, ma.Multiaddr) {}

// ListenClose is called when host stops listening on an address
func (m *ConnectionManager) ListenClose(net.Network, ma.Multiaddr) {}

// Connected is called when a connection is opened
func (m *ConnectionManager) Connected(net net.Network, conn net.Conn) {

	// if !m.needConnections() {
	// 	m.log.Debug("Closed unneeded connection")
	// 	conn.Close()
	// }

}

// Disconnected is called when a connection is closed
func (m *ConnectionManager) Disconnected(net net.Network, conn net.Conn) {
	addr := util.RemoteAddrFromConn(conn)
	m.pm.OnPeerDisconnect(addr)
}

// OpenedStream is called when a stream is openned
func (m *ConnectionManager) OpenedStream(n net.Network, s net.Stream) {

	// If the local node has stopped,
	// close the stream on both ends.
	if m.pm.localNode.HasStopped() {
		s.Reset()
	}
}

// ClosedStream is called when a stream is openned
func (m *ConnectionManager) ClosedStream(nt net.Network, s net.Stream) {}
