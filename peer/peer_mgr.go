package peer

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/ellcrys/druid/configdir"

	"github.com/ellcrys/druid/util"
	ma "github.com/multiformats/go-multiaddr"
	"go.uber.org/zap"
)

// Manager manages known peers connected to the local peer.
// It is responsible for initiating the peer discovery process
// according to the current protocol
type Manager struct {
	kpm            *sync.Mutex        // known peer mutex
	gm             *sync.Mutex        // general mutex
	localPeer      *Peer              // local peer
	bootstrapPeers map[string]*Peer   // bootstrap peers
	knownPeers     map[string]*Peer   // peers known to the peer manager
	log            *zap.SugaredLogger // manager's logger
	config         *configdir.Config  // manager's configuration
	connMgr        *ConnectionManager // connection manager
	getAddrTicker  *time.Ticker       // ticker that sends "getaddr" messages
	pingTicker     *time.Ticker       // ticker that sends "ping" messages
	stop           bool               // signifies the start of the manager
}

// NewManager creates an instance of the peer manager
func NewManager(cfg *configdir.Config, localPeer *Peer, log *zap.SugaredLogger) *Manager {

	if cfg == nil {
		cfg = &configdir.Config{}
		cfg.Peer = &configdir.PeerConfig{}
	}

	cfg.Peer.GetAddrInterval = 10
	cfg.Peer.PingInterval = 60

	m := &Manager{
		kpm:            new(sync.Mutex),
		gm:             new(sync.Mutex),
		localPeer:      localPeer,
		log:            log,
		bootstrapPeers: make(map[string]*Peer),
		knownPeers:     make(map[string]*Peer),
		config:         cfg,
	}

	m.connMgr = NewConnMrg(m, log.Named("conn_manager"))
	m.localPeer.host.Network().Notify(m.connMgr)

	return m
}

// PeerExist checks whether a peer is a known peer
func (m *Manager) PeerExist(peerID string) bool {
	m.kpm.Lock()
	defer m.kpm.Unlock()
	_, exist := m.knownPeers[peerID]
	return exist
}

// GetKnownPeer returns a known peer
func (m *Manager) GetKnownPeer(peerID string) *Peer {
	if !m.PeerExist(peerID) {
		return nil
	}

	m.kpm.Lock()
	defer m.kpm.Unlock()
	peer, _ := m.knownPeers[peerID]
	return peer
}

// onPeerDisconnect is called when peer disconnects.
// Decrement active peer count but do not remove from the known peer list
// because the peer might come back in a short time. Subtract 2 hours from
// its current timestamp. Eventually, it will be removed if it does not reconnect.
func (m *Manager) onPeerDisconnect(peerAddr ma.Multiaddr) {

	peerID := util.IDFromAddr(peerAddr).Pretty()
	if m.PeerExist(peerID) {
		peer := m.GetKnownPeer(peerID)
		peer.Timestamp = peer.Timestamp.Add(-2 * time.Hour)
		m.log.Infow("Peer has disconnected", "PeerID", peer.ShortID())
	}

	m.CleanKnownPeers()
}

// AddBootstrapPeer adds a peer to the manager
func (m *Manager) AddBootstrapPeer(peer *Peer) {
	m.bootstrapPeers[peer.StringID()] = peer
}

// GetBootstrapPeers returns the bootstrap peers
func (m *Manager) GetBootstrapPeers() map[string]*Peer {
	return m.bootstrapPeers
}

// GetBootstrapPeer returns a peer in the boostrap peer list
func (m *Manager) GetBootstrapPeer(id string) *Peer {
	return m.bootstrapPeers[id]
}

// connectToPeer attempts to connect to a peer
func (m *Manager) connectToPeer(peerID string) error {
	peer := m.GetKnownPeer(peerID)
	if peer == nil {
		return fmt.Errorf("peer not found")
	}
	return m.localPeer.connectToPeer(peer)
}

// getUnconnectedPeers returns the peers that are not connected
// to the local peer. Hardcoded bootstrap peers are not included.
func (m *Manager) getUnconnectedPeers() (peers []*Peer) {
	for _, p := range m.GetActivePeers(0) {
		if !p.isHardcodedSeed && !p.Connected() {
			peers = append(peers, p)
		}
	}
	return
}

// Manage starts managing peer connections.
func (m *Manager) Manage() {
	m.connMgr.Manage()
	// go m.sendPeriodicGetAddrMsg()
	// go m.sendPeriodicPingMsgs()
}

// sendPeriodicGetAddrMsg sends "getaddr" message to all known active
// peers as long as the number of known peers is less than 1000
func (m *Manager) sendPeriodicGetAddrMsg() {
	m.getAddrTicker = time.NewTicker(time.Duration(m.config.Peer.GetAddrInterval) * time.Second)
	for {
		if m.stop {
			break
		}
		select {
		case <-m.getAddrTicker.C:
			m.localPeer.protoc.SendGetAddr(m.GetActivePeers(0))
		}
	}
}

// sendPeriodicPingMsgs sends "ping" messages to all peers
// as a basic health check routine.
func (m *Manager) sendPeriodicPingMsgs() {
	m.pingTicker = time.NewTicker(time.Duration(m.config.Peer.PingInterval) * time.Second)
	for {
		if m.stop {
			break
		}
		select {
		case <-m.pingTicker.C:
			m.localPeer.protoc.SendPing(m.GetKnownPeers())
		}
	}
}

// AddOrUpdatePeer adds a peer to the list of known peers if it doesn't
// exist. If the peer already exists:
// - if the peer has been seen in the last 24 hours and its current
// 	 timestamp is over 60 minutes old, then update the timestamp to 60 minutes ago.
// - else if the peer has not been seen in the last 24 hours and its current timestamp is
//	 over 24 hours, then update the timestamp to 24 hours ago.
// - else use whatever timestamp is returned
func (m *Manager) AddOrUpdatePeer(p *Peer) error {

	if p == nil {
		return fmt.Errorf("nil received as *Peer")
	}

	if p.IsSame(m.localPeer) {
		return fmt.Errorf("peer is local peer")
	}

	if !util.IsValidAddr(p.GetMultiAddr()) {
		return fmt.Errorf("peer address is not valid")
	}

	if !m.config.Peer.Dev && !util.IsRoutableAddr(p.GetMultiAddr()) {
		return fmt.Errorf("peer address is not routable")
	}

	m.kpm.Lock()
	defer m.kpm.Unlock()

	// set timestamp only if not set by caller or elsewhere
	if p.Timestamp.IsZero() {
		p.Timestamp = time.Now()
	}

	existingPeer, exist := m.knownPeers[p.StringID()]
	if !exist {
		m.knownPeers[p.StringID()] = p
		return nil
	}

	if existingPeer.GetMultiAddr() != p.GetMultiAddr() {
		return fmt.Errorf("existing peer address do not match")
	}

	now := time.Now()
	if now.Add(-24*time.Hour).Before(p.Timestamp) && now.Add(-60*time.Minute).Before(existingPeer.Timestamp) {
		existingPeer.Timestamp = now.Add(-60 * time.Minute)
		return nil
	}

	if !now.Add(-24*time.Hour).Before(p.Timestamp) && !now.Add(-24*time.Hour).Before(existingPeer.Timestamp) {
		existingPeer.Timestamp = now.Add(-24 * time.Hour)
		return nil
	}

	existingPeer.Timestamp = p.Timestamp
	return nil
}

// KnownPeers returns the map of known peers
func (m *Manager) KnownPeers() map[string]*Peer {
	return m.knownPeers
}

// NeedMorePeers checks whether we need more peers
func (m *Manager) NeedMorePeers() bool {
	return len(m.GetActivePeers(0)) < 1000 && m.connMgr.needMoreConnections()
}

// IsLocalPeer checks if a peer is the local peer
func (m *Manager) IsLocalPeer(p *Peer) bool {
	return p != nil && m.localPeer != nil && p.StringID() == m.localPeer.StringID()
}

// isActive returns true of a peer is considered active.
// First rule, its timestamp must be within the last 3 hours
func (m *Manager) isActive(p *Peer) bool {
	return time.Now().Add(-3 * (60 * 60) * time.Second).Before(p.Timestamp)
}

// TimestampPunishment sets a new timestamp on a peer by deducting a fixed
// amount of time from the current timestamp and resigning the new value.
// It will also call CleanKnowPeer. The purpose is to gradually, remove
// old, disconnected peers.
func (m *Manager) TimestampPunishment(remotePeer *Peer) error {
	if remotePeer == nil {
		return fmt.Errorf("nil passed")
	}
	remotePeer.Timestamp = remotePeer.Timestamp.Add(-1 * time.Hour)
	m.CleanKnownPeers()
	return nil
}

// CleanKnownPeers removes old peers from the known peers
func (m *Manager) CleanKnownPeers() {
	activePeers := m.GetActivePeers(0)

	m.kpm.Lock()
	defer m.kpm.Unlock()

	newKnownPeers := make(map[string]*Peer)
	for _, p := range activePeers {
		newKnownPeers[p.StringID()] = p
	}

	m.knownPeers = newKnownPeers
}

// GetKnownPeers gets all the known peers (active or inactive)
func (m *Manager) GetKnownPeers() (peers []*Peer) {
	for _, p := range m.knownPeers {
		peers = append(peers, p)
	}
	return peers
}

// GetActivePeers returns active peers. Passing a zero or negative value
// as limit means no limit is applied.
func (m *Manager) GetActivePeers(limit int) (peers []*Peer) {
	m.kpm.Lock()
	defer m.kpm.Unlock()
	for _, p := range m.knownPeers {
		if limit > 0 && len(peers) >= limit {
			return
		}
		if m.isActive(p) {
			peers = append(peers, p)
		}
	}
	return
}

// CopyActivePeers is like GetActivePeers but a different slice is returned
func (m *Manager) CopyActivePeers(limit int) (peers []*Peer) {
	activePeers := m.GetActivePeers(limit)
	copiedActivePeers := make([]*Peer, len(activePeers))
	copy(copiedActivePeers, activePeers)
	return copiedActivePeers
}

// GetRandomActivePeers returns a slice of randomly selected peers
// whose timestamp is within 3 hours ago.
func (m *Manager) GetRandomActivePeers(limit int) []*Peer {

	knownActivePeers := m.CopyActivePeers(0)
	m.kpm.Lock()
	defer m.kpm.Unlock()

	// shuffle known peer slice
	for i := range knownActivePeers {
		j := rand.Intn(i + 1)
		knownActivePeers[i], knownActivePeers[j] = knownActivePeers[j], knownActivePeers[i]
	}

	if len(knownActivePeers) <= limit {
		return knownActivePeers
	}

	return knownActivePeers[:limit]
}

// CreatePeerFromAddress creates a new peer and assign the multiaddr to it.
func (m *Manager) CreatePeerFromAddress(addr string) error {

	var err error

	if !util.IsValidAddr(addr) {
		return fmt.Errorf("failed to create peer from address. Peer address is invalid")
	}

	if !m.config.Peer.Dev && !util.IsRoutableAddr(addr) {
		return fmt.Errorf("failed to create peer from address. Peer address is invalid")
	}

	mAddr, _ := ma.NewMultiaddr(addr)
	remotePeer := NewRemotePeer(mAddr, m.localPeer)
	if m.PeerExist(remotePeer.StringID()) {
		m.log.Infof("peer (%s) already exists", remotePeer.StringID())
		return nil
	}

	remotePeer.Timestamp = time.Now()
	err = m.AddOrUpdatePeer(remotePeer)
	m.log.Infow("added a peer", "PeerAddr", mAddr.String())

	return err
}

// Stop gracefully stops running routines managed by the manager
func (m *Manager) Stop() {
	m.stop = true
	m.getAddrTicker.Stop()
	m.pingTicker.Stop()
}
