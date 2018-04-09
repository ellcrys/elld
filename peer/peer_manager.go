package peer

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/ellcrys/gcoin/util"
	ma "github.com/multiformats/go-multiaddr"
	"go.uber.org/zap"
)

// Manager manages known peers connected to the local peer.
// It is responsible for initiating the peer discovery process
// according to the current protocol
type Manager struct {
	*sync.Mutex
	localPeer      *Peer
	bootstrapPeers map[string]*Peer
	knownPeers     map[string]*Peer
	log            *zap.SugaredLogger
}

// NewManager creates an instance of the peer manager
func NewManager(localPeer *Peer) *Manager {
	m := &Manager{
		Mutex:          new(sync.Mutex),
		localPeer:      localPeer,
		log:            peerLog.Named("manager"),
		bootstrapPeers: make(map[string]*Peer),
		knownPeers:     make(map[string]*Peer),
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

}

// AddPeer adds a peer to the list of known peers
func (m *Manager) AddPeer(p *Peer) error {
	if p == nil {
		return fmt.Errorf("nil received as *Peer")
	}
	m.Lock()
	defer m.Unlock()
	m.knownPeers[p.IDPretty()] = p
	return nil
}

// KnownPeers returns the map of known peers
func (m *Manager) KnownPeers() map[string]*Peer {
	return m.knownPeers
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

// getActivePeers returns all active remote peers
func (m *Manager) getActivePeers() (peers []*Peer) {
	m.Lock()
	defer m.Unlock()
	for _, p := range m.knownPeers {
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

	knownActivePeers := m.getActivePeers()

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

// ActivePeers returns some of the recently active peers
func (m *Manager) ActivePeers(limit int) []*Peer {
	m.Lock()
	defer m.Unlock()
	var peers []*Peer
	for _, p := range m.knownPeers {
		if !m.IsLocalPeer(p) {
			peers = append(peers, p)
		}
	}
	return peers
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

	m.AddPeer(remotePeer)
	m.log.Infow("added a peer", "PeerAddr", mAddr.String())

	return nil
}
