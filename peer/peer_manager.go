package peer

import (
	"fmt"
	"sync"

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
	peers          map[string]*Peer
	log            *zap.SugaredLogger
}

// NewManager creates an instance of the peer manager
func NewManager(localPeer *Peer) *Manager {
	m := &Manager{
		Mutex:          new(sync.Mutex),
		localPeer:      localPeer,
		log:            peerLog.Named("manager"),
		bootstrapPeers: make(map[string]*Peer),
		peers:          make(map[string]*Peer),
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
	m.peers[p.IDPretty()] = p
	return nil
}

// Peers returns the peers
func (m *Manager) Peers() map[string]*Peer {
	return m.peers
}

// IsLocalPeer checks if a peer is the local peer
func (m *Manager) IsLocalPeer(p *Peer) bool {
	return p.IDPretty() == m.localPeer.IDPretty()
}

// ActivePeers returns some of the recently active peers
func (m *Manager) ActivePeers() []*Peer {
	m.Lock()
	defer m.Unlock()
	var peers []*Peer
	for _, p := range m.peers {
		if !m.IsLocalPeer(p) {
			peers = append(peers, p)
		}
	}
	return peers
}

// PeerExist checks if a peer exists
func (m *Manager) PeerExist(peer *Peer) bool {
	if _, ok := m.peers[peer.IDPretty()]; ok {
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
