package node

import (
	"bufio"
	"context"
	"time"

	"github.com/ellcrys/druid/util"
	"github.com/ellcrys/druid/wire"
	pc "github.com/multiformats/go-multicodec/protobuf"
)

// SelfAdvertise sends an Addr message containing the address of only the local peer
// to all connected peers known to the local peer.
// The caller is responsible for ensuring only connected peers are passed.
// Returns the number of peers advertised to.
func (pt *Inception) SelfAdvertise(connectedPeers []*Node) int {

	pt.log.Info("Attempting to advertise self", "ConnectedPeers", len(connectedPeers))

	msg := &wire.Addr{Addresses: []*wire.Address{{Address: pt.LocalPeer().GetMultiAddr(), Timestamp: time.Now().Unix()}}}

	successfullySent := 0
	for _, peer := range connectedPeers {

		s, err := pt.LocalPeer().addToPeerStore(peer).newStream(context.Background(), peer.ID(), util.AddrVersion)
		if err != nil {
			pt.log.Error("selfAdvertise failed. Failed to connect to peer", "Err", err, "PeerID", peer.ShortID())
			continue
		}

		w := bufio.NewWriter(s)
		if err := pc.Multicodec(nil).Encoder(w).Encode(msg); err != nil {
			pt.log.Debug("Addr failed. failed to write to stream", "Err", err, "PeerID", peer.ShortID())
			continue
		}

		successfullySent++
		w.Flush()
		s.Close()
	}

	pt.log.Info("Self advertisement attempt completed", "ConnectedPeers", len(connectedPeers), "NumAdvertisedTo", successfullySent)
	return successfullySent
}
