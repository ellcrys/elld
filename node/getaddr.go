package node

import (
	"context"
	"fmt"

	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/wire"
	net "github.com/libp2p/go-libp2p-net"
)

// SendGetAddrToPeer sends a wire.GetAddr message to a remote peer.
// The remote peer will respond with a wire.Addr message which
// must be processed using the OnAddr handler and return the response.
func (g *Gossip) SendGetAddrToPeer(remotePeer types.Engine) ([]*wire.Address, error) {

	remotePeerIDShort := remotePeer.ShortID()

	s, err := g.NewStream(context.Background(), remotePeer, config.GetAddrVersion)
	if err != nil {
		g.log.Debug("GetAddr message failed. failed to connect to peer", "Err", err, "PeerID", remotePeerIDShort)
		return nil, fmt.Errorf("getaddr failed. failed to connect to peer. %s", err)
	}
	defer s.Close()

	msg := &wire.GetAddr{}
	if err := WriteStream(s, msg); err != nil {
		s.Reset()
		g.log.Debug("GetAddr failed. failed to write to stream", "Err", err, "PeerID", remotePeerIDShort)
		return nil, fmt.Errorf("getaddr failed. failed to write to stream")
	}

	g.log.Debug("GetAddr message sent to peer", "PeerID", remotePeerIDShort)

	addr, err := g.onAddr(s)
	if err != nil {
		return nil, err
	}

	defer g.engine.event.Emit(EventReceivedAddr)

	return addr, nil
}

// SendGetAddr sends simultaneous GetAddr message
// to the given peers. GetAddr returns with a list
// of address that should be relayed to other peers.
func (g *Gossip) SendGetAddr(remotePeers []types.Engine) error {

	// we need to know if wee need more peers before we requests
	// more addresses from other peers.
	if !g.PM().NeedMorePeers() {
		return nil
	}

	for _, remotePeer := range remotePeers {
		go func(rp types.Engine) {
			// send GetAddr and receive a list of address
			addressToRelay, err := g.SendGetAddrToPeer(rp)
			if err != nil {
				return
			}

			// As per discovery protocol,
			// relay the addresses received
			if len(addressToRelay) > 0 {
				g.RelayAddresses(addressToRelay)
			}
		}(remotePeer)
	}

	return nil
}

// OnGetAddr processes a wire.GetAddr request.
// Sends a list of active addresses to the sender
func (g *Gossip) OnGetAddr(s net.Stream) {

	defer s.Close()

	remotePeerIDShort := util.ShortID(s.Conn().RemotePeer())
	remoteAddr := util.FullRemoteAddressFromStream(s)
	remotePeer := NewRemoteNode(remoteAddr, g.engine)

	// check whether we can interact with this remote peer
	if ok, err := g.engine.canAcceptPeer(remotePeer); !ok {
		s.Reset()
		g.log.Debug(fmt.Sprintf("Can't accept message from peer: %s", err.Error()),
			"Addr", remotePeer.GetMultiAddr(), "Msg", "GetAddr")
		return
	}

	// read the message
	msg := &wire.GetAddr{}
	if err := ReadStream(s, msg); err != nil {
		s.Reset()
		g.log.Error("failed to read getaddr message", "Err", err, "PeerID", remotePeerIDShort)
		return
	}

	g.log.Debug("Received GetAddr message", "PeerID", remotePeerIDShort)

	// get active addresses we know about. If we have more 2500
	// addresses, then we select 2500 randomly
	activePeers := g.PM().GetActivePeers(0)
	if len(activePeers) > 2500 {
		activePeers = g.PM().GetRandomActivePeers(2500)
	}

	// Construct an Addr message and add addresses to it
	addr := &wire.Addr{}
	for _, peer := range activePeers {
		// Ignore an address if it is the same with the local node
		// and if it is an hardcoded seed address
		if g.PM().IsLocalNode(peer) || peer.IsSame(remotePeer) || peer.IsHardcodedSeed() {
			continue
		}
		addr.Addresses = append(addr.Addresses, &wire.Address{
			Address:   peer.GetMultiAddr(),
			Timestamp: peer.GetTimestamp().Unix(),
		})
	}

	if err := WriteStream(s, addr); err != nil {
		s.Reset()
		g.log.Error("failed to send GetAddr response", "Err", err)
		return
	}

	g.log.Debug("Sent GetAddr response to peer", "PeerID", remotePeerIDShort)
}
