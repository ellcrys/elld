package gossip

import (
	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/types/core"
	net "github.com/libp2p/go-libp2p-net"
)

// SendGetAddrToPeer sends a core.GetAddr message to a remote peer.
// The remote peer will respond with a core.Addr message which
// must be processed using the OnAddr handler and return the response.
func (g *Gossip) SendGetAddrToPeer(rp core.Engine) ([]*core.Address, error) {

	s, c, err := g.NewStream(rp, config.GetAddrVersion)
	if err != nil {
		return nil, g.logConnectErr(err, rp, "[SendGetAddrToPeer] Failed to connect")
	}
	defer c()
	defer s.Close()

	msg := &core.GetAddr{}
	if err := WriteStream(s, msg); err != nil {
		s.Reset()
		return nil, g.logErr(err, rp, "[SendGetAddrToPeer] Failed to write")
	}

	g.log.Debug("GetAddr message sent to peer", "PeerID", rp.ShortID())

	addr, err := g.onAddr(s, rp)
	if err != nil {
		return nil, err
	}

	g.engine.GetEventEmitter().Emit(EventReceivedAddr)

	return addr, nil
}

// SendGetAddr sends simultaneous GetAddr message
// to the given peers. GetAddr returns with a list
// of address that should be relayed to other peers.
func (g *Gossip) SendGetAddr(remotePeers []core.Engine) error {

	// we need to know if we need more peers before we requests
	// more addresses from other peers.
	if !g.PM().RequirePeers() {
		return nil
	}

	for _, remotePeer := range remotePeers {
		go func(rp core.Engine) {

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

// OnGetAddr processes a core.GetAddr request.
// Sends a list of active addresses to the sender
func (g *Gossip) OnGetAddr(s net.Stream, rp core.Engine) error {

	defer s.Close()

	// read the message
	msg := &core.GetAddr{}
	if err := ReadStream(s, msg); err != nil {
		s.Reset()
		return g.logErr(err, rp, "[OnGetAddr] Failed to read")
	}

	g.log.Debug("Received GetAddr message", "PeerID", rp.ShortID())

	// get active addresses we know about. If we have more 2500
	// addresses, then we select 2500 randomly
	activePeers := g.PM().GetActivePeers(0)
	if len(activePeers) > 2500 {
		activePeers = g.PM().GetRandomActivePeers(2500)
	}

	// Construct an Addr message and add addresses to it
	addr := &core.Addr{}
	for _, peer := range activePeers {
		// Ignore an address if it is the same with the local node
		// and if it is an hardcoded seed address
		if g.PM().IsLocalNode(peer) || peer.IsSame(rp) ||
			peer.IsHardcodedSeed() {
			continue
		}
		addr.Addresses = append(addr.Addresses, &core.Address{
			Address:   peer.GetAddress(),
			Timestamp: peer.GetLastSeen().Unix(),
		})
	}

	if err := WriteStream(s, addr); err != nil {
		s.Reset()
		return g.logErr(err, rp, "[OnGetAddr] Failed to write")
	}

	g.log.Debug("Sent GetAddr response to peer", "PeerID", rp.ShortID())
	return nil
}
