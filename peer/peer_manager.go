package peer

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/ellcrys/druid/util"
	ma "github.com/multiformats/go-multiaddr"
	"go.uber.org/zap"
)

// ManagerConfig represents the configuration for the manager
type ManagerConfig struct {
	GetAddrInterval int
}

// Manager manages known peers connected to the local peer.
// It is responsible for initiating the peer discovery process
// according to the current protocol
type Manager struct {
	kpm            *sync.Mutex        // known peer mutex
	localPeer      *Peer              // local peer
	bootstrapPeers map[string]*Peer   // bootstrap peers
	knownPeers     map[string]*Peer   // peers known to the peer manager
	log            *zap.SugaredLogger // manager's logger
	config         *ManagerConfig     // manager's configuration
	getAddrTicker  *time.Ticker       // ticker that sends "getaddr" messages
}

// NewManager creates an instance of the peer manager
func NewManager(localPeer *Peer) *Manager {

	defaultConfig := &ManagerConfig{
		GetAddrInterval: 10,
	}

	m := &Manager{
		kpm:            new(sync.Mutex),
		localPeer:      localPeer,
		log:            peerLog.Named("manager"),
		bootstrapPeers: make(map[string]*Peer),
		knownPeers:     make(map[string]*Peer),
		config:         defaultConfig,
	}

	notif := &Notification{}
	m.localPeer.host.Network().Notify(notif)

	return m
}

// AddBootstrapPeer adds a peer to the manager
func (m *Manager) AddBootstrapPeer(peer *Peer) {
	m.bootstrapPeers[peer.IDPretty()] = peer
}

// GetBootstrapPeers returns the bootstrap peers
func (m *Manager) GetBootstrapPeers() map[string]*Peer {
	return m.bootstrapPeers
}

// GetBootstrapPeer returns a peer in the boostrap peer list
func (m *Manager) GetBootstrapPeer(id string) *Peer {
	return m.bootstrapPeers[id]
}

// Manage starts managing peer connections.
func (m *Manager) Manage() {
	go m.sendPeriodicGetAddrMsg()
}

// sendPeriodicGetAddrMsg sends "getaddr" message to all known active
// peers as long as the number of known peers is less than 1000
func (m *Manager) sendPeriodicGetAddrMsg() {
	m.getAddrTicker = time.NewTicker(time.Duration(m.config.GetAddrInterval) * time.Second)
	for {
		select {
		case <-m.getAddrTicker.C:
			m.localPeer.protoc.DoGetAddr()
		}
	}
}

// AddOrUpdatePeer adds a peer to the list of known peers if it doesn't
// exist already. The new peer's timestamp is updated.
// TODO: clear old and inactive peers
func (m *Manager) AddOrUpdatePeer(p *Peer) error {
	if p == nil {
		return fmt.Errorf("nil received as *Peer")
	}
	m.kpm.Lock()
	defer m.kpm.Unlock()

	// update timestamp
	p.Timestamp = time.Now().UTC()

	// add peer if it does not exist
	if _, ok := m.knownPeers[p.IDPretty()]; !ok {
		m.knownPeers[p.IDPretty()] = p
	}

	return nil
}

// KnownPeers returns the map of known peers
func (m *Manager) KnownPeers() map[string]*Peer {
	return m.knownPeers
}

// NeedMorePeers checks whether we need more peers
func (m *Manager) NeedMorePeers() bool {
	return len(m.GetActivePeers(0)) < 1000
}

// IsLocalPeer checks if a peer is the local peer
func (m *Manager) IsLocalPeer(p *Peer) bool {
	return p.IDPretty() == m.localPeer.IDPretty()
}

// isActive returns true of a peer is considered active.
// First rule, its timestamp must be within the last 3 hours
func (m *Manager) isActive(p *Peer) bool {
	return time.Now().UTC().Add(-3 * (60 * 60) * time.Second).Before(p.Timestamp.UTC())
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

// GetRandomActivePeers returns a slice of randomly selected peers
// whose timestamp is within 3 hours ago.
// Returns error if number of known and active peers is less than limit
func (m *Manager) GetRandomActivePeers(limit int) ([]*Peer, error) {

	knownActivePeers := m.GetActivePeers(-1)
	m.kpm.Lock()
	defer m.kpm.Unlock()

	// shuffle known peer slice
	for i := range knownActivePeers {
		j := rand.Intn(i + 1)
		knownActivePeers[i], knownActivePeers[j] = knownActivePeers[j], knownActivePeers[i]
	}

	if len(knownActivePeers) <= limit {
		return knownActivePeers, nil
	}

	return knownActivePeers[:limit], nil
}

// PeerExist checks if a peer exists
func (m *Manager) PeerExist(peer *Peer) bool {
	if _, ok := m.knownPeers[peer.IDPretty()]; ok {
		return true
	}
	return false
}

// CreatePeerFromAddress creates a new peer and assign the multiaddr to it.
func (m *Manager) CreatePeerFromAddress(addr string) error {

	if !util.IsValidAddress(addr) {
		return fmt.Errorf("failed to create peer from address. Peer address is invalid")
	}

	mAddr, _ := ma.NewMultiaddr(addr)
	remotePeer := &Peer{address: mAddr}
	if m.PeerExist(remotePeer) {
		m.log.Infof("peer (%s) already exists", remotePeer.IDPretty())
		return nil
	}

	m.AddOrUpdatePeer(remotePeer)
	m.log.Infow("added a peer", "PeerAddr", mAddr.String())

	return nil
}

// Stop gracefully stops running routines managed by the manager
func (m *Manager) Stop() {
	m.getAddrTicker.Stop()
}
