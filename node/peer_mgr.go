package node

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/ellcrys/elld/elldb"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/util/logger"

	"github.com/ellcrys/elld/config"

	"github.com/ellcrys/elld/util"
	ma "github.com/multiformats/go-multiaddr"
)

// Manager manages known peers connected to the local peer.
// It is responsible for initiating and managing peers
// according to the current protocol and engine rules.
type Manager struct {
	knownPeerMtx   *sync.Mutex             // known peer mutex
	generalMtx     *sync.Mutex             // general mutex
	localNode      *Node                   // local node
	bootstrapNodes map[string]types.Engine // bootstrap peers
	knownPeers     map[string]types.Engine // peers known to the peer manager
	log            logger.Logger           // manager's logger
	config         *config.EngineConfig    // manager's configuration
	connMgr        *ConnectionManager      // connection manager
	getAddrTicker  *time.Ticker            // ticker that sends "getaddr" messages
	pingTicker     *time.Ticker            // ticker that sends "ping" messages
	selfAdvTicker  *time.Ticker            // ticker that sends "addr" message for self advertisement
	cleanUpTicker  *time.Ticker            // ticker that cleans up the peer
	stop           bool                    // signifies the start of the manager
}

// NewManager creates an instance of the peer manager
func NewManager(cfg *config.EngineConfig, localPeer *Node, log logger.Logger) *Manager {

	if cfg == nil {
		cfg = &config.EngineConfig{}
		cfg.Node = &config.PeerConfig{}
	}

	// Set hardcoded config in production mode
	if localPeer.ProdMode() {
		cfg.Node.GetAddrInterval = 30 * 60
		cfg.Node.PingInterval = 30 * 60
		cfg.Node.SelfAdvInterval = 24 * 60 * 60
		cfg.Node.CleanUpInterval = 10 * 60
	}

	m := &Manager{
		knownPeerMtx:   new(sync.Mutex),
		generalMtx:     new(sync.Mutex),
		localNode:      localPeer,
		log:            log,
		bootstrapNodes: make(map[string]types.Engine),
		knownPeers:     make(map[string]types.Engine),
		config:         cfg,
	}

	m.connMgr = NewConnMrg(m, log)
	m.localNode.host.Network().Notify(m.connMgr)
	return m
}

// PeerExist checks whether a peer is a known peer
func (m *Manager) PeerExist(peerID string) bool {
	m.knownPeerMtx.Lock()
	defer m.knownPeerMtx.Unlock()
	_, exist := m.knownPeers[peerID]
	return exist
}

// GetKnownPeer returns a known peer
func (m *Manager) GetKnownPeer(peerID string) types.Engine {
	if !m.PeerExist(peerID) {
		return nil
	}

	m.knownPeerMtx.Lock()
	defer m.knownPeerMtx.Unlock()
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
func (m *Manager) AddBootstrapPeer(peer *Node) {
	m.bootstrapNodes[peer.StringID()] = peer
}

// GetBootstrapNodes returns the bootstrap peers
func (m *Manager) GetBootstrapNodes() map[string]types.Engine {
	return m.bootstrapNodes
}

// GetBootstrapPeer returns a peer in the boostrap peer list
func (m *Manager) GetBootstrapPeer(id string) types.Engine {
	return m.bootstrapNodes[id]
}

// connectToPeer attempts to connect to a peer
func (m *Manager) connectToPeer(peerID string) error {
	peer := m.GetKnownPeer(peerID)
	if peer == nil {
		return fmt.Errorf("peer not found")
	}
	return m.localNode.connectToNode(peer)
}

// getUnconnectedPeers returns the peers that are not connected
// to the local peer. Hardcoded bootstrap peers are not included.
func (m *Manager) getUnconnectedPeers() (peers []types.Engine) {
	for _, p := range m.GetActivePeers(0) {
		if !p.IsHardcodedSeed() && !p.Connected() {
			peers = append(peers, p)
		}
	}
	return
}

// Manage starts managing peer connections.
// Load peers that were serialized and stored in database.
// Start connection manager
// Start periodic self advertisement to other peers
// Start periodic clean up of known peer list
// Start periodic ping messages to peers
func (m *Manager) Manage() {

	if err := m.loadPeers(); err != nil {
		m.log.Error("failed to load peer addresses from database", "Err", err.Error())
	}

	go m.connMgr.Manage()
	go m.periodicSelfAdvertisement()
	go m.periodicCleanUp()
	go m.periodicPingMsgs()
	// go m.sendPeriodicGetAddrMsg()
}

// sendPeriodicGetAddrMsg sends "getaddr" message to all known active
// peers as long as the number of known peers is less than 1000
func (m *Manager) sendPeriodicGetAddrMsg() {
	m.getAddrTicker = time.NewTicker(time.Duration(m.config.Node.GetAddrInterval) * time.Second)
	for {
		if m.stop {
			break
		}
		select {
		case <-m.getAddrTicker.C:
			m.localNode.gProtoc.SendGetAddr(m.GetActivePeers(0))
		}
	}
}

// periodicPingMsgs sends "ping" messages to all peers
// as a basic health check routine.
func (m *Manager) periodicPingMsgs() {
	m.pingTicker = time.NewTicker(time.Duration(m.config.Node.PingInterval) * time.Second)
	for {
		if m.stop {
			break
		}
		select {
		case <-m.pingTicker.C:
			m.localNode.gProtoc.SendPing(m.GetKnownPeers())
		}
	}
}

// periodicSelfAdvertisement send an Addr message containing only the
// local peer address to all connected peers
func (m *Manager) periodicSelfAdvertisement() {
	m.selfAdvTicker = time.NewTicker(time.Duration(m.config.Node.SelfAdvInterval) * time.Second)
	for {
		if m.stop {
			break
		}
		select {
		case <-m.selfAdvTicker.C:
			connectedPeers := []types.Engine{}
			for _, p := range m.GetKnownPeers() {
				if p.Connected() {
					connectedPeers = append(connectedPeers, p)
				}
			}
			m.localNode.gProtoc.SelfAdvertise(connectedPeers)
			m.CleanKnownPeers()
		}
	}
}

// periodicCleanUp performs peer clean up such as
// removing old know peers.
func (m *Manager) periodicCleanUp() {
	m.cleanUpTicker = time.NewTicker(time.Duration(m.config.Node.CleanUpInterval) * time.Second)
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
func (m *Manager) AddOrUpdatePeer(p types.Engine) error {

	defer m.CleanKnownPeers()

	if p == nil {
		return fmt.Errorf("nil received")
	}

	// Peer address must not be same as the local node
	if p.IsSame(m.localNode) {
		return fmt.Errorf("peer is the local peer")
	}

	// It must have a valid address
	if !util.IsValidAddr(p.GetMultiAddr()) {
		return fmt.Errorf("peer address is not valid")
	}

	// In production mode, it must be a routable address
	if m.localNode.ProdMode() && !util.IsRoutableAddr(p.GetMultiAddr()) {
		return fmt.Errorf("peer address is not routable")
	}

	// Save the known peers.
	// don't do this in test environment (we will test savePeer alone)
	if !m.localNode.TestMode() {
		defer m.savePeers()
	}

	m.knownPeerMtx.Lock()
	defer m.knownPeerMtx.Unlock()

	// update the timestamp only if not already
	// set by caller or elsewhere
	if p.GetTimestamp().IsZero() {
		p.SetTimestamp(time.Now())
	}

	// get a peer matching the ID from the list of known peers.
	// if it does not exist, we add it immediately
	existingPeer, exist := m.knownPeers[p.StringID()]
	if !exist {
		m.knownPeers[p.StringID()] = p
		return nil
	}

	// if a peer exists, return error if the peer's
	// full address matches the candidate peer
	if existingPeer.GetMultiAddr() != p.GetMultiAddr() {
		return fmt.Errorf("existing peer address do not match")
	}

	// If the candidate peer's timestamp is within
	// the last 24 hours and the existing/matching peer we already know
	// has a timestamp within the last hour, we set the existing peer's
	// timestamp to an hour ago.
	now := time.Now()
	if now.Add(-24*time.Hour).Before(p.GetTimestamp()) && now.Add(-60*time.Minute).Before(existingPeer.GetTimestamp()) {
		existingPeer.SetTimestamp(now.Add(-60 * time.Minute))
		return nil
	}

	// If the candidate peer's timestamp is not within
	// the last 24 hours and the existing/matching peer we already know
	// has a timestamp also not within the last hour, we set the existing peer's
	// timestamp to 24 hours ago.
	if !now.Add(-24*time.Hour).Before(p.GetTimestamp()) && !now.Add(-24*time.Hour).Before(existingPeer.GetTimestamp()) {
		existingPeer.SetTimestamp(now.Add(-24 * time.Hour))
		return nil
	}

	// At this point, we simple update the existing peer's
	// timestamp with the candidate's peer timestamp
	existingPeer.SetTimestamp(p.GetTimestamp())

	return nil
}

// KnownPeers returns the map of known peers
func (m *Manager) KnownPeers() map[string]types.Engine {
	return m.knownPeers
}

// NeedMorePeers checks whether we need more peers
func (m *Manager) NeedMorePeers() bool {
	return len(m.GetActivePeers(0)) < 1000 && m.connMgr.needMoreConnections()
}

// IsLocalNode checks if a peer is the local peer
func (m *Manager) IsLocalNode(p types.Engine) bool {
	return p != nil && m.localNode != nil && p.StringID() == m.localNode.StringID()
}

// isActive returns true of a peer is considered active.
// First rule, its timestamp must be within the last 3 hours
func (m *Manager) isActive(p types.Engine) bool {
	return time.Now().Add(-3 * (60 * 60) * time.Second).Before(p.GetTimestamp())
}

// onFailedConnection sets a new timestamp on a peer by deducting a fixed
// amount of time from its current timestamp.
// It will also call CleanKnowPeer. The purpose is to expedite the removal
// of disconnected
func (m *Manager) onFailedConnection(remotePeer types.Engine) error {
	if remotePeer == nil {
		return fmt.Errorf("nil passed")
	}
	remotePeer.SetTimestamp(remotePeer.GetTimestamp().Add(-1 * time.Hour))
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

	m.knownPeerMtx.Lock()
	defer m.knownPeerMtx.Unlock()

	before := len(m.knownPeers)

	newKnownPeers := make(map[string]types.Engine)
	for k, p := range m.knownPeers {
		if m.isActive(p) {
			newKnownPeers[k] = p
		}
	}

	m.knownPeers = newKnownPeers

	return before - len(newKnownPeers)
}

// GetKnownPeers gets all the known peers (active or inactive)
func (m *Manager) GetKnownPeers() (peers []types.Engine) {
	m.knownPeerMtx.Lock()
	defer m.knownPeerMtx.Unlock()

	for _, p := range m.knownPeers {
		peers = append(peers, p)
	}

	return
}

// GetActivePeers returns active peers. Passing a zero or negative value
// as limit means no limit is applied.
func (m *Manager) GetActivePeers(limit int) (peers []types.Engine) {
	m.knownPeerMtx.Lock()
	defer m.knownPeerMtx.Unlock()
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
func (m *Manager) CopyActivePeers(limit int) (peers []types.Engine) {
	activePeers := m.GetActivePeers(limit)
	copiedActivePeers := make([]types.Engine, len(activePeers))
	copy(copiedActivePeers, activePeers)
	return copiedActivePeers
}

// GetRandomActivePeers returns a slice of randomly selected peers
// whose timestamp is within 3 hours ago.
func (m *Manager) GetRandomActivePeers(limit int) []types.Engine {

	knownActivePeers := m.CopyActivePeers(0)
	m.knownPeerMtx.Lock()
	defer m.knownPeerMtx.Unlock()

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

	if err = validateAddress(m.localNode, addr); err != nil {
		return err
	}

	// The peer must not already exists be known
	mAddr, _ := ma.NewMultiaddr(addr)
	remotePeer := NewRemoteNode(mAddr, m.localNode)
	if m.PeerExist(remotePeer.StringID()) {
		m.log.Info("Peer already exists", "PeerID", remotePeer.StringID())
		return nil
	}

	remotePeer.Timestamp = time.Now()
	err = m.AddOrUpdatePeer(remotePeer)
	m.log.Info("Added a peer", "PeerAddr", mAddr.String())

	return err
}

// deserializePeers takes a slice of bytes which was created by
// serializeActivePeers and creates new remote node
func (m *Manager) deserializePeers(serPeers [][]byte) ([]*Node, error) {

	var peers = make([]*Node, len(serPeers))

	for i, p := range serPeers {
		var data []interface{}
		if err := json.Unmarshal(p, &data); err != nil {
			return nil, err
		}

		addr, _ := ma.NewMultiaddr(data[0].(string))
		peer := NewRemoteNode(addr, m.localNode)
		peer.Timestamp = time.Unix(int64(data[1].(float64)), 0)
		peers[i] = peer
	}

	return peers, nil
}

// savePeers stores active peer addresses
func (m *Manager) savePeers() error {

	var numAddrs = 0
	var kvObjs []*elldb.KVObject

	// Determine the active addresses that are eligible.
	// Hardcoded seed peers are no eligible.
	// Peers that are not up to 20 minutes old are also not
	// eligible
	peers := m.CopyActivePeers(0)
	for _, p := range peers {
		if !p.IsHardcodedSeed() && time.Now().Add(20*time.Minute).Before(p.GetTimestamp()) {
			key := []byte(p.StringID())
			value := util.ObjectToBytes(map[string]interface{}{
				"addr": p.GetMultiAddr(),
				"ts":   p.GetTimestamp().Unix(),
			})
			kvObjs = append(kvObjs, elldb.NewKVObject(key, value, []byte("address")))
			numAddrs++
		}
	}

	if err := m.localNode.db.Put(kvObjs); err != nil {
		return err
	}

	m.log.Debug("Saved peer addresses", "NumAddrs", numAddrs)

	return nil
}

// LoadPeers loads peers stored in the local database
func (m *Manager) loadPeers() error {

	kvObjs := m.localNode.db.GetByPrefix([]byte("address"))

	// create remote node to represent the addresses
	// and add them to the managers active peer list
	for _, o := range kvObjs {

		var addrData map[string]interface{}
		if err := o.Scan(&addrData); err != nil {
			return err
		}

		addr, _ := ma.NewMultiaddr(addrData["addr"].(string))
		peer := NewRemoteNode(addr, m.localNode)
		peer.Timestamp = time.Unix(int64(addrData["ts"].(uint32)), 0)
		m.AddOrUpdatePeer(peer)
	}

	return nil
}

// Stop gracefully stops running routines managed by the manager
func (m *Manager) Stop() {
	m.stop = true

	if m.getAddrTicker != nil {
		m.getAddrTicker.Stop()
	}

	if m.pingTicker != nil {
		m.pingTicker.Stop()
	}

	m.log.Info("Peer manager has stopped")
}
