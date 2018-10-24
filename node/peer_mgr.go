package node

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/ellcrys/elld/elldb"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/util/logger"

	"github.com/ellcrys/elld/config"

	"github.com/ellcrys/elld/util"
)

// Manager manages known peers connected to the local peer.
// It is responsible for initiating and managing peers
// according to the current protocol and engine rules.
type Manager struct {
	mtx              *sync.RWMutex           // general mutex
	cacheMtx         *sync.RWMutex           // Cache mutex
	localNode        *Node                   // local node
	peers            map[string]types.Engine // peers known to the peer manager
	log              logger.Logger           // manager's logger
	config           *config.EngineConfig    // manager's configuration
	connMgr          *ConnectionManager      // connection manager
	stop             bool                    // signifies the start of the manager
	timeBan          map[string]time.Time    // Stores the time where time banned peers are free
	acquainted       map[string]struct{}     // Store peers that sent and acknowledged handshake messages
	connectFailCount map[string]int          // Keeps count of connection attempt failure
	tickersDone      chan bool
}

// NewManager creates an instance of the peer manager
func NewManager(cfg *config.EngineConfig, localPeer *Node, log logger.Logger) *Manager {

	if cfg == nil {
		cfg = &config.EngineConfig{}
		cfg.Node = &config.PeerConfig{}
	}

	m := &Manager{
		mtx:              new(sync.RWMutex),
		cacheMtx:         new(sync.RWMutex),
		localNode:        localPeer,
		log:              log,
		peers:            make(map[string]types.Engine),
		config:           cfg,
		tickersDone:      make(chan bool),
		acquainted:       make(map[string]struct{}),
		timeBan:          make(map[string]time.Time),
		connectFailCount: make(map[string]int),
	}

	m.connMgr = NewConnMrg(m, log)
	m.localNode.host.Network().Notify(m.connMgr)
	return m
}

// TimeBanIndex get the time ban index
func (m *Manager) TimeBanIndex() map[string]time.Time {
	m.cacheMtx.RLock()
	defer m.cacheMtx.RUnlock()
	return m.timeBan
}

// AddTimeBan stores a time a peer is considered
// banned from outbound or inbound communication.
// If an existing entry exist for peer, add dur
// to it.
func (m *Manager) AddTimeBan(peer types.Engine, dur time.Duration) {
	m.cacheMtx.Lock()
	defer m.cacheMtx.Unlock()
	curBanTime := m.timeBan[peer.StringID()]

	// If the cur ban time of the peer is in the
	// past, set it to now before updating it with dur
	now := time.Now()
	if curBanTime.Before(now) {
		curBanTime = now
	}

	m.timeBan[peer.GetAddress().IP().String()] = curBanTime.Add(dur)
}

// GetBanTime gets the ban end time of peer
func (m *Manager) GetBanTime(peer types.Engine) time.Time {
	m.cacheMtx.RLock()
	defer m.cacheMtx.RUnlock()
	return m.timeBan[peer.GetAddress().IP().String()]
}

// IsBanned checks whether a peer has been banned.
func (m *Manager) IsBanned(peer types.Engine) bool {
	m.cacheMtx.RLock()
	defer m.cacheMtx.RUnlock()

	// Check if peer has been time banned
	curBanTime := m.timeBan[peer.GetAddress().IP().String()]
	if !curBanTime.IsZero() && curBanTime.After(time.Now()) {
		return true
	}

	return false
}

// IncrConnFailCount increases the connection failure count of n
func (m *Manager) IncrConnFailCount(n types.Engine) {
	m.cacheMtx.Lock()
	defer m.cacheMtx.Unlock()
	ip := n.GetAddress().IP().String()
	m.connectFailCount[ip]++
}

// ClearConnFailCount clears the connection failure count of n
func (m *Manager) ClearConnFailCount(n types.Engine) {
	m.cacheMtx.Lock()
	defer m.cacheMtx.Unlock()
	ip := n.GetAddress().IP().String()
	m.connectFailCount[ip] = 0
}

// GetConnFailCount returns the connection failure count of n
func (m *Manager) GetConnFailCount(n types.Engine) int {
	m.cacheMtx.RLock()
	defer m.cacheMtx.RUnlock()
	ip := n.GetAddress().IP().String()
	return m.connectFailCount[ip]
}

// PeerExist checks whether a peer is a known peer
func (m *Manager) PeerExist(peerID string) bool {
	m.mtx.RLock()
	defer m.mtx.RUnlock()
	_, exist := m.peers[peerID]
	return exist
}

// GetPeer returns a peer
func (m *Manager) GetPeer(peerID string) types.Engine {
	m.mtx.RLock()
	defer m.mtx.RUnlock()

	if !m.PeerExist(peerID) {
		return nil
	}

	return m.peers[peerID]
}

// AddAcquainted marks a peer has haven gone passed the
// handshake step
func (m *Manager) AddAcquainted(peer types.Engine) {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	m.acquainted[peer.StringID()] = struct{}{}
}

// IsAcquainted checks whether the peer passed through
// the handshake step
func (m *Manager) IsAcquainted(peer types.Engine) bool {
	m.mtx.RLock()
	defer m.mtx.RUnlock()
	_, has := m.acquainted[peer.StringID()]
	return has
}

// AddPeer adds a peer
func (m *Manager) AddPeer(peer types.Engine) {
	m.mtx.Lock()
	m.peers[peer.StringID()] = peer
	m.mtx.Unlock()
}

// ConnectToPeer attempts to establish
// a connection to a peer with the given id
func (m *Manager) ConnectToPeer(peerID string) error {
	peer := m.GetPeer(peerID)
	if peer == nil {
		return fmt.Errorf("peer not found")
	}
	return m.localNode.connectToNode(peer)
}

// GetUnconnectedPeers returns the peers that
// are currently not connected to the local peer.
func (m *Manager) GetUnconnectedPeers() (peers []types.Engine) {
	for _, p := range m.GetActivePeers(0) {
		if !p.Connected() {
			peers = append(peers, p)
		}
	}
	return
}

// GetConnectedPeers returns the connected peers
func (m *Manager) GetConnectedPeers() (peers []types.Engine) {
	for _, p := range m.GetActivePeers(0) {
		if p.Connected() {
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

	if err := m.LoadPeers(); err != nil {
		m.log.Error("failed to load peer addresses from database", "Err", err.Error())
	}

	go m.connMgr.Manage()
	go m.doSelfAdvert(m.tickersDone)
	go m.doCleanUp(m.tickersDone)
	go m.doPingMsgs(m.tickersDone)
	go m.doGetAddrMsg(m.tickersDone)
	go m.doIntro(m.tickersDone)
}

// doGetAddrMsg periodically sends wire.GetAddr
// message to all active peers
func (m *Manager) doGetAddrMsg(done chan bool) {
	ticker := time.NewTicker(time.Duration(m.config.Node.GetAddrInterval) * time.Second)
	for {
		select {
		case <-ticker.C:
			m.localNode.gProtoc.SendGetAddr(m.GetActivePeers(0))
		case <-done:
			ticker.Stop()
			return
		}
	}
}

// doPingMsgs periodically sends wire.Ping
// messages to all peers.
func (m *Manager) doPingMsgs(done chan bool) {
	ticker := time.NewTicker(time.Duration(m.config.Node.PingInterval) * time.Second)
	for {
		select {
		case <-ticker.C:
			m.localNode.gProtoc.SendPing(m.GetPeers())
		case <-done:
			ticker.Stop()
			return
		}
	}
}

// doSelfAdvert periodically send an wire.Addr
// message containing only the local peer's
// address to all connected peers.
func (m *Manager) doSelfAdvert(done chan bool) {
	ticker := time.NewTicker(time.Duration(m.config.Node.SelfAdvInterval) * time.Second)
	for {
		select {
		case <-ticker.C:
			peers := m.GetConnectedPeers()
			if len(peers) > 0 {
				m.localNode.gProtoc.SelfAdvertise(peers)
			}
			m.CleanPeers()
		case <-done:
			ticker.Stop()
			return
		}
	}
}

// doCleanUp periodically cleans the peer list,
// removing inactive peers.
func (m *Manager) doCleanUp(done chan bool) {
	ticker := time.NewTicker(time.Duration(m.config.Node.CleanUpInterval) * time.Second)
	for {
		select {
		case <-ticker.C:
			nCleaned := m.CleanPeers()
			m.log.Debug("Cleaned up old peers",
				"NumKnownPeers", len(m.GetPeers()),
				"NumPeersCleaned", nCleaned)
		case <-done:
			ticker.Stop()
			return
		}
	}
}

// doIntro periodically sends out wire.Intro messages
func (m *Manager) doIntro(done chan bool) {
	ticker := time.NewTicker(time.Duration(m.config.Node.SelfAdvInterval) * time.Second)
	for {
		select {
		case <-ticker.C:
			peers := m.GetConnectedPeers()
			if len(peers) > 0 {
				m.localNode.gProtoc.SendIntro(nil)
			}
		case <-done:
			ticker.Stop()
			return
		}
	}
}

// CanAcceptNode determines whether we can continue to
// interact with a given node.
func (m *Manager) CanAcceptNode(node *Node) (bool, error) {

	// Don't do this in test mode
	if m.localNode.TestMode() {
		return true, nil
	}

	// When the remote and local peer have not performed
	// the handshake ritual, other messages can't be accepted.
	if !m.IsAcquainted(node) {
		return false, fmt.Errorf("unacquainted node")
	}

	// When a remote peer is has an active ban time
	// period, we can receive messages from it.
	if m.IsBanned(node) {
		return false, fmt.Errorf("currently serving ban time")
	}

	return true, nil
}

// AddOrUpdateNode adds a peer to peer list if
// it hasn't been added. It updates the timestamp
// of existing peers.
func (m *Manager) AddOrUpdateNode(_peer types.Engine) {

	defer m.CleanPeers()
	defer m.SavePeers()

	peer := m.GetPeer(_peer.StringID())

	// For unknown peers, set 'last seen' time to an hour ago
	if peer == nil {
		_peer.SetLastSeen(time.Now().Add(-1 * time.Hour))
		m.AddPeer(_peer)
		return
	}

	// For connected peers, set 'last seen' time to the current time
	if peer.Connected() {
		peer.SetLastSeen(time.Now())
		return
	}

	// At this point, we know the peer but we are not
	// currently connected to it. To accelerate its removal,
	// deduct 1 hour from its current time
	peer.SetLastSeen(peer.GetLastSeen().Add(-1 * time.Hour))
}

// Peers returns the map of known peers
func (m *Manager) Peers() map[string]types.Engine {
	m.mtx.RLock()
	defer m.mtx.RUnlock()
	return m.peers
}

// SetPeers sets the known peers
func (m *Manager) SetPeers(d map[string]types.Engine) {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	m.peers = d
}

// hasReachedOutConnLimit checks whether the
// local peer has reached its outgoing
// connection limit
func (m *Manager) hasReachedOutConnLimit() bool {
	_, outbound := m.connMgr.GetConnsCount().Info()
	return int64(outbound) >= m.config.Node.MaxOutboundConnections
}

// RequirePeers checks whether we need more peers
func (m *Manager) RequirePeers() bool {
	return len(m.GetActivePeers(0)) < 1000 && !m.hasReachedOutConnLimit()
}

// IsLocalNode checks if a peer is the local peer
func (m *Manager) IsLocalNode(p types.Engine) bool {
	return p != nil && m.localNode != nil && m.localNode.IsSame(p)
}

// ConnMgr gets the connection manager
func (m *Manager) ConnMgr() *ConnectionManager {
	return m.connMgr
}

// SetLocalNode sets the local node
func (m *Manager) SetLocalNode(n *Node) {
	m.localNode = n
}

// IsActive returns true of a peer is considered active.
func (m *Manager) IsActive(p types.Engine) bool {
	return time.Now().Add(-3*(60*60)*time.Second).Before(p.GetLastSeen()) &&
		!m.IsBanned(p)
}

// HasDisconnected is called with a address belonging
// to a peer that had just disconnected. It will set
// the last seen time of the peer to an hour ago to
// to quicken the time to clean up. The peer may
// reconnect before clean up.
func (m *Manager) HasDisconnected(peerAddr util.NodeAddr) error {

	peer := m.GetPeer(peerAddr.StringID())
	if peer == nil {
		return fmt.Errorf("unknown peer")
	}

	m.log.Info("Peer has disconnected", "PeerID", peer.ShortID())

	peer.SetLastSeen(peer.GetLastSeen().Add(-1 * time.Hour))

	m.CleanPeers()

	return nil
}

// CleanPeers removes old peers from the list
// of peers known by the local peer. Typically,
// we remove peers based on their active status.
// It returns the number of peers removed
func (m *Manager) CleanPeers() int {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	before := len(m.peers)
	newKnownPeers := make(map[string]types.Engine)

	for k, p := range m.peers {

		// If last communication was received within 3 hours
		// ago, we can keep the peer
		if !m.IsBanned(p) && time.Now().Add(-3*(60*60)*time.Second).
			Before(p.GetLastSeen()) {
			newKnownPeers[k] = p
			continue
		}

		// If peer has been banned but have a ban time
		// that is less than <= 3 hours from now, we can keep the peer
		if m.IsBanned(p) && m.GetBanTime(p).Before(time.Now().Add(3*time.Hour)) {
			newKnownPeers[k] = p
			continue
		}

		delete(m.acquainted, k)
	}

	after := len(newKnownPeers)
	m.peers = newKnownPeers

	return before - after
}

// GetPeers gets all the known
// peers (connected or unconnected).
// Banned peers are exempted.
func (m *Manager) GetPeers() (peers []types.Engine) {
	m.mtx.RLock()
	defer m.mtx.RUnlock()

	for _, p := range m.peers {
		if !m.IsBanned(p) {
			peers = append(peers, p)
		}
	}

	return
}

// GetActivePeers returns active peers.
// Passing a zero or negative value
// as limit means no limit is applied.
func (m *Manager) GetActivePeers(limit int) (peers []types.Engine) {
	m.mtx.RLock()
	defer m.mtx.RUnlock()
	for _, p := range m.peers {
		if limit > 0 && len(peers) >= limit {
			return
		}
		if m.IsActive(p) {
			peers = append(peers, p)
		}
	}
	return
}

// CopyActivePeers is like GetActivePeers
// but a different slice is returned
func (m *Manager) CopyActivePeers(limit int) (peers []types.Engine) {
	m.mtx.RLock()
	defer m.mtx.RUnlock()

	activePeers := m.GetActivePeers(limit)
	copiedActivePeers := make([]types.Engine, len(activePeers))
	copy(copiedActivePeers, activePeers)
	return copiedActivePeers
}

// GetRandomActivePeers returns a slice
// of randomly selected peers whose
// timestamp is within 3 hours ago.
func (m *Manager) GetRandomActivePeers(limit int) []types.Engine {
	m.mtx.RLock()
	defer m.mtx.RUnlock()

	peers := m.CopyActivePeers(0)

	// shuffle known peer slice
	for i := range peers {
		j := rand.Intn(i + 1)
		peers[i], peers[j] = peers[j], peers[i]
	}

	if len(peers) <= limit {
		return peers
	}

	return peers[:limit]
}

// SavePeers stores active peer addresses
func (m *Manager) SavePeers() error {
	m.mtx.RLock()
	defer m.mtx.RUnlock()

	var numAddrs = 0
	var kvObjs []*elldb.KVObject

	// Hardcoded seed peers and peers that are
	// not up to 20 minutes old are also not saved.
	peers := m.CopyActivePeers(0)
	for _, p := range peers {
		isOldEnough := time.Now().Sub(p.CreatedAt()).Minutes() >= 20
		if !p.IsHardcodedSeed() && isOldEnough {
			key := []byte(p.StringID())
			value := util.ObjectToBytes(map[string]interface{}{
				"address":   p.GetAddress(),
				"createdAt": p.CreatedAt().Unix(),
				"lastSeen":  p.GetLastSeen().Unix(),
			})
			obj := elldb.NewKVObject(key, value, []byte("address"))
			kvObjs = append(kvObjs, obj)
			numAddrs++
		}
	}

	if err := m.localNode.db.Put(kvObjs); err != nil {
		return err
	}

	m.log.Debug("Saved peer addresses", "NumAddrs", numAddrs)

	return nil
}

// LoadPeers loads peers stored in
// the local database
func (m *Manager) LoadPeers() error {

	kvObjs := m.localNode.db.GetByPrefix([]byte("address"))

	// create remote node to represent
	// the addresses and add them to the
	// managers active peer list
	for _, o := range kvObjs {

		var addrData map[string]interface{}
		if err := o.Scan(&addrData); err != nil {
			return err
		}

		addr := util.NodeAddr(addrData["address"].(string))
		peer := NewRemoteNode(addr, m.localNode)
		peer.createdAt = time.Unix(int64(addrData["createdAt"].(uint32)), 0)
		peer.lastSeen = time.Unix(int64(addrData["lastSeen"].(uint32)), 0)
		m.AddPeer(peer)
	}

	return nil
}

// Stop gracefully stops running
// routines managed by the manager
func (m *Manager) Stop() {
	m.SavePeers()

	m.mtx.Lock()
	defer m.mtx.Unlock()
	if m.stop {
		return
	}

	if m.tickersDone != nil {
		close(m.tickersDone)
	}

	if m.connMgr.tickerDone != nil {
		close(m.connMgr.tickerDone)
	}

	m.stop = true
	m.log.Info("Peer manager has stopped")
}
