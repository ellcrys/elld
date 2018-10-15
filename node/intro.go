package node

import (
	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/util/cache"
	"github.com/ellcrys/elld/wire"
	net "github.com/libp2p/go-libp2p-net"
)

// SendIntro sends a wire.Intro message to
// random, connected peers. If intro is nil,
// a new wire.Intro message is created and
// broadcast to the selected peers
func (g *Gossip) SendIntro(intro *wire.Intro) {

	// Get the addresses of the nodes
	// this peer is connected to.
	var connectedAddresses = []*wire.Address{}
	for _, p := range g.PM().GetConnectedPeers() {
		connectedAddresses = append(connectedAddresses, &wire.Address{
			Address:   p.GetAddress(),
			Timestamp: p.GetLastSeen().Unix(),
		})
	}

	// From our peer list, randomly pick 2 peers to broad cast to.
	broadcasters := g.PickBroadcasters(connectedAddresses, 2)

	var msg = intro
	if msg == nil {
		msg = &wire.Intro{PeerID: g.engine.StringID()}
	}

	// Send intro message to the selected
	// broadcast nodes.
	sent := 0
	for _, peer := range broadcasters {

		// Don't relay an intro back to the
		// peer that authored it
		if peer.StringID() == msg.PeerID {
			continue
		}

		// If we had recently relayed or received
		// an intro from/to this peer, don't relay
		if g.engine.history.HasMulti(peer.StringID(), msg.Hash().HexStr()) {
			continue
		}

		s, c, err := g.NewStream(peer, config.IntroVersion)
		if err != nil {
			g.log.Error("Intro message failed. Unable to connect to peer",
				"Err", err,
				"PeerID", peer.ShortID())
			continue
		}
		defer c()
		defer s.Close()

		if err := WriteStream(s, msg); err != nil {
			s.Reset()
			g.log.Error("Intro message failed. Unable to write to stream",
				"Err", err,
				"PeerID", peer.ShortID())
			continue
		}

		g.PM().UpdateLastSeenTime(peer)

		sent++

		g.engine.history.AddMulti(cache.Sec(3600), peer.StringID(), msg.Hash().HexStr())
	}

	g.log.Debug("Sent Intro to peer(s)",
		"NumBroadcastPeers", broadcasters.Len(),
		"NumSentTo", sent)
}

// OnIntro handles incoming wire.Intro messages.
// Received messages are relayed to 2 random peers.
func (g *Gossip) OnIntro(s net.Stream) {

	defer s.Reset()
	remoteAddr := util.RemoteAddrFromStream(s)
	remotePeer := NewRemoteNode(remoteAddr, g.engine)

	var msg wire.Intro
	if err := ReadStream(s, &msg); err != nil {
		g.log.Debug("Failed to read Intro message", "Err", err, "PeerID",
			remotePeer.ShortID())
		return
	}

	// Add remote peer into the intro cache
	// with a TTL of 1 hour.
	g.engine.intros.AddWithExp(msg.PeerID, struct{}{}, cache.Sec(3600))

	g.log.Debug("Received and cached intro message",
		"TotalIntros", g.engine.intros.Len())

	g.engine.event.Emit(EventIntroReceived)

	// Relay the received message to our own peers
	go g.SendIntro(&msg)
}
