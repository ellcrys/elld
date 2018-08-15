package node

import (
	"context"

	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/node/histcache"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/wire"
	net "github.com/libp2p/go-libp2p-net"
)

func makeBlockHistoryKey(block core.Block, peer types.Engine) histcache.MultiKey {
	return []interface{}{block.HashToHex(), peer.StringID()}
}

// RelayBlock sends a given block to remote peers.
func (g *Gossip) RelayBlock(block core.Block, remotePeers []types.Engine) error {

	sent := 0

	g.log.Debug("Relaying block to peers", "BlockNo", block.GetNumber(), "NumPeers", len(remotePeers))

	for _, peer := range remotePeers {

		historyKey := makeBlockHistoryKey(block, peer)

		// check if we have an history of sending or receiving this block
		// from this remote peer. If yes, do not relay
		if g.engine.History().Has(historyKey) {
			continue
		}

		// create a stream to the remote peer
		s, err := g.newStream(context.Background(), peer, config.BlockVersion)
		if err != nil {
			g.log.Debug("Block message failed. failed to connect to peer", "Err", err, "PeerID", peer.ShortID())
			continue
		}
		defer s.Close()

		// write to the stream
		if err := writeStream(s, block); err != nil {
			s.Reset()
			g.log.Debug("Block message failed. failed to write to stream", "Err", err, "PeerID", peer.ShortID())
			continue
		}

		// add new history
		g.engine.History().Add(historyKey)

		sent++
	}

	g.log.Info("Finished relaying block", "BlockNo", block.GetNumber(), "NumPeersSentTo", sent)

	return nil
}

// OnBlock handles incoming block message
func (g *Gossip) OnBlock(s net.Stream) {

	defer s.Close()

	remotePeer := NewRemoteNode(util.FullRemoteAddressFromStream(s), g.engine)
	remotePeerIDShort := remotePeer.ShortID()

	// read the message
	block := &wire.Block{}
	if err := readStream(s, block); err != nil {
		s.Reset()
		g.log.Error("Failed to read block message", "Err", err, "PeerID", remotePeerIDShort)
		return
	}

	g.log.Info("Received a block", "BlockNo", block.GetNumber(), "Difficulty", block.GetHeader().GetDifficulty())

	// make a key for this block to be added to the history cache
	// so we always know when we have processed it in case
	// we see it again.
	historyKey := makeBlockHistoryKey(block, remotePeer)

	// check if we have an history about this block
	// with the remote peer, if no, process the block.
	if !g.engine.History().Has(historyKey) {

		// Add the transaction to the transaction pool and wait for error response
		if _, err := g.engine.bchain.ProcessBlock(block); err != nil {
			return
		}

		// add transaction to the history cache using the key we created earlier
		g.engine.History().Add(historyKey)
	}
}
