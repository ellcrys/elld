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
	gmx        *sync.Mutex
	pm         *Manager
	activeConn int64
	log        logger.Logger
	connEstInt *time.Ticker
}

// NewConnMrg creates a new connection manager
func NewConnMrg(m *Manager, log logger.Logger) *ConnectionManager {
	return &ConnectionManager{
		pm:  m,
		gmx: &sync.Mutex{},
		log: log,
	}
}

// Manage starts connection management
func (m *ConnectionManager) Manage() {
	go m.establishConnections()
}

// connectionCount returns the number of active connections
func (m *ConnectionManager) connectionCount() int64 {
	m.gmx.Lock()
	defer m.gmx.Unlock()
	return m.activeConn
}

// needMoreConnections checks whether the local peer needs new connections
func (m *ConnectionManager) needMoreConnections() bool {
	return m.connectionCount() < m.pm.config.Node.MaxConnections
}

// establishConnections will attempt to send a handshake to
// addresses that have not been connected to as long as the max
// connection limit has not been reached
func (m *ConnectionManager) establishConnections() {
	m.connEstInt = time.NewTicker(time.Duration(m.pm.config.Node.ConnEstInterval) * time.Second)
	for {
		select {
		case <-m.connEstInt.C:
			unconnectedPeers := m.pm.GetUnconnectedPeers()
			if !m.pm.NeedMorePeers() || len(unconnectedPeers) == 0 {
				continue
			}
			m.log.Info("Establishing connection with more peers", "UnconnectedPeers", len(unconnectedPeers))
			for _, p := range unconnectedPeers {
				m.pm.ConnectToPeer(p.StringID())
			}
		}
	}
}

// Listen is called when network starts listening on an address
func (m *ConnectionManager) Listen(net.Network, ma.Multiaddr) {

}

// ListenClose is called when network stops listening on an address
func (m *ConnectionManager) ListenClose(net.Network, ma.Multiaddr) {

}

// Connected is called when a connection is opened
func (m *ConnectionManager) Connected(net net.Network, conn net.Conn) {
	m.gmx.Lock()
	defer m.gmx.Unlock()
	m.activeConn++
}

// Disconnected is called when a connection is closed
func (m *ConnectionManager) Disconnected(net net.Network, conn net.Conn) {
	fullAddr := util.FullRemoteAddressFromConn(conn)
	go m.pm.OnPeerDisconnect(fullAddr)

	m.gmx.Lock()
	defer m.gmx.Unlock()
	m.activeConn--
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
func (m *ConnectionManager) ClosedStream(nt net.Network, s net.Stream) {
}
