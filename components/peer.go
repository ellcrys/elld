package components

import (
	"context"
	"fmt"
	mrand "math/rand"
	"sync"

	"github.com/ellcrys/garagecoin/protocols"
	libp2p "github.com/libp2p/go-libp2p"
	host "github.com/libp2p/go-libp2p-host"
	protocol "github.com/libp2p/go-libp2p-protocol"
	ma "github.com/multiformats/go-multiaddr"
)

// Peer represents a network node
type Peer struct {
	handlers map[string]protocols.Protocol
	host     host.Host
	wg       sync.WaitGroup
}

// NewPeer creates a peer instance at the specified port
func NewPeer(port int, idSeed int64) (*Peer, error) {

	// generate peer identity
	priv, _, err := GenerateKeyPair(mrand.New(mrand.NewSource(idSeed)))
	if err != nil {
		return nil, fmt.Errorf("failed to create keypair")
	}

	// construct host options
	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", port)),
		libp2p.Identity(priv),
	}

	// create host
	host, err := libp2p.New(context.Background(), opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create host > %s", err)
	}

	return &Peer{
		handlers: make(map[string]protocols.Protocol),
		host:     host,
		wg:       sync.WaitGroup{},
	}, nil
}

// SetProtocolHandler sets the protocol handler for a specific protocol
func (p *Peer) SetProtocolHandler(protoc protocols.Protocol) {
	p.handlers[protoc.GetVersion()] = protoc
	p.host.SetStreamHandler(protocol.ID(protoc.GetVersion()), protoc.Handle)
}

// GetAddress returns the full address of the peer
func (p *Peer) GetAddress() string {
	hostAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", p.host.ID().Pretty()))
	return p.host.Addrs()[0].Encapsulate(hostAddr).String()
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
