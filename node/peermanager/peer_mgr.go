package peermanager

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/ellcrys/elld/elldb"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/util/logger"

	"github.com/ellcrys/elld/config"

	"github.com/ellcrys/elld/util"
)

// Manager manages known peers connected to the local peer.
// It is responsible for initiating and managing peers
// according to the current protocol and engine rules.
type Manager struct {
	mtx              sync.RWMutex           // general mutex
	ptx              sync.RWMutex           // peer cache mutex
	cacheMtx         sync.RWMutex           // Cache mutex
	localNode        core.Engine            // local node
	peers            map[string]core.Engine // peers known to the peer manager
	log              logger.Logger          // manager's logger
	config           *config.EngineConfig   // manager's configuration
	connMgr          *ConnectionManager     // connection manager
	stop             bool                   // signifies the start of the manager
	timeBan          map[string]time.Time   // Stores the time where time banned peers are free
	acquainted       map[string]struct{}    // Store peers that sent and acknowledged handshake messages
	connectFailCount map[string]int         // Keeps count of connection attempt failure
	tickersDone      chan bool
}

// NewManager creates an instance of the peer manager
func NewManager(cfg *config.EngineConfig, localPeer core.Engine, log logger.Logger) *Manager {

	if cfg == nil {
		cfg = &config.EngineConfig{}
		cfg.Node = &config.NodeConfig{}
	}

	m := &Manager{
		mtx:              sync.RWMutex{},
		ptx:              sync.RWMutex{},
		cacheMtx:         sync.RWMutex{},
		localNode:        localPeer,
		log:              log,
		peers:            make(map[string]core.Engine),
		config:           cfg,
		tickersDone:      make(chan bool),
		acquainted:       make(map[string]struct{}),
		timeBan:          make(map[string]time.Time),
		connectFailCount: make(map[string]int),
	}

	m.connMgr = NewConnMrg(m, log)
	m.localNode.GetHost().Network().Notify(m.connMgr)
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
func (m *Manager) AddTimeBan(peer core.Engine, dur time.Duration) {

	// We can't ban hardcoded seeds
	if peer.IsHardcodedSeed() {
		return
	}

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
func (m *Manager) GetBanTime(peer core.Engine) time.Time {
	m.cacheMtx.RLock()
	defer m.cacheMtx.RUnlock()
	return m.timeBan[peer.GetAddress().IP().String()]
}

// IsBanned checks whether a peer has been banned.
func (m *Manager) IsBanned(peer core.Engine) bool {
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
func (m *Manager) IncrConnFailCount(nodeAddr util.NodeAddr) {
	m.cacheMtx.Lock()
	defer m.cacheMtx.Unlock()
	m.connectFailCount[nodeAddr.IP().String()]++
}

// ClearConnFailCount clears the connection failure count of n
func (m *Manager) ClearConnFailCount(nodeAddr util.NodeAddr) {
	m.cacheMtx.Lock()
	defer m.cacheMtx.Unlock()
	m.connectFailCount[nodeAddr.IP().String()] = 0
}

// GetConnFailCount returns the connection failure count of n
func (m *Manager) GetConnFailCount(nodeAddr util.NodeAddr) int {
	m.cacheMtx.RLock()
	defer m.cacheMtx.RUnlock()
	return m.connectFailCount[nodeAddr.IP().String()]
}

// PeerExist checks whether a peer is a known peer
func (m *Manager) PeerExist(peerID string) bool {
	m.ptx.RLock()
	defer m.ptx.RUnlock()
	_, exist := m.peers[peerID]
	return exist
}

// GetPeer returns a peer
func (m *Manager) GetPeer(peerID string) core.Engine {
	m.ptx.RLock()
	defer m.ptx.RUnlock()

	if !m.PeerExist(peerID) {
		return nil
	}

	return m.peers[peerID]
}

// AddAcquainted marks a peer has haven gone passed the
// handshake step
func (m *Manager) AddAcquainted(peer core.Engine) {
	m.cacheMtx.Lock()
	defer m.cacheMtx.Unlock()
	m.acquainted[peer.StringID()] = struct{}{}
}

// RemoveAcquainted makes a peer unacquainted
func (m *Manager) RemoveAcquainted(peer core.Engine) {
	m.cacheMtx.Lock()
	defer m.cacheMtx.Unlock()
	delete(m.acquainted, peer.StringID())
}

// IsAcquainted checks whether the peer passed through
// the handshake step
func (m *Manager) IsAcquainted(peer core.Engine) bool {
	m.cacheMtx.RLock()
	defer m.cacheMtx.RUnlock()
	_, has := m.acquainted[peer.StringID()]
	return has
}

// AddPeer adds a peer
func (m *Manager) AddPeer(peer core.Engine) {
	m.ptx.Lock()
	m.peers[peer.StringID()] = peer
	m.ptx.Unlock()
}

// LocalPeer returns the local peer
func (m *Manager) LocalPeer() core.Engine {
	return m.localNode
}

// ConnectToPeer attempts to establish
// a connection to a peer with the given id
func (m *Manager) ConnectToPeer(peerID string) error {

	peer := m.GetPeer(peerID)
	if peer == nil {
		return fmt.Errorf("peer not found")
	}

	m.log.Debug("Attempting to connect to peer",
		"PeerID", peer.ShortID())

	gsp := m.localNode.Gossip()
	err := gsp.SendHandshake(peer)
	if err != nil {
		return err
	}

	return gsp.SendGetAddr([]core.Engine{peer})
}

// ConnectToNode attempts to a Handshake message
// to a remote node. If successful, it sends a
// GetAddr message.
func (m *Manager) ConnectToNode(node core.Engine) error {
	gsp := m.localNode.Gossip()
	err := gsp.SendHandshake(node)
	if err != nil {
		return err
	}
	return gsp.SendGetAddr([]core.Engine{node})
}

// GetUnconnectedPeers returns the peers that
// are currently not connected to the local peer.
func (m *Manager) GetUnconnectedPeers() (peers []core.Engine) {
	for _, p := range m.GetPeers() {
		if !p.Connected() {
			peers = append(peers, p)
		}
	}
	return
}

// GetLonelyPeers returns the peers that
// are currently not connected or are connected
// unacquainted to the local peer.
func (m *Manager) GetLonelyPeers() (peers []core.Engine) {
	for _, p := range m.GetPeers() {
		if !p.Connected() || !m.IsAcquainted(p) {
			peers = append(peers, p)
		}
	}
	return
}

// GetConnectedPeers returns connected peers
func (m *Manager) GetConnectedPeers() (peers []core.Engine) {
	for _, p := range m.GetPeers() {
		if p.Connected() {
			peers = append(peers, p)
		}
	}
	return
}

// GetAcquaintedPeers returns connected and acquainted peers
func (m *Manager) GetAcquaintedPeers() (peers []core.Engine) {
	for _, p := range m.GetPeers() {
		if p.Connected() && m.IsAcquainted(p) {
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
}

// doGetAddrMsg periodically sends wire.GetAddr
// message to all active peers
func (m *Manager) doGetAddrMsg(done chan bool) {
	ticker := time.NewTicker(time.Duration(m.config.Node.GetAddrInterval) * time.Second)
	for {
		select {
		case <-ticker.C:

			if m.localNode.IsNetworkDisabled() {
				continue
			}

			m.localNode.Gossip().SendGetAddr(m.GetActivePeers(0))
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

			if m.localNode.IsNetworkDisabled() {
				continue
			}

			m.localNode.Gossip().SendPing(m.GetActivePeers(0))
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

			if m.localNode.IsNetworkDisabled() {
				continue
			}

			peers := m.GetConnectedPeers()
			if len(peers) > 0 {
				m.localNode.Gossip().SelfAdvertise(peers)
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
			m.SavePeers()
			m.log.Debug("Cleaned and saved peers", "NumKnownPeers", len(m.GetPeers()),
				"NumPeersCleaned", nCleaned)

		case <-done:
			ticker.Stop()
			return
		}
	}
}

// CanAcceptNode determines whether we can continue to
// interact with a given node.
func (m *Manager) CanAcceptNode(node core.Engine, opts ...bool) (bool, error) {

	// Don't do this in test mode
	if m.localNode.TestMode() {
		return true, nil
	}

	skipAcquaintanceCheck := len(opts) > 0 && opts[0] == true

	// When the remote and local peer have not performed
	// the handshake ritual, other messages can't be accepted.
	if !skipAcquaintanceCheck && !m.IsAcquainted(node) {
		return false, fmt.Errorf("unacquainted node")
	}

	// When a remote peer is has an active ban time
	// period which is over 3 hours, we cannot receive messages from it.
	if m.IsBanned(node) && m.GetBanTime(node).After(time.Now().Add(3*time.Hour)) {
		return false, fmt.Errorf("currently serving ban time")
	}

	return true, nil
}

// AddOrUpdateNode adds a peer to peer list if
// it hasn't been added. It updates the timestamp
// of existing peers.
func (m *Manager) AddOrUpdateNode(n core.Engine) {
	defer m.CleanPeers()
	defer m.SavePeers()

	peer := m.GetPeer(n.StringID())
	// For unknown peers, set 'last seen' time to an hour ago
	if peer == nil {
		n.SetLastSeen(time.Now().Add(-1 * time.Hour))
		m.AddPeer(n)
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
func (m *Manager) Peers() map[string]core.Engine {
	m.ptx.RLock()
	defer m.ptx.RUnlock()
	return m.peers
}

// SetPeers sets the known peers
func (m *Manager) SetPeers(d map[string]core.Engine) {
	m.ptx.Lock()
	m.peers = d
	m.ptx.Unlock()
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
func (m *Manager) IsLocalNode(p core.Engine) bool {
	return p != nil && m.localNode != nil && m.localNode.IsSame(p)
}

// ConnMgr gets the connection manager
func (m *Manager) ConnMgr() *ConnectionManager {
	return m.connMgr
}

// SetLocalNode sets the local node
func (m *Manager) SetLocalNode(n core.Engine) {
	m.localNode = n
}

// hasSeenRecently checks whether the peer has been
// seen within a given duration
func (m *Manager) hasSeenRecently(p core.Engine) bool {
	return time.Now().Add(-3 * (60 * 60) * time.Second).Before(p.GetLastSeen())
}

// IsActive returns true of a peer is considered active.
func (m *Manager) IsActive(p core.Engine) bool {

	// If not banned and last communication was received
	// within 3 hours ago, we consider the peer active
	if !m.IsBanned(p) && m.hasSeenRecently(p) {
		return true
	}

	return false
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

	m.log.Debug("Peer has disconnected", "PeerID", peer.ShortID())

	peer.SetLastSeen(peer.GetLastSeen().Add(-1 * time.Hour))

	m.CleanPeers()

	return nil
}

// CleanPeers removes old peers from the list
// of peers known by the local peer. Typically,
// we remove peers based on their active status.
// It returns the number of peers removed
func (m *Manager) CleanPeers() int {
	peers := m.GetPeers()
	before := len(peers)
	clean := map[string]core.Engine{}

	for _, p := range peers {

		// If last communication was received within 3 hours
		// ago, we consider the peer active
		if !m.IsBanned(p) && time.Now().Add(-3*(60*60)*time.Second).
			Before(p.GetLastSeen()) {
			clean[p.StringID()] = p
			continue
		}

		// If peer has been banned but have a ban time
		// that is <= 3 hours in the future, we can keep the peer
		if m.IsBanned(p) && m.GetBanTime(p).Before(time.Now().Add(3*time.Hour)) {
			clean[p.StringID()] = p
			continue
		}

		delete(m.acquainted, p.StringID())
	}

	after := len(clean)
	m.SetPeers(clean)

	return before - after
}

// GetPeers gets all the known
// peers (connected or unconnected).
func (m *Manager) GetPeers() (peers []core.Engine) {
	m.ptx.RLock()
	defer m.ptx.RUnlock()

	for _, p := range m.peers {
		peers = append(peers, p)
	}

	return
}

// GetActivePeers returns active peers.
// Passing a zero or negative value
// as limit means no limit is applied.
func (m *Manager) GetActivePeers(limit int) (peers []core.Engine) {
	m.ptx.RLock()
	defer m.ptx.RUnlock()
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
func (m *Manager) CopyActivePeers(limit int) (peers []core.Engine) {
	m.ptx.RLock()
	defer m.ptx.RUnlock()

	activePeers := m.GetActivePeers(limit)
	copiedActivePeers := make([]core.Engine, len(activePeers))
	copy(copiedActivePeers, activePeers)
	return copiedActivePeers
}

// GetRandomActivePeers returns a slice
// of randomly selected peers whose
// timestamp is within 3 hours ago.
func (m *Manager) GetRandomActivePeers(limit int) []core.Engine {
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

// ForgetPeers deletes peers in memory and on disk
func (m *Manager) ForgetPeers() {
	m.SetPeers(map[string]core.Engine{})
	m.localNode.DB().DeleteByPrefix([]byte("address"))
}

// SavePeers stores active peer addresses
func (m *Manager) SavePeers() error {

	var numAddrs = 0
	var kvObjs []*elldb.KVObject
	var peers = m.GetPeers()

	// Hardcoded seed peers and peers that are
	// not up to 20 minutes old are also not saved.
	for _, p := range peers {

		isOldEnough := time.Now().Sub(p.CreatedAt()).Minutes() >= 20
		if !isOldEnough || !m.hasSeenRecently(p) || p.IsHardcodedSeed() {
			continue
		}

		key := []byte(p.StringID())
		value := map[string]interface{}{
			"address":   p.GetAddress(),
			"createdAt": p.CreatedAt().Unix(),
			"lastSeen":  p.GetLastSeen().Unix(),
		}

		if banTime := m.GetBanTime(p); !banTime.IsZero() {
			value["banTime"] = banTime.Unix()
		}

		obj := elldb.NewKVObject(key, util.ObjectToBytes(value), []byte("address"))
		kvObjs = append(kvObjs, obj)
		numAddrs++
	}

	if err := m.localNode.DB().Put(kvObjs); err != nil {
		return err
	}

	return nil
}

// LoadPeers loads peers stored in
// the local database
func (m *Manager) LoadPeers() error {

	kvObjs := m.localNode.DB().GetByPrefix([]byte("address"))

	// create remote node to represent
	// the addresses and add them to the
	// managers active peer list
	for _, o := range kvObjs {

		var addrData map[string]interface{}
		if err := o.Scan(&addrData); err != nil {
			return err
		}

		addr := util.NodeAddr(addrData["address"].(string))
		peer := m.localNode.NewRemoteNode(addr)

		// Do not overwrite peer if it already exists
		// in the peer list. The peer might have been
		// added by a different process during bootstrap.
		if m.PeerExist(peer.StringID()) {
			continue
		}

		peer.SetCreatedAt(time.Unix(int64(addrData["createdAt"].(uint32)), 0))
		peer.SetLastSeen(time.Unix(int64(addrData["lastSeen"].(uint32)), 0))
		m.AddPeer(peer)

		if addrData["banTime"] != nil {
			banTime := time.Unix(int64(addrData["banTime"].(uint32)), 0)
			m.cacheMtx.Lock()
			m.timeBan[addr.IP().String()] = banTime
			m.cacheMtx.Unlock()
		}
	}

	return nil
}

// Stop gracefully stops running
// routines managed by the manager
func (m *Manager) Stop() {

	m.CleanPeers()
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
