package node

import (
	"context"
	"fmt"

	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/wire/messages"
	net "github.com/libp2p/go-libp2p-net"
)

// sendGetAddr sends a messages.GetAddr message to a remote peer.
// The remote peer will respond with a messages.Addr message which the function
// must process using the OnAddr handler and return the response.
func (g *Gossip) sendGetAddr(remotePeer types.Engine) ([]*messages.Address, error) {

	remotePeerIDShort := remotePeer.ShortID()

	s, err := g.newStream(context.Background(), remotePeer, config.GetAddrVersion)
	if err != nil {
		g.log.Debug("GetAddr message failed. failed to connect to peer", "Err", err, "PeerID", remotePeerIDShort)
		return nil, fmt.Errorf("getaddr failed. failed to connect to peer. %s", err)
	}
	defer s.Close()

	msg := &messages.GetAddr{}
	if err := writeStream(s, msg); err != nil {
		s.Reset()
		g.log.Debug("GetAddr failed. failed to write to stream", "Err", err, "PeerID", remotePeerIDShort)
		return nil, fmt.Errorf("getaddr failed. failed to write to stream")
	}

	g.log.Debug("GetAddr message sent to peer", "PeerID", remotePeerIDShort)

	return g.onAddr(s)
}

// SendGetAddr sends GetAddr message to peers in separate goroutines.
// GetAddr returns with a list of addr that should be relayed to other peers.
func (g *Gossip) SendGetAddr(remotePeers []types.Engine) error {

	// we need to know if wee need more peers before we requests
	// more addresses from other peers.
	if !g.PM().NeedMorePeers() {
		return nil
	}

	// for each remore peers, send the GetAddr message in different goroutines
	for _, remotePeer := range remotePeers {
		rp := remotePeer
		go func() {

			// send GetAddr and receive a list of address
			addressToRelay, err := g.sendGetAddr(rp)
			if err != nil {
				return
			}

			// As per discovery protocol, relay the addresses received
			if len(addressToRelay) > 0 {
				g.RelayAddr(addressToRelay)
			}
		}()
	}

	return nil
}

// OnGetAddr processes a messages.GetAddr request.
// Sends a list of active addresses to the sender
func (g *Gossip) OnGetAddr(s net.Stream) {

	defer s.Close()

	remotePeerIDShort := util.ShortID(s.Conn().RemotePeer())
	remoteAddr := util.FullRemoteAddressFromStream(s)
	remotePeer := NewRemoteNode(remoteAddr, g.engine)

	// check whether we can interact with this remote peer
	if yes, reason := g.engine.canAcceptPeer(remotePeer); !yes {
		s.Reset()
		g.log.Debug(fmt.Sprintf("Can't accept message from peer: %s", reason), "Addr", remotePeer.GetMultiAddr(), "Msg", "GetAddr")
		return
	}

	g.log.Debug("Received GetAddr message", "PeerID", remotePeerIDShort)

	// read the message
	msg := &messages.GetAddr{}
	if err := readStream(s, msg); err != nil {
		s.Reset()
		g.log.Error("failed to read getaddr message", "Err", err, "PeerID", remotePeerIDShort)
		return
	}

	// get active addresses we know about. If we have more 2500
	// addresses, then we select 2500 randomly
	activePeers := g.PM().GetActivePeers(0)
	if len(activePeers) > 2500 {
		activePeers = g.PM().GetRandomActivePeers(2500)
	}

	// Construct an Addr message and add addresses to it
	addr := &messages.Addr{}
	for _, peer := range activePeers {
		// Ignore an address if it is the same with the local node
		// and if it is a hardcoded seed address
		if !g.PM().IsLocalNode(peer) && !peer.IsSame(remotePeer) && !peer.IsHardcodedSeed() {
			addr.Addresses = append(addr.Addresses, &messages.Address{
				Address:   peer.GetMultiAddr(),
				Timestamp: peer.GetTimestamp().Unix(),
			})
		}
	}

	if err := writeStream(s, addr); err != nil {
		s.Reset()
		g.log.Error("failed to send GetAddr response", "Err", err)
		return
	}

	g.log.Debug("Sent GetAddr response to peer", "PeerID", remotePeerIDShort)
}
