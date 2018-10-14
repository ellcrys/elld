package node

import (
	"sync"
	"time"

	"github.com/ellcrys/elld/util/logger"

	"github.com/ellcrys/elld/util"
	net "github.com/libp2p/go-libp2p-net"
	ma "github.com/multiformats/go-multiaddr"
)

// ConnsInfo stores information about connections
// such as the number of inbound and outbound
// connections, etc
type ConnsInfo struct {
	Inbound  int
	Outbound int
}

// ConnectionManager manages the active connections
// ensuring the required number of connections at any given
// time is maintained
type ConnectionManager struct {
	sync.RWMutex
	pm         *Manager
	log        logger.Logger
	tickerDone chan bool
	connsInfo  *ConnsInfo
}

// NewConnMrg creates a new connection manager
func NewConnMrg(m *Manager, log logger.Logger) *ConnectionManager {
	return &ConnectionManager{
		pm:        m,
		log:       log,
		connsInfo: &ConnsInfo{},
	}
}

// Manage starts connection management
func (m *ConnectionManager) Manage() {
	go m.makeConnections(m.tickerDone)
}

// SetConnsInfo sets the connections information.
// Only used in tests.
func (m *ConnectionManager) SetConnsInfo(info *ConnsInfo) {
	m.connsInfo = info
}

// GetConnsCount gets the inbound and outbound
// connections count.
func (m *ConnectionManager) GetConnsCount() *ConnsInfo {
	m.RLock()
	defer m.RUnlock()
	return m.connsInfo
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

// Listen is called when hosts starts
// listening on an address
func (m *ConnectionManager) Listen(net.Network, ma.Multiaddr) {}

// ListenClose is called when host stops
// listening on an address
func (m *ConnectionManager) ListenClose(net.Network, ma.Multiaddr) {}

// Connected is called when a connection is opened.
// Check inbound and outbound connection count state
// and close connections when limits are reached.
func (m *ConnectionManager) Connected(n net.Network, conn net.Conn) {
	m.Lock()
	defer m.Unlock()

	if conn.Stat().Direction == net.DirInbound {
		m.connsInfo.Inbound++
		if int64(m.connsInfo.Inbound) > m.pm.config.Node.MaxInboundConnections {
			m.log.Debug("Closed inbound connection. Max. limit reached",
				"MaxAllowed", m.pm.config.Node.MaxInboundConnections)
			conn.Close()
		}
	}

	if conn.Stat().Direction == net.DirOutbound {
		m.connsInfo.Outbound++
		if int64(m.connsInfo.Outbound) > m.pm.config.Node.MaxOutboundConnections {
			m.log.Debug("Closed outbound connection. Max. limit reached",
				"MaxAllowed", m.pm.config.Node.MaxOutboundConnections)
			conn.Close()
		}
	}
}

// Disconnected is called when a connection is closed.
// Update the connection count and inform the peer
// manager of the disconnection event.
func (m *ConnectionManager) Disconnected(n net.Network, conn net.Conn) {

	m.Lock()
	if conn.Stat().Direction == net.DirInbound {
		m.connsInfo.Inbound--
	}

	if conn.Stat().Direction == net.DirOutbound {
		m.connsInfo.Outbound--
	}
	m.Unlock()

	addr := util.RemoteAddrFromConn(conn)
	m.pm.HasDisconnected(addr)
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
