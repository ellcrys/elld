package peer

import (
	"context"
	"fmt"
	"io"
	mrand "math/rand"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/thoas/go-funk"

	crypto "github.com/libp2p/go-libp2p-crypto"
	peer "github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"

	"github.com/ellcrys/gcoin/util"
	libp2p "github.com/libp2p/go-libp2p"
	host "github.com/libp2p/go-libp2p-host"
	inet "github.com/libp2p/go-libp2p-net"
	protocol "github.com/libp2p/go-libp2p-protocol"
	ma "github.com/multiformats/go-multiaddr"
	"go.uber.org/zap"
)

var (
	peerLog   *zap.SugaredLogger
	protocLog *zap.SugaredLogger
)

func init() {
	peerLog = util.NewLogger("/peer")
	protocLog = peerLog.Named("protocol.inception")
}

// SilenceLoggers changes the loggers in this package to NopLoggers. Called in test environment.
func SilenceLoggers() {
	peerLog = util.NewNopLogger()
	protocLog = util.NewNopLogger()
}

// Peer represents a network node
type Peer struct {
	address     ma.Multiaddr
	host        host.Host
	wg          sync.WaitGroup
	localPeer   *Peer
	peerManager *Manager
	protoc      Protocol
	remote      bool
	Timestamp   time.Time
}

// NewPeer creates a peer instance at the specified port
func NewPeer(address string, idSeed int64) (*Peer, error) {

	// generate peer identity
	priv, _, err := GenerateKeyPair(mrand.New(mrand.NewSource(idSeed)))
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
		address: util.FullAddressFromHost(host),
		host:    host,
		wg:      sync.WaitGroup{},
	}

	peer.localPeer = peer
	peer.peerManager = NewManager(peer)

	go func() {
		tm := time.NewTicker(10 * time.Second)
		for {
			select {
			case <-tm.C:
				fmt.Println("Num Address", len(peer.PM().GetActivePeers(-1)))
			}
		}
	}()

	return peer, nil
}

// NewRemotePeer creates a Peer that represents a remote peer
func NewRemotePeer(address ma.Multiaddr, localPeer *Peer) *Peer {
	return &Peer{
		address:   address,
		localPeer: localPeer,
		remote:    true,
	}
}

// GenerateKeyPair generates private and public keys
func GenerateKeyPair(r io.Reader) (crypto.PrivKey, crypto.PubKey, error) {
	return crypto.GenerateEd25519Key(r)
}

// PM returns the peer manager
func (p *Peer) PM() *Manager {
	return p.peerManager
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

// IDPretty is like ID() but returns string
func (p *Peer) IDPretty() string {
	if p.address == nil {
		return ""
	}

	pid, _ := p.address.ValueForProtocol(ma.P_IPFS)
	return pid
}

// PrivKey returns the peer's private key
func (p *Peer) PrivKey() crypto.PrivKey {
	return p.host.Peerstore().PrivKey(p.host.ID())
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
func (p *Peer) AddBootstrapPeers(peerAddresses []string) error {
	for _, addr := range peerAddresses {
		if !util.IsValidAddress(addr) {
			peerLog.Debugw("invalid bootstrap peer address", "PeerAddr", addr)
			continue
		}
		pAddr, _ := ma.NewMultiaddr(addr)
		rp := NewRemotePeer(pAddr, p)
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

// Start starts the peer
func (p *Peer) Start() {
	p.PM().Manage()

	// send handshake to bootstrap peers
	for _, b := range p.PM().bootstrapPeers {
		go p.protoc.DoSendHandshake(b)
	}
}

// Wait forces the current thread to wait for the peer
func (p *Peer) Wait() {
	p.wg.Add(1)
	p.wg.Wait()
}

// Stop stops the peer and release any held resources.
func (p *Peer) Stop() {
	p.wg.Done()
}

// PeerFromAddr creates a Peer object from a multiaddr
func (p *Peer) PeerFromAddr(addr string, remote bool) (*Peer, error) {
	pAddr, err := ma.NewMultiaddr(addr)
	if err != nil {
		return nil, err
	}
	return &Peer{
		address:   pAddr,
		localPeer: p,
		protoc:    p.protoc,
		remote:    remote,
	}, nil
}
