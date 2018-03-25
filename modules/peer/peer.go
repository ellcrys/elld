package peer

import (
	"context"
	"fmt"
	mrand "math/rand"
	"net"
	"strings"
	"sync"

	"github.com/ellcrys/garagecoin/modules/types"

	pstore "github.com/libp2p/go-libp2p-peerstore"

	"github.com/ellcrys/garagecoin/modules"
	libp2p "github.com/libp2p/go-libp2p"
	host "github.com/libp2p/go-libp2p-host"
	"github.com/libp2p/go-libp2p-peer"
	protocol "github.com/libp2p/go-libp2p-protocol"
	ma "github.com/multiformats/go-multiaddr"
	"go.uber.org/zap"
)

var peerLog *zap.SugaredLogger

func init() {
	peerLog = modules.NewLogger("/peer")
}

// Peer represents a network node
type Peer struct {
	address            ma.Multiaddr
	handlers           map[string]types.StreamProtocol
	host               host.Host
	wg                 sync.WaitGroup
	peers              []*Peer
	do                 Logic
	localPeer          *Peer
	curProtocolVersion protocol.ID
}

// NewPeer creates a peer instance at the specified port
func NewPeer(address string, idSeed int64) (*Peer, error) {

	// generate peer identity
	priv, _, err := modules.GenerateKeyPair(mrand.New(mrand.NewSource(idSeed)))
	if err != nil {
		return nil, fmt.Errorf("failed to create keypair")
	}

	h, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, fmt.Errorf("failed to parse address")
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

	return &Peer{
		address:  host.Addrs()[0],
		handlers: make(map[string]types.StreamProtocol),
		host:     host,
		wg:       sync.WaitGroup{},
		do:       &Do{},
	}, nil
}

// SetCurrentProtocol sets the protocol version to use in future communications
func (p *Peer) SetCurrentProtocol(value string) {
	p.curProtocolVersion = protocol.ID(value)
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

// SetProtocolHandler sets the protocol handler for a specific protocol
func (p *Peer) SetProtocolHandler(protoc types.StreamProtocol) {
	p.handlers[protoc.GetVersion()] = protoc
	p.host.SetStreamHandler(protocol.ID(protoc.GetVersion()), protoc.Handle)
}

// GetMultiAddr returns the full multi address of the peer
func (p *Peer) GetMultiAddr() string {
	hostAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", p.host.ID().Pretty()))
	return p.host.Addrs()[0].Encapsulate(hostAddr).String()
}

// GetAddr returns the host and port of the peer as "host:port"
func (p *Peer) GetAddr() string {
	parts := strings.Split(strings.Trim(p.host.Addrs()[0].String(), "/"), "/")
	return fmt.Sprintf("%s:%s", parts[1], parts[3])
}

// GetIP4Addr returns ip4 part of the host's multi address
func (p *Peer) GetIP4Addr() ma.Multiaddr {
	ipfsAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", p.ID().Pretty()))
	return p.address.Decapsulate(ipfsAddr)
}

// GetBindAddress returns the bind address
func (p *Peer) GetBindAddress() string {
	return p.address.String()
}

// PostMessage sends a message to the peer
func (p *Peer) PostMessage(msg []byte) error {
	return nil
}

// SetBootstrapNodes sets the initial nodes to communicate to
func (p *Peer) SetBootstrapNodes(peerAddresses []string) error {
	for _, addr := range peerAddresses {
		pAddr, err := ma.NewMultiaddr(addr)
		if err != nil {
			return fmt.Errorf("invalid bootstrap node address. Expected a valid multi address")
		}
		go func() {
			err := p.do.SendHandShake(&Peer{
				address:   pAddr,
				localPeer: p,
			})
			fmt.Println(err)
		}()
	}
	return nil
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
