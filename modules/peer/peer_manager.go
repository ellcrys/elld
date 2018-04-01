package peer

import (
	"fmt"
	"time"

	"github.com/ellcrys/gcoin/modules/util"
	ma "github.com/multiformats/go-multiaddr"
	"go.uber.org/zap"
)

var (
	// ManageBootstrapInterval is the time between each bootstrap manage operation
	ManageBootstrapInterval = 10 * time.Second
)

// Manager manages known peers connected to the local peer.
// It is responsible for initiating the peer discovery process
// according to the current protocol
type Manager struct {
	localPeer      *Peer
	bootstrapPeers []*Peer
	peers          map[string]*Peer
	log            *zap.SugaredLogger
}

// NewManager creates an instance of the peer manager
func NewManager(localPeer *Peer) *Manager {
	return &Manager{
		localPeer: localPeer,
		log:       peerLog.Named("manager"),
		peers:     make(map[string]*Peer),
	}
}

// AddBootstrapPeer adds a peer to the manager
func (m *Manager) AddBootstrapPeer(peer *Peer) {
	m.bootstrapPeers = append(m.bootstrapPeers, peer)
}

// startHandshakeWithBootstrapPeers sends a handshake to bootstrap peers
func (m Manager) startHandshakeWithBootstrapPeers() {
	for _, p := range m.bootstrapPeers {
		go SendHandshake(p)
	}
}

// Manage starts the periodic routine of managing peer connection.
// - Sends "op_handshake" to bootstrap nodes and receiving addr message as reply
func (m *Manager) Manage() {
	m.startHandshakeWithBootstrapPeers()
}

func (m *Manager) addPeer(p *Peer) {
	m.peers[p.IDPretty()] = p
}

func (m *Manager) isLocalPeer(p *Peer) bool {
	return p.IDPretty() == m.localPeer.IDPretty()
}

// getActivePeers returns some of the recently active peers
func (m *Manager) getActivePeers() []*Peer {
	var peers []*Peer
	for _, p := range m.peers {
		if !m.isLocalPeer(p) {
			peers = append(peers, p)
		}
	}
	return peers
}

// peerExist checks if a peer exists
func (m *Manager) peerExist(peer *Peer) bool {
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
	if m.peerExist(remotePeer) {
		m.log.Infof("peer (%s) already exists", remotePeer.IDPretty())
		return nil
	}

	m.addPeer(remotePeer)
	m.log.Infof("added a peer (%s)", mAddr.String())

	return nil
}
