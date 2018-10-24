package node

import (
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

	rpIDShort := remotePeer.ShortID()

	s, c, err := g.NewStream(remotePeer, config.GetAddrVersion)
	if err != nil {
		return nil, g.logErr(err, remotePeer, "[SendGetAddrToPeer] Failed to connect")
	}
	defer c()
	defer s.Close()

	msg := &wire.GetAddr{}
	if err := WriteStream(s, msg); err != nil {
		s.Reset()
		return nil, g.logErr(err, remotePeer, "[SendGetAddrToPeer] Failed to write")
	}

	g.PM().AddOrUpdatePeer(remotePeer)
	g.log.Debug("GetAddr message sent to peer", "PeerID", rpIDShort)

	addr, err := g.onAddr(s)
	if err != nil {
		return nil, err
	}

	g.engine.event.Emit(EventReceivedAddr)

	return addr, nil
}

// SendGetAddr sends simultaneous GetAddr message
// to the given peers. GetAddr returns with a list
// of address that should be relayed to other peers.
func (g *Gossip) SendGetAddr(remotePeers []types.Engine) error {

	// we need to know if we need more peers before we requests
	// more addresses from other peers.
	if !g.PM().RequirePeers() {
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
				go g.RelayAddresses(addressToRelay)
			}
		}(remotePeer)
	}

	return nil
}

// OnGetAddr processes a wire.GetAddr request.
// Sends a list of active addresses to the sender
func (g *Gossip) OnGetAddr(s net.Stream) {

	defer s.Close()

	remoteAddr := util.RemoteAddrFromStream(s)
	rp := NewRemoteNode(remoteAddr, g.engine)
	rpIDShort := rp.ShortID()

	// check whether we are allowed to receive this peer's message
	if ok, err := g.engine.canAcceptPeer(rp); !ok {
		g.logErr(err, rp, "message unaccepted")
		return
	}

	// read the message
	msg := &wire.GetAddr{}
	if err := ReadStream(s, msg); err != nil {
		s.Reset()
		g.logErr(err, rp, "[OnGetAddr] Failed to read")
		return
	}

	g.PM().AddOrUpdatePeer(rp)
	g.log.Debug("Received GetAddr message", "PeerID", rpIDShort)

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
		if g.PM().IsLocalNode(peer) || peer.IsSame(rp) ||
			peer.IsHardcodedSeed() {
			continue
		}
		addr.Addresses = append(addr.Addresses, &wire.Address{
			Address:   peer.GetAddress(),
			Timestamp: peer.GetLastSeen().Unix(),
		})
	}

	if err := WriteStream(s, addr); err != nil {
		s.Reset()
		g.logErr(err, rp, "[OnGetAddr] Failed to write")
		return
	}

	g.log.Debug("Sent GetAddr response to peer", "PeerID", rpIDShort)
}
