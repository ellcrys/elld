package node

import (
	"time"

	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/wire"
)

// SelfAdvertise sends an Addr message containing
// the address of the local peer to all connected peers.
func (g *Gossip) SelfAdvertise(connectedPeers []types.Engine) int {

	msg := &wire.Addr{
		Addresses: []*wire.Address{
			{Address: g.engine.GetAddress(), Timestamp: time.Now().Unix()},
		},
	}

	// Get the address of connected peers into a wire.Address
	var peersAddress = []*wire.Address{}
	for _, p := range connectedPeers {
		peersAddress = append(peersAddress, &wire.Address{
			Address:   p.GetAddress(),
			Timestamp: p.GetLastSeen().Unix(),
		})
	}

	// Select up to 2 peers to act as broadcasters
	g.PickBroadcasters(peersAddress, 2)

	sent := 0
	for _, peer := range g.broadcasters.Peers() {

		s, c, err := g.NewStream(peer, config.AddrVersion)
		if err != nil {
			g.logErr(err, peer, "[SelfAdvertise] Failed to connect")
			continue
		}
		defer c()
		defer s.Close()

		if err := WriteStream(s, msg); err != nil {
			s.Reset()
			g.logErr(err, peer, "[SelfAdvertise] Failed to write")
			continue
		}

		g.PM().AddOrUpdatePeer(peer)

		sent++
	}

	g.log.Debug("Self advertisement completed", "ConnectedPeers", len(connectedPeers),
		"NumAdvertisedTo", sent)

	return sent
}
