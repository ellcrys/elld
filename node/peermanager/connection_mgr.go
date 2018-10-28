package peermanager

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
	sync.RWMutex
	inbound  int
	outbound int
}

// NewConnsInfo creates an instance of ConnsInfo
func NewConnsInfo(inbound, outbound int) *ConnsInfo {
	return &ConnsInfo{
		inbound:  inbound,
		outbound: outbound,
	}
}

// Info returns the number of inbound and outbound connections
func (i *ConnsInfo) Info() (inbound int, outbound int) {
	i.RLock()
	defer i.RUnlock()
	return i.inbound, i.outbound
}

// IncOutbound increments outbound count
func (i *ConnsInfo) IncOutbound() {
	i.Lock()
	defer i.Unlock()
	i.outbound++
}

// DecrOutbound decrements outbound count
func (i *ConnsInfo) DecrOutbound() {
	i.Lock()
	defer i.Unlock()
	i.outbound--
}

// IncInbound increments inbound count
func (i *ConnsInfo) IncInbound() {
	i.Lock()
	defer i.Unlock()
	i.inbound++
}

// DecrInbound decrements outbound count
func (i *ConnsInfo) DecrInbound() {
	i.Lock()
	defer i.Unlock()
	i.inbound--
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
		connsInfo: NewConnsInfo(0, 0),
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

			if !m.pm.RequirePeers() {
				continue
			}

			// Get unconnected/unacquainted peers
			peers := m.pm.GetLonelyPeers()
			if len(peers) == 0 {
				continue
			}

			m.log.Debug("Establishing connection with more peers", "UnconnectedPeers",
				len(peers))

			for _, p := range peers {

				// Do not establish connection with banned peers
				if m.pm.IsBanned(p) {
					continue
				}

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

	// Update inbound/outbound connection count
	go func(n net.Network, conn net.Conn) {
		curInboundConns, curOutboundConns := m.connsInfo.Info()

		if conn.Stat().Direction == net.DirInbound {
			m.connsInfo.IncInbound()
			if int64(curInboundConns) > m.pm.config.Node.MaxInboundConnections {
				m.log.Debug("Closed inbound connection. Max. limit reached",
					"MaxAllowed", m.pm.config.Node.MaxInboundConnections)
				conn.Close()
			}
		}

		if conn.Stat().Direction == net.DirOutbound {
			m.connsInfo.IncOutbound()
			if int64(curOutboundConns) > m.pm.config.Node.MaxOutboundConnections {
				m.log.Debug("Closed outbound connection. Max. limit reached",
					"MaxAllowed", m.pm.config.Node.MaxOutboundConnections)
				conn.Close()
			}
		}
	}(n, conn)

	// Reset connection failure count
	rnAddr := util.RemoteAddrFromConn(conn)
	m.pm.ClearConnFailCount(rnAddr)
}

// Disconnected is called when a connection is closed.
// Update the connection count and inform the peer
// manager of the disconnection event.
func (m *ConnectionManager) Disconnected(n net.Network, conn net.Conn) {

	if conn.Stat().Direction == net.DirInbound {
		m.connsInfo.DecrInbound()
	}

	if conn.Stat().Direction == net.DirOutbound {
		m.connsInfo.DecrOutbound()
	}

	addr := util.RemoteAddrFromConn(conn)
	m.pm.HasDisconnected(addr)
}

// OpenedStream is called when a stream is openned
func (m *ConnectionManager) OpenedStream(n net.Network, s net.Stream) {

	// Reset the stream if local node has stopped
	if m.pm.localNode.HasStopped() {
		s.Reset()
	}

	rAddr := util.RemoteAddrFromConn(s.Conn())

	// Reset connection failure count
	m.pm.ClearConnFailCount(rAddr)
}

// ClosedStream is called when a stream is openned
func (m *ConnectionManager) ClosedStream(nt net.Network, s net.Stream) {}
