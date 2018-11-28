package gossip

import (
	"fmt"

	"github.com/jinzhu/copier"

	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/types/core"

	net "github.com/libp2p/go-libp2p-net"
)

// SendPingToPeer sends a Ping message to the given peer.
// It receives the response Pong message and will start
// blockchain synchronization if the Pong message includes
// blockchain information that is better than the local blockchain
func (g *Gossip) SendPingToPeer(remotePeer core.Engine) error {

	rpIDShort := remotePeer.ShortID()
	s, c, err := g.NewStream(remotePeer, config.Versions.Ping)
	if err != nil {
		return g.logConnectErr(err, remotePeer, "[SendPingToPeer] Failed to connect")
	}
	defer c()
	defer s.Close()

	// Determine the best block and the total difficulty of the block.
	// Add the info to the ping message.
	bestBlock, err := g.GetBlockchain().ChainReader().GetBlock(0)
	if err != nil {
		g.logErr(err, remotePeer, "[SendPingToPeer] Failed to determine best block")
		return fmt.Errorf("failed to determine best block: %s", err)
	}

	msg := &core.Ping{
		BestBlockHash:            bestBlock.GetHash(),
		BestBlockNumber:          bestBlock.GetNumber(),
		BestBlockTotalDifficulty: bestBlock.GetHeader().GetTotalDifficulty(),
	}

	// construct the message and write it to the stream
	if err := WriteStream(s, msg); err != nil {
		return g.logErr(err, remotePeer, "[SendPingToPeer] Failed to write message")
	}

	g.log.Debug("Sent ping to peer",
		"PeerID", rpIDShort)

	// receive pong response from the remote peer
	pongMsg := &core.Pong{}
	if err := ReadStream(s, pongMsg); err != nil {
		return g.logErr(err, remotePeer, "[SendPingToPeer] Failed to read message")
	}

	// update the remote peer's timestamp
	g.PM().AddOrUpdateNode(remotePeer)

	g.log.Info("Received pong response from peer", "PeerID", rpIDShort)

	// compare best chain.
	// If the blockchain best block has a less
	// total difficulty, then need to start the block sync process
	bestBlock, _ = g.GetBlockchain().ChainReader().Current()
	if bestBlock.GetHeader().GetTotalDifficulty().
		Cmp(pongMsg.BestBlockTotalDifficulty) == -1 {
		g.log.Info("Local blockchain is behind peer",
			"PeerID", rpIDShort,
			"Height", pongMsg.BestBlockNumber,
			"TotalDifficulty", pongMsg.BestBlockTotalDifficulty)
		g.log.Info("Attempting to sync blockchain with peer", "PeerID", rpIDShort)
		go g.SendGetBlockHashes(remotePeer, nil)
	}

	// Update the current known best
	// remote block information
	var bestBlockInfo core.BestBlockInfo
	copier.Copy(&bestBlockInfo, pongMsg)
	g.engine.UpdateSyncInfo(&bestBlockInfo)

	return nil
}

// SendPing sends a ping message
func (g *Gossip) SendPing(remotePeers []core.Engine) {
	sent := 0
	for _, peer := range remotePeers {
		if !g.PM().IsAcquainted(peer) {
			continue
		}
		sent++
		go func(peer core.Engine) {
			if err := g.SendPingToPeer(peer); err != nil {
				g.PM().HasDisconnected(peer.GetAddress())
			}
		}(peer)
	}
	g.log.Debug("Sent ping to peer(s)", "NumPeers", len(remotePeers),
		"NumSentTo", sent)
}

// OnPing processes a Ping message sent by a remote peer.
// It sends back a Pong message containing information
// about the main block chain. It will start
// blockchain synchronization if the Ping message includes
// blockchain information that is better than the local
// blockchain
func (g *Gossip) OnPing(s net.Stream, rp core.Engine) error {

	defer s.Close()

	g.log.Info("Received ping message", "PeerID", rp.ShortID())

	// read the message from the stream
	msg := &core.Ping{}
	if err := ReadStream(s, msg); err != nil {
		return g.logErr(err, rp, "[OnPing] Failed to read message")
	}

	// determine the best block an the total
	// difficulty of the block. Add these info
	// to the pong message.
	bestBlock, err := g.GetBlockchain().ChainReader().GetBlock(0)
	if err != nil {
		g.log.Error("Pong failed. Failed to determine best block", "Err", err)
		return err
	}

	pongMsg := &core.Pong{
		BestBlockHash:            bestBlock.GetHash(),
		BestBlockNumber:          bestBlock.GetNumber(),
		BestBlockTotalDifficulty: bestBlock.GetHeader().GetTotalDifficulty(),
	}

	// send pong message
	if err := WriteStream(s, pongMsg); err != nil {
		return g.logErr(err, rp, "[OnPing] Failed to write message")
	}

	g.log.Debug("Sent pong response to peer", "PeerID", rp.ShortID())

	// compare best chain.
	// If the blockchain best block has a less
	// total difficulty, then need to start the block sync process
	bestBlock, _ = g.GetBlockchain().ChainReader().Current()
	if bestBlock.GetHeader().GetTotalDifficulty().
		Cmp(msg.BestBlockTotalDifficulty) == -1 {

		g.log.Info("Local blockchain is behind peer",
			"PeerID", rp.ShortID(),
			"Height", msg.BestBlockNumber,
			"TotalDifficulty", msg.BestBlockTotalDifficulty)
		g.log.Info("Attempting to sync blockchain with peer",
			"PeerID", rp.ShortID())
		go g.SendGetBlockHashes(rp, nil)
	}

	// Update the current known best
	// remote block information
	var bestBlockInfo core.BestBlockInfo
	copier.Copy(&bestBlockInfo, msg)
	g.engine.UpdateSyncInfo(&bestBlockInfo)

	return nil
}
