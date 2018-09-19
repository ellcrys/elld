package node

import (
	"context"
	"fmt"
	"time"

	"github.com/jinzhu/copier"

	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/wire"

	"github.com/ellcrys/elld/util"
	net "github.com/libp2p/go-libp2p-net"
)

// SendPingToPeer sends a Ping message to the given peer.
// It receives the response Pong message and will start
// blockchain synchronization if the Pong message includes
// blockchain information that is better than the local blockchain
func (g *Gossip) SendPingToPeer(remotePeer types.Engine) error {

	remotePeerIDShort := remotePeer.ShortID()

	// create stream to the remote peer
	s, err := g.NewStream(context.Background(), remotePeer, config.PingVersion)
	if err != nil {
		g.log.Debug("Ping failed. failed to connect to peer", "Err", err, "PeerID", remotePeerIDShort)
		return fmt.Errorf("ping failed. failed to connect to peer. %s", err.Error())
	}
	defer s.Close()

	// determine the best block an the total
	// difficulty of the block. Add these info
	// to the ping message.
	bestBlock, err := g.GetBlockchain().ChainReader().GetBlock(0)
	if err != nil {
		g.log.Error("Ping failed. Failed to determine best block", "Err", err)
		return fmt.Errorf("ping failed: failed to determine best block: %s", err)
	}

	msg := &wire.Ping{
		BestBlockHash:            bestBlock.GetHash(),
		BestBlockNumber:          bestBlock.GetNumber(),
		BestBlockTotalDifficulty: bestBlock.GetHeader().GetTotalDifficulty(),
	}

	// construct the message and write it to the stream
	if err := WriteStream(s, msg); err != nil {
		g.log.Debug("ping failed. failed to write to stream", "Err", err, "PeerID", remotePeerIDShort)
		return fmt.Errorf("ping failed. failed to write to stream")
	}

	g.log.Info("Sent ping to peer", "PeerID", remotePeerIDShort)

	// receive pong response from the remote peer
	pongMsg := &wire.Pong{}
	if err := ReadStream(s, pongMsg); err != nil {
		g.log.Debug("Failed to read pong response", "Err", err, "PeerID", remotePeerIDShort)
		return fmt.Errorf("failed to read pong response")
	}

	// update the remote peer's timestamp
	// and add the remote peer to the peer manager's list
	remotePeer.SetTimestamp(time.Now())
	g.PM().AddOrUpdatePeer(remotePeer)

	g.log.Info("Received pong response from peer", "PeerID", remotePeerIDShort)

	// compare best chain.
	// If the blockchain best block has a less
	// total difficulty, then need to start the block sync process
	bestBlock, _ = g.GetBlockchain().ChainReader().Current()
	if bestBlock.GetHeader().GetTotalDifficulty().Cmp(pongMsg.BestBlockTotalDifficulty) == -1 {
		g.log.Info("Local blockchain is behind peer",
			"PeerID", remotePeerIDShort,
			"Height", pongMsg.BestBlockNumber,
			"TotalDifficulty", pongMsg.BestBlockTotalDifficulty)
		g.log.Info("Attempting to sync blockchain with peer", "PeerID", remotePeerIDShort)
		go g.SendGetBlockHashes(remotePeer, util.Hash{})
	}

	// Update the current known best
	// remote block information
	var bestBlockInfo BestBlockInfo
	copier.Copy(&bestBlockInfo, pongMsg)
	g.engine.updateSyncInfo(&bestBlockInfo)

	return nil
}

// SendPing sends a ping message
func (g *Gossip) SendPing(remotePeers []types.Engine) {
	g.log.Info("Sending ping to peer(s)", "NumPeers", len(remotePeers))
	for _, remotePeer := range remotePeers {
		_remotePeer := remotePeer
		go func() {
			if err := g.SendPingToPeer(_remotePeer); err != nil {
				g.PM().HasDisconnected(_remotePeer)
			}
		}()
	}
}

// OnPing processes a Ping message sent by a remote peer.
// It sends back a Pong message containing information
// about the main block chain. It will start
// blockchain synchronization if the Ping message includes
// blockchain information that is better than the local
// blockchain
func (g *Gossip) OnPing(s net.Stream) {

	defer s.Close()

	remotePeer := NewRemoteNode(util.FullRemoteAddressFromStream(s), g.engine)
	remotePeerIDShort := remotePeer.ShortID()

	g.log.Info("Received ping message", "PeerID", remotePeerIDShort)

	// read the message from the stream
	msg := &wire.Ping{}
	if err := ReadStream(s, msg); err != nil {
		g.log.Error("failed to read ping message", "Err", err, "PeerID", remotePeerIDShort)
		return
	}

	// determine the best block an the total
	// difficulty of the block. Add these info
	// to the pong message.
	bestBlock, err := g.GetBlockchain().ChainReader().GetBlock(0)
	if err != nil {
		g.log.Error("Pong failed. Failed to determine best block", "Err", err)
		return
	}

	pongMsg := &wire.Pong{
		BestBlockHash:            bestBlock.GetHash(),
		BestBlockNumber:          bestBlock.GetNumber(),
		BestBlockTotalDifficulty: bestBlock.GetHeader().GetTotalDifficulty(),
	}

	// send pong message
	if err := WriteStream(s, pongMsg); err != nil {
		g.log.Error("failed to send pong response", "Err", err)
		return
	}

	// update the remote peer's timestamp in the peer manager's list
	remotePeer.SetTimestamp(time.Now())
	g.PM().AddOrUpdatePeer(remotePeer)

	g.log.Info("Sent pong response to peer", "PeerID", remotePeerIDShort)

	// compare best chain.
	// If the blockchain best block has a less
	// total difficulty, then need to start the block sync process
	bestBlock, _ = g.GetBlockchain().ChainReader().Current()
	if bestBlock.GetHeader().GetTotalDifficulty().Cmp(msg.BestBlockTotalDifficulty) == -1 {
		g.log.Info("Local blockchain is behind peer",
			"PeerID", remotePeerIDShort,
			"Height", msg.BestBlockNumber,
			"TotalDifficulty", msg.BestBlockTotalDifficulty)
		g.log.Info("Attempting to sync blockchain with peer", "PeerID", remotePeerIDShort)
		go g.SendGetBlockHashes(remotePeer, util.Hash{})
	}

	// Update the current known best
	// remote block information
	var bestBlockInfo BestBlockInfo
	copier.Copy(&bestBlockInfo, msg)
	g.engine.updateSyncInfo(&bestBlockInfo)
}
