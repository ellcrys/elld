package node

import (
	"time"

	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/wire"
)

// SelfAdvertise sends an Addr message
// containing the address of the local peer
// to all connected peers.
func (g *Gossip) SelfAdvertise(connectedPeers []types.Engine) int {

	msg := &wire.Addr{
		Addresses: []*wire.Address{
			{Address: g.engine.GetAddress(), Timestamp: time.Now().Unix()},
		},
	}

	sent := 0
	for _, peer := range connectedPeers {

		s, c, err := g.NewStream(peer, config.AddrVersion)
		if err != nil {
			g.log.Error("selfAdvertise failed. Failed to connect to peer",
				"Err", err, "PeerID", peer.ShortID())
			continue
		}
		defer c()
		defer s.Close()

		// write to the stream
		if err := WriteStream(s, msg); err != nil {
			s.Reset()
			g.log.Error("Addr failed. failed to write to stream",
				"Err", err, "PeerID", peer.ShortID())
			continue
		}

		g.PM().UpdateLastSeenTime(peer)

		sent++
	}

	g.log.Debug("Self advertisement completed",
		"ConnectedPeers", len(connectedPeers), "NumAdvertisedTo", sent)

	return sent
}
