package peer

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/ellcrys/druid/util/logger"

	"github.com/ellcrys/druid/configdir"

	"github.com/ellcrys/druid/util"
	ma "github.com/multiformats/go-multiaddr"
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
	log            logger.Logger      // manager's logger
	config         *configdir.Config  // manager's configuration
	connMgr        *ConnectionManager // connection manager
	getAddrTicker  *time.Ticker       // ticker that sends "getaddr" messages
	pingTicker     *time.Ticker       // ticker that sends "ping" messages
	selfAdvTicker  *time.Ticker       // ticker that sends "addr" message for self advertisement
	cleanUpTicker  *time.Ticker       // ticker that cleans up the peer
	stop           bool               // signifies the start of the manager
}

// NewManager creates an instance of the peer manager
func NewManager(cfg *configdir.Config, localPeer *Peer, log logger.Logger) *Manager {

	if cfg == nil {
		cfg = &configdir.Config{}
		cfg.Peer = &configdir.PeerConfig{}
	}

	if !cfg.Peer.Dev {
		cfg.Peer.GetAddrInterval = 30 * 60
		cfg.Peer.PingInterval = 30 * 60
		cfg.Peer.SelfAdvInterval = 24 * 60 * 60
		cfg.Peer.CleanUpInterval = 10 * 60
	}

	m := &Manager{
		kpm:            new(sync.Mutex),
		gm:             new(sync.Mutex),
		localPeer:      localPeer,
		log:            log,
		bootstrapPeers: make(map[string]*Peer),
		knownPeers:     make(map[string]*Peer),
		config:         cfg,
	}

	m.connMgr = NewConnMrg(m, log)
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

	if peer := m.GetKnownPeer(peerID); peer != nil {
		m.onFailedConnection(peer)
		m.log.Info("Peer has disconnected", "PeerID", peer.ShortID())
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
	go m.connMgr.Manage()
	go m.periodicSelfAdvertisement()
	go m.periodicCleanUp()
	go m.periodicPingMsgs()
	// go m.sendPeriodicGetAddrMsg()
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

// periodicPingMsgs sends "ping" messages to all peers
// as a basic health check routine.
func (m *Manager) periodicPingMsgs() {
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

// periodicSelfAdvertisement send an Addr message containing only the
// local peer address to all connected peers
func (m *Manager) periodicSelfAdvertisement() {
	m.selfAdvTicker = time.NewTicker(time.Duration(m.config.Peer.SelfAdvInterval) * time.Second)
	for {
		if m.stop {
			break
		}
		select {
		case <-m.selfAdvTicker.C:
			connectedPeers := []*Peer{}
			for _, p := range m.GetKnownPeers() {
				if p.Connected() {
					connectedPeers = append(connectedPeers, p)
				}
			}
			m.localPeer.protoc.SelfAdvertise(connectedPeers)
			m.CleanKnownPeers()
		}
	}
}

// periodicCleanUp performs peer clean up such as
// removing old know peers.
func (m *Manager) periodicCleanUp() {
	m.cleanUpTicker = time.NewTicker(time.Duration(m.config.Peer.CleanUpInterval) * time.Second)
	for {
		if m.stop {
			break
		}
		select {
		case <-m.cleanUpTicker.C:
			nCleaned := m.CleanKnownPeers()
			m.log.Debug("Cleaned up old peers", "NumKnownPeers", len(m.knownPeers), "NumPeersCleaned", nCleaned)
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
// - clean old addresses
func (m *Manager) AddOrUpdatePeer(p *Peer) error {

	defer m.CleanKnownPeers()

	if p == nil {
		return fmt.Errorf("nil received")
	}

	if p.IsSame(m.localPeer) {
		return fmt.Errorf("peer is the local peer")
	}

	if !util.IsValidAddr(p.GetMultiAddr()) {
		return fmt.Errorf("peer address is not valid")
	}

	if !m.localPeer.DevMode() && !util.IsRoutableAddr(p.GetMultiAddr()) {
		return fmt.Errorf("peer address is not routable")
	}

	if !m.config.Peer.Test { // don't do this in test environment (we will test savePeer alone)
		defer m.savePeers()
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

// onFailedConnection sets a new timestamp on a peer by deducting a fixed
// amount of time from its current timestamp.
// It will also call CleanKnowPeer. The purpose is to expedite the removal
// of disconnected
func (m *Manager) onFailedConnection(remotePeer *Peer) error {
	if remotePeer == nil {
		return fmt.Errorf("nil passed")
	}
	remotePeer.Timestamp = remotePeer.Timestamp.Add(-1 * time.Hour)
	m.CleanKnownPeers()
	return nil
}

// CleanKnownPeers removes old peers from the list
// of peers known by the local peer. Typically, we remove
// peers based on the last time they were seen. At least 3 connections
// must be active before we can clean.
// It returns the number of peers removed
// TODO: Also remove based on connection failure count?
func (m *Manager) CleanKnownPeers() int {

	if m.connMgr.connectionCount() < 3 {
		return 0
	}

	m.kpm.Lock()
	defer m.kpm.Unlock()

	before := len(m.knownPeers)

	newKnownPeers := make(map[string]*Peer)
	for k, p := range m.knownPeers {
		if m.isActive(p) {
			newKnownPeers[k] = p
		}
	}

	m.knownPeers = newKnownPeers

	return before - len(newKnownPeers)
}

// GetKnownPeers gets all the known peers (active or inactive)
func (m *Manager) GetKnownPeers() (peers []*Peer) {

	m.kpm.Lock()
	defer m.kpm.Unlock()

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

	if !m.localPeer.DevMode() && !util.IsRoutableAddr(addr) {
		return fmt.Errorf("failed to create peer from address. Peer address is invalid")
	}

	mAddr, _ := ma.NewMultiaddr(addr)
	remotePeer := NewRemotePeer(mAddr, m.localPeer)
	if m.PeerExist(remotePeer.StringID()) {
		m.log.Info("Peer already exists", "PeerID", remotePeer.StringID())
		return nil
	}

	remotePeer.Timestamp = time.Now()
	err = m.AddOrUpdatePeer(remotePeer)
	m.log.Info("Added a peer", "PeerAddr", mAddr.String())

	return err
}

// serializeActivePeers returns a json encoded list of active
// peers. This is needed to persist peer addresses along with other
// state information. Hardcoded peers are not included.
func (m *Manager) serializeActivePeers() ([][]byte, error) {

	peers := m.CopyActivePeers(0)
	serPeer := [][]byte{}

	for _, p := range peers {
		if !p.isHardcodedSeed {
			bs, _ := json.Marshal(map[string]interface{}{
				"addr": p.GetMultiAddr(),
				"ts":   p.Timestamp.Unix(),
			})
			serPeer = append(serPeer, bs)
		}
	}

	return serPeer, nil
}

// deserializePeers takes a slice of bytes which was created by
// serializeActivePeers and create a new remote peer
func (m *Manager) deserializePeers(serPeers [][]byte) ([]*Peer, error) {

	var peers = make([]*Peer, len(serPeers))

	for i, p := range serPeers {
		var data map[string]interface{}
		if err := json.Unmarshal(p, &data); err != nil {
			return nil, err
		}

		addr, _ := ma.NewMultiaddr(data["addr"].(string))
		peer := NewRemotePeer(addr, m.localPeer)
		peer.Timestamp = time.Unix(int64(data["ts"].(float64)), 0)
		peers[i] = peer
	}

	return peers, nil
}

// savePeers stores peer addresses to a persistent store
func (m *Manager) savePeers() error {

	serPeer, err := m.serializeActivePeers()
	if err != nil {
		m.log.Error("failed to serialize active addresses", "Err", err.Error(), "NumAddrs", len(serPeer))
		return fmt.Errorf("failed to serialize active addresses")
	}

	if err := m.localPeer.db.Address().ClearAll(); err != nil {
		m.log.Error("failed to clear persistent addresses", "Err", err.Error(), "NumAddrs", len(serPeer))
		return fmt.Errorf("failed to clear persistent addresses")
	}

	if err := m.localPeer.db.Address().SaveAll(serPeer); err != nil {
		m.log.Error("failed to save addresses to storage", "Err", err.Error(), "NumAddrs", len(serPeer))
		return fmt.Errorf("failed to clear persistent addresses")
	}

	m.log.Debug("Saved addresses", "NumAddrs", len(serPeer))
	return nil
}

// Stop gracefully stops running routines managed by the manager
func (m *Manager) Stop() {
	m.stop = true
	m.getAddrTicker.Stop()
	m.pingTicker.Stop()
}
