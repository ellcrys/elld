package peer

import (
	"context"
	"fmt"
	mrand "math/rand"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/ellcrys/druid/util/logger"

	"github.com/ellcrys/druid/configdir"

	"github.com/thoas/go-funk"

	crypto "github.com/libp2p/go-libp2p-crypto"
	peer "github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"

	"github.com/ellcrys/druid/util"
	libp2p "github.com/libp2p/go-libp2p"
	host "github.com/libp2p/go-libp2p-host"
	inet "github.com/libp2p/go-libp2p-net"
	protocol "github.com/libp2p/go-libp2p-protocol"
	ma "github.com/multiformats/go-multiaddr"
)

// Peer represents a network node
type Peer struct {
	cfg             *configdir.Config // peer config
	address         ma.Multiaddr      // peer multiaddr
	IP              net.IP            // peer ip
	host            host.Host         // peer libp2p host
	wg              sync.WaitGroup    // wait group for preventing the main thread from exiting
	localPeer       *Peer             // local peer
	peerManager     *Manager          // peer manager for managing connections to other remote peers
	protoc          Protocol          // protocol instance
	remote          bool              // remote indicates the peer represents a remote peer
	Timestamp       time.Time         // the last time this peer was seen/active
	isHardcodedSeed bool              // whether the peer was hardcoded as a seed
	log             logger.Logger     // peer logger
	rSeed           []byte            // random 256 bit seed to be used for seed random operations
}

// NewPeer creates a peer instance at the specified port
func NewPeer(config *configdir.Config, address string, idSeed int64, log logger.Logger) (*Peer, error) {

	// generate peer identity
	priv, _, err := util.GenerateKeyPair(mrand.New(mrand.NewSource(idSeed)))
	if err != nil {
		return nil, fmt.Errorf("failed to create keypair")
	}

	h, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, fmt.Errorf("failed to parse address. Expects 'ip:port' format")
	}

	if h == "" {
		h = "127.0.0.1"
	}

	// construct host options
	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/%s/tcp/%s", h, port)),
		libp2p.Identity(priv),
	}

	// create host
	host, err := libp2p.New(context.Background(), opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create host > %s", err)
	}

	peer := &Peer{
		cfg:     config,
		address: util.FullAddressFromHost(host),
		host:    host,
		wg:      sync.WaitGroup{},
		log:     log,
		rSeed:   util.RandBytes(64),
	}

	peer.localPeer = peer
	peer.peerManager = NewManager(config, peer, peer.log)
	peer.IP = peer.ip()

	return peer, nil
}

// NewRemotePeer creates a Peer that represents a remote peer
func NewRemotePeer(address ma.Multiaddr, localPeer *Peer) *Peer {
	peer := &Peer{
		address:   address,
		localPeer: localPeer,
		remote:    true,
	}
	peer.IP = peer.ip()
	return peer
}

// PM returns the peer manager
func (p *Peer) PM() *Manager {
	return p.peerManager
}

// IsSame checks if p is the same as peer
func (p *Peer) IsSame(peer *Peer) bool {
	return p.StringID() == peer.StringID()
}

// DevMode returns whether the peer is in dev mode
func (p *Peer) DevMode() bool {
	return p.cfg.Peer.Dev
}

// IsSameID is like IsSame except it accepts string
func (p *Peer) IsSameID(id string) bool {
	return p.StringID() == id
}

// SetLocalPeer sets the local peer
func (p *Peer) SetLocalPeer(peer *Peer) {
	p.localPeer = peer
}

// addToPeerStore adds a remote peer to the host's peerstore
func (p *Peer) addToPeerStore(remote *Peer) *Peer {
	p.localPeer.Peerstore().AddAddr(remote.ID(), remote.GetIP4TCPAddr(), pstore.PermanentAddrTTL)
	return p
}

// newStream creates a stream to a remote peer
func (p *Peer) newStream(ctx context.Context, peerID peer.ID, protocolID string) (inet.Stream, error) {
	return p.Host().NewStream(ctx, peerID, protocol.ID(protocolID))
}

// SetProtocol sets the protocol implementation
func (p *Peer) SetProtocol(protoc Protocol) {
	p.protoc = protoc
}

// GetHost returns the peer's host
func (p *Peer) GetHost() host.Host {
	return p.host
}

// Peerstore returns the Peerstore of the host
func (p *Peer) Peerstore() pstore.Peerstore {
	if h := p.Host(); h != nil {
		return h.Peerstore()
	}
	return nil
}

// Host returns the internal host instance
func (p *Peer) Host() host.Host {
	return p.host
}

// ID returns the peer id of the host
func (p *Peer) ID() peer.ID {
	if p.address == nil {
		return ""
	}

	pid, _ := p.address.ValueForProtocol(ma.P_IPFS)
	id, _ := peer.IDB58Decode(pid)
	return id
}

// StringID is like ID() but returns string
func (p *Peer) StringID() string {
	if p.address == nil {
		return ""
	}

	pid, _ := p.address.ValueForProtocol(ma.P_IPFS)
	return pid
}

// ShortID is like IDPretty but shorter
func (p *Peer) ShortID() string {
	id := p.StringID()
	if len(id) == 0 {
		return ""
	}
	return id[0:12] + ".." + id[40:52]
}

// Connected checks whether the peer is connected to the local peer.
// Returns false if peer is the local peer.
func (p *Peer) Connected() bool {
	if p.localPeer == nil {
		return false
	}
	return len(p.localPeer.host.Network().ConnsToPeer(p.ID())) > 0
}

func (p *Peer) isDevMode() bool {
	return p.cfg.Peer.Dev
}

// IsKnown checks whether a peer is known to the local peer
func (p *Peer) IsKnown() bool {
	if p.localPeer == nil {
		return false
	}
	return p.localPeer.PM().GetKnownPeer(p.StringID()) != nil
}

// PrivKey returns the peer's private key
func (p *Peer) PrivKey() crypto.PrivKey {
	return p.host.Peerstore().PrivKey(p.host.ID())
}

// PubKey returns the peer's private key
func (p *Peer) PubKey() crypto.PubKey {
	return p.host.Peerstore().PubKey(p.host.ID())
}

// SetProtocolHandler sets the protocol handler for a specific protocol
func (p *Peer) SetProtocolHandler(version string, handler inet.StreamHandler) {
	p.host.SetStreamHandler(protocol.ID(version), handler)
}

// GetMultiAddr returns the full multi address of the peer
func (p *Peer) GetMultiAddr() string {
	if p.host == nil && !p.remote {
		return ""
	} else if p.remote {
		return p.address.String()
	}
	hostAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", p.host.ID().Pretty()))
	return p.host.Addrs()[0].Encapsulate(hostAddr).String()
}

// GetAddr returns the host and port of the peer as "host:port"
func (p *Peer) GetAddr() string {
	parts := strings.Split(strings.Trim(p.host.Addrs()[0].String(), "/"), "/")
	return fmt.Sprintf("%s:%s", parts[1], parts[3])
}

// GetIP4TCPAddr returns ip4 and tcp parts of the host's multi address
func (p *Peer) GetIP4TCPAddr() ma.Multiaddr {
	ipfsAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", p.ID().Pretty()))
	return p.address.Decapsulate(ipfsAddr)
}

// GetBindAddress returns the bind address
func (p *Peer) GetBindAddress() string {
	return p.address.String()
}

// AddBootstrapPeers sets the initial nodes to communicate to
func (p *Peer) AddBootstrapPeers(peerAddresses []string, hardcoded bool) error {

	for _, addr := range peerAddresses {

		if !util.IsValidAddr(addr) {
			p.log.Debug("Invalid bootstrap peer address", "PeerAddr", addr)
			continue
		}

		if p.isDevMode() && !util.IsDevAddr(util.GetIPFromAddr(addr)) {
			p.log.Debug("Only local or private address are allowed in dev mode", "Addr", addr)
			continue
		}

		if !p.DevMode() && !util.IsRoutableAddr(addr) {
			p.log.Debug("Invalid bootstrap peer address", "PeerAddr", addr)
			continue
		}

		pAddr, _ := ma.NewMultiaddr(addr)
		rp := NewRemotePeer(pAddr, p)
		rp.isHardcodedSeed = hardcoded
		rp.protoc = p.protoc
		p.peerManager.AddBootstrapPeer(rp)
	}
	return nil
}

// GetPeersPublicAddrs gets all the peers' public address.
// It will ignore any peer whose ID is specified in peerIDsToIgnore
func (p *Peer) GetPeersPublicAddrs(peerIDsToIgnore []string) (peerAddrs []ma.Multiaddr) {
	for _, _p := range p.host.Peerstore().Peers() {
		if !funk.Contains(peerIDsToIgnore, _p.Pretty()) {
			if _pAddrs := p.host.Peerstore().Addrs(_p); len(_pAddrs) > 0 {
				peerAddrs = append(peerAddrs, _pAddrs[0])
			}
		}
	}
	return
}

// connectToPeer handshake to each bootstrap peer.
// Then send GetAddr message if handshake is successful
func (p *Peer) connectToPeer(remotePeer *Peer) error {
	if p.protoc.SendHandshake(remotePeer) == nil {
		return p.protoc.SendGetAddr([]*Peer{remotePeer})
	}
	return nil
}

// Start starts the peer.
// Send handshake to each bootstrap peer.
// Then send GetAddr message if handshake is successful
func (p *Peer) Start() {
	p.PM().Manage()
	for _, peer := range p.PM().bootstrapPeers {
		go p.connectToPeer(peer)
	}
}

// Wait forces the current thread to wait for the peer
func (p *Peer) Wait() {
	p.wg.Add(1)
	p.wg.Wait()
}

// Stop stops the peer and release any held resources.
func (p *Peer) Stop() {
	p.PM().Stop()
	p.wg.Done()
}

// PeerFromAddr creates a Peer object from a multiaddr
func (p *Peer) PeerFromAddr(addr string, remote bool) (*Peer, error) {
	if !util.IsValidAddr(addr) {
		return nil, fmt.Errorf("addr is not valid")
	}
	pAddr, _ := ma.NewMultiaddr(addr)
	return &Peer{
		address:   pAddr,
		localPeer: p,
		protoc:    p.protoc,
		remote:    remote,
	}, nil
}

// ip returns the IP of the peer
func (p *Peer) ip() net.IP {
	addr := p.GetIP4TCPAddr()
	if addr == nil {
		return nil
	}
	ip, _ := addr.ValueForProtocol(ma.P_IP6)
	if ip == "" {
		ip, _ = addr.ValueForProtocol(ma.P_IP4)
	}
	return net.ParseIP(ip)
}

// IsBadTimestamp checks whether the timestamp of the peer is bad.
// It is bad when:
// - It has no timestamp
// - The timestamp is 10 minutes in the future or over 3 hours ago
func (p *Peer) IsBadTimestamp() bool {
	if p.Timestamp.IsZero() {
		return true
	}

	now := time.Now()
	if p.Timestamp.After(now.Add(time.Minute*10)) || p.Timestamp.Before(now.Add(-3*time.Hour)) {
		return true
	}

	return false
}
