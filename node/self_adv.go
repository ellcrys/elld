package node

import (
	"context"
	"time"

	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/wire"
)

// SelfAdvertise sends an Addr message containing the address of the local peer
// to all connected peers known to the local peer.
// The caller is responsible for ensuring only connected peers are passed.
// Returns the number of peers advertised to.
func (g *Gossip) SelfAdvertise(connectedPeers []*Node) int {

	msg := &wire.Addr{Addresses: []*wire.Address{{Address: g.engine.GetMultiAddr(), Timestamp: time.Now().Unix()}}}
	successfullySent := 0

	for _, peer := range connectedPeers {

		// create stream
		s, err := g.newStream(context.Background(), peer, config.AddrVersion)
		if err != nil {
			g.log.Error("selfAdvertise failed. Failed to connect to peer", "Err", err, "PeerID", peer.ShortID())
			continue
		}
		defer s.Close()

		// write to the stream
		if err := writeStream(s, msg); err != nil {
			s.Reset()
			g.log.Error("Addr failed. failed to write to stream", "Err", err, "PeerID", peer.ShortID())
			continue
		}

		successfullySent++
	}

	g.log.Info("Self advertisement completed", "ConnectedPeers", len(connectedPeers), "NumAdvertisedTo", successfullySent)

	return successfullySent
}
