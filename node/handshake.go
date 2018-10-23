package node

import (
	"fmt"

	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/util/cache"
	"github.com/ellcrys/elld/util/logger"
	"github.com/ellcrys/elld/wire"
	"github.com/jinzhu/copier"
	net "github.com/libp2p/go-libp2p-net"
)

// createHandshakeMsg creates an Handshake message
func createHandshakeMsg(bestChain core.ChainReader,
	log logger.Logger) (*wire.Handshake, error) {

	msg := &wire.Handshake{
		Version: config.ClientVersion,
	}

	// determine the best block and the total
	// difficulty of the block. Add these data to
	// the handshake message.
	bestBlock, err := bestChain.GetBlock(0)
	if err != nil {
		log.Error("Handshake failed. Failed to determine "+
			"best block", "Err", err.Error())
		return nil, fmt.Errorf("handshake failed: failed to "+
			"determine best block: %s", err.Error())
	}

	msg.BestBlockHash = bestBlock.GetHash()
	msg.BestBlockTotalDifficulty = bestBlock.GetHeader().GetTotalDifficulty()
	msg.BestBlockNumber = bestBlock.GetNumber()

	return msg, nil
}

func (g *Gossip) logErr(err error, remotePeer types.Engine, msg string) error {
	g.log.Debug(msg, "Err", err, "PeerID", remotePeer.ShortID())
	return err
}

// SendHandshake sends an introductory message to a peer
func (g *Gossip) SendHandshake(remotePeer types.Engine) error {

	remotePeerIDShort := remotePeer.ShortID()

	s, c, err := g.NewStream(remotePeer, config.HandshakeVersion)
	if err != nil {
		return g.logErr(err, remotePeer, "[SendHandshake] Failed to connect to peer")
	}
	defer c()
	defer s.Close()

	g.log.Info("Sent handshake to peer", "PeerID", remotePeerIDShort)

	nodeMsg, err := createHandshakeMsg(g.GetBlockchain().ChainReader(), g.log)
	if err != nil {
		return err
	}

	if err := WriteStream(s, nodeMsg); err != nil {
		return g.logErr(err, remotePeer, "[SendHandshake] Failed to write to stream")
	}

	g.log.Info("Handshake sent to peer", "PeerID", remotePeerIDShort, "ClientVersion",
		nodeMsg.Version, "TotalDifficulty",
		nodeMsg.BestBlockTotalDifficulty)

	resp := &wire.Handshake{}
	if err := ReadStream(s, resp); err != nil {
		return g.logErr(err, remotePeer, "[SendHandshake] Failed to read from stream")
	}

	g.PM().UpdateLastSeenTime(remotePeer)

	// Set the remote peer's acquainted status to
	// true, which will allow some unsolicited
	// messages to be accepted and health check
	// to be performed
	remotePeer.Acquainted()

	g.log.Info("Received handshake response", "PeerID", remotePeerIDShort,
		"ClientVersion", resp.Version, "Height", resp.BestBlockNumber,
		"TotalDifficulty", resp.BestBlockTotalDifficulty)

	// compare best chain.
	// If the blockchain best block has a lesser
	// total difficulty, then need to start the block sync process
	bestBlock, _ := g.GetBlockchain().ChainReader().Current()
	if bestBlock.GetHeader().GetTotalDifficulty().
		Cmp(resp.BestBlockTotalDifficulty) == -1 {
		g.log.Info("Local blockchain is behind peer",
			"ChainHeight", bestBlock.GetNumber(),
			"TotalDifficulty", bestBlock.GetHeader().GetTotalDifficulty(),
			"PeerID", remotePeerIDShort,
			"PeerChainHeight", resp.BestBlockNumber,
			"PeerChainTotalDifficulty", resp.BestBlockTotalDifficulty)
		g.log.Info("Attempting to sync blockchain with peer", "PeerID",
			remotePeerIDShort)
		go g.SendGetBlockHashes(remotePeer, nil)
	}

	// Update the current known best
	// remote block information
	var bestBlockInfo BestBlockInfo
	copier.Copy(&bestBlockInfo, resp)
	g.engine.updateSyncInfo(&bestBlockInfo)

	// Add remote peer into the intro cache with a TTL of 1 hour.
	g.engine.intros.AddWithExp(remotePeer.StringID(), struct{}{}, cache.Sec(3600))

	return nil
}

// OnHandshake handles incoming handshake request
func (g *Gossip) OnHandshake(s net.Stream) {

	remotePeer := NewRemoteNode(util.RemoteAddrFromStream(s), g.Engine())
	remotePeerIDShort := remotePeer.ShortID()

	// In non-production mode, ensure wire from public addresses are ignored
	if !g.Engine().ProdMode() && !util.IsDevAddr(remotePeer.IP) {
		g.log.Debug("In development mode, we cannot interact with "+
			"peers with public IP",
			"Addr", remotePeer.GetAddress(), "Msg", "Handshake")
		return
	}

	// read the message from the stream
	msg := &wire.Handshake{}
	if err := ReadStream(s, msg); err != nil {
		g.logErr(err, remotePeer, "[OnHandshake] Failed to read message")
		return
	}

	g.log.Info("Received handshake", "PeerID", remotePeerIDShort,
		"ClientVersion", msg.Version,
		"Height", msg.BestBlockNumber,
		"TotalDifficulty", msg.BestBlockTotalDifficulty)

	nodeMsg, err := createHandshakeMsg(g.GetBlockchain().
		ChainReader(), g.log)
	if err != nil {
		return
	}

	// send back a Handshake as response
	if err := WriteStream(s, nodeMsg); err != nil {
		g.logErr(err, remotePeer, "[OnHandshake] Failed to send response")
		return
	}

	// Add the remote peer if not previously
	// known and update the last seen time. Also,
	// mark the remote peer as an inbound connection.
	g.PM().UpdateLastSeenTime(remotePeer)
	g.PM().GetPeer(remotePeer.StringID()).SetInbound(true)

	// Set the remote peer's acquainted status to
	// true, which will allow some unsolicited
	// messages to be accepted and health check
	// to be performed
	remotePeer.Acquainted()

	g.log.Info("Responded to handshake with chain state", "PeerID",
		remotePeerIDShort, "ClientVersion",
		nodeMsg.Version, "TotalDifficulty",
		nodeMsg.BestBlockTotalDifficulty)

	// compare best chain.
	// If the blockchain best block has a less
	// total difficulty, then need to start the block sync process
	bestBlock, _ := g.GetBlockchain().ChainReader().Current()
	if bestBlock.GetHeader().GetTotalDifficulty().
		Cmp(msg.BestBlockTotalDifficulty) == -1 {
		g.log.Info("Local blockchain is behind peer",
			"ChainHeight", bestBlock.GetNumber(),
			"TotalDifficulty", bestBlock.GetHeader().GetTotalDifficulty(),
			"PeerID", remotePeerIDShort,
			"PeerChainHeight", msg.BestBlockNumber,
			"PeerChainTotalDifficulty", msg.BestBlockTotalDifficulty)
		go g.SendGetBlockHashes(remotePeer, nil)
	}

	// Update the current known best
	// remote block information
	var bestBlockInfo BestBlockInfo
	copier.Copy(&bestBlockInfo, msg)
	g.engine.updateSyncInfo(&bestBlockInfo)

	// Add remote peer into the intro cache with a TTL of 1 hour.
	g.engine.intros.AddWithExp(remotePeer.StringID(), struct{}{}, cache.Sec(3600))
}
