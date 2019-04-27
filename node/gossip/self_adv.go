package gossip

import (
	"time"

	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/types/core"
)

// SelfAdvertise sends an Addr message containing
// the address of the local peer to all connected peers.
func (g *Manager) SelfAdvertise(connectedPeers []core.Engine) int {

	msg := &core.Addr{
		Addresses: []*core.Address{
			{Address: g.engine.GetAddress(), Timestamp: time.Now().Unix()},
		},
	}

	// Select peers to act as broadcasters
	bp := g.PickBroadcastersFromPeers(g.randBroadcasters, connectedPeers, 3)

	sent := 0
	for _, peer := range bp.Peers() {

		s, c, err := g.NewStream(peer, config.GetVersions().Addr)
		if err != nil {
			g.logConnectErr(err, peer, "[SelfAdvertise] Failed to connect")
			continue
		}
		defer c()
		defer s.Close()

		if err := WriteStream(s, msg); err != nil {
			s.Reset()
			g.logErr(err, peer, "[SelfAdvertise] Failed to write")
			continue
		}

		sent++
	}

	g.log.Debug("Self advertisement completed",
		"ConnectedPeers", len(connectedPeers),
		"NumAdvertisedTo", sent)

	return sent
}
