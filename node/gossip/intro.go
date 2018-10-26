package gossip

import (
	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/util/cache"
	net "github.com/libp2p/go-libp2p-net"
)

// SendIntro sends a core.Intro message to
// random, connected peers. If intro is nil,
// a new core.Intro message is created and
// broadcast to the selected peers
func (g *Gossip) SendIntro(intro *core.Intro) {

	// Get the addresses of the nodes
	// this peer is connected to.
	var connectedAddresses = []*core.Address{}
	for _, p := range g.PM().GetConnectedPeers() {
		connectedAddresses = append(connectedAddresses, &core.Address{
			Address:   p.GetAddress(),
			Timestamp: p.GetLastSeen().Unix(),
		})
	}

	// Select up to 2 peers to act as broadcasters
	g.PickBroadcasters(connectedAddresses, 2)

	var msg = intro
	if msg == nil {
		msg = &core.Intro{PeerID: g.engine.StringID()}
	}

	// Send intro message to the selected
	// broadcast nodes.
	sent := 0
	for _, peer := range g.broadcasters.Peers() {

		// Don't relay an intro back to the
		// peer that authored it
		if peer.StringID() == msg.PeerID {
			continue
		}

		// If we had recently relayed or received
		// an intro from/to this peer, don't relay
		if g.engine.GetHistory().HasMulti(peer.StringID(), msg.Hash().HexStr()) {
			continue
		}

		s, c, err := g.NewStream(peer, config.IntroVersion)
		if err != nil {
			g.logConnectErr(err, peer, "[SendIntro] Failed to connect")
			continue
		}
		defer c()
		defer s.Close()

		if err := WriteStream(s, msg); err != nil {
			s.Reset()
			g.logErr(err, peer, "[SendIntro] Failed to write")
			continue
		}

		g.PM().AddOrUpdateNode(peer)

		sent++

		g.engine.GetHistory().AddMulti(cache.Sec(3600), peer.StringID(),
			msg.Hash().HexStr())
	}

	g.log.Debug("Sent Intro to peer(s)", "NumBroadcastPeers", g.broadcasters.Len(),
		"NumSentTo", sent)
}

// OnIntro handles incoming core.Intro messages.
// Received messages are relayed to 2 random peers.
func (g *Gossip) OnIntro(s net.Stream) {

	defer s.Reset()
	remoteAddr := util.RemoteAddrFromStream(s)
	rp := g.engine.NewRemoteNode(remoteAddr)

	// check whether we are allowed to receive this peer's message
	if ok, err := g.PM().CanAcceptNode(rp); !ok {
		g.logErr(err, rp, "message unaccepted")
		return
	}

	var msg core.Intro
	if err := ReadStream(s, &msg); err != nil {
		g.logErr(err, rp, "[OnIntro] Failed to read")
		return
	}

	// Add remote peer into the intro cache
	// with a TTL of 1 hour.
	introCache := g.engine.GetIntros()
	introCache.AddWithExp(msg.PeerID, struct{}{}, cache.Sec(3600))

	g.log.Debug("Received and cached intro message", "Total", introCache.Len())

	g.engine.GetEventEmitter().Emit(EventIntroReceived)

	// Relay the received message to our own peers
	go g.SendIntro(&msg)
}
