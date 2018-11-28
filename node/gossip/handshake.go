package gossip

import (
	"fmt"

	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/util/cache"
	"github.com/ellcrys/elld/util/logger"
	"github.com/ellcrys/go-ethereum/log"
	"github.com/jinzhu/copier"
	net "github.com/libp2p/go-libp2p-net"
)

// createHandshakeMsg creates an Handshake message
func createHandshakeMsg(msg *core.Handshake, bestChain types.ChainReader,
	slog logger.Logger) (*core.Handshake, error) {

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

// SendHandshake sends an introductory message to a peer
func (g *Gossip) SendHandshake(rp core.Engine) error {

	rpIDShort := rp.ShortID()

	s, c, err := g.NewStream(rp, config.Versions.Handshake)
	if err != nil {
		return g.logConnectErr(err, rp, "[SendHandshake] Failed to connect to peer")
	}
	defer c()
	defer s.Close()

	g.log.Info("Sent handshake to peer", "PeerID", rpIDShort)

	nodeMsg, err := createHandshakeMsg(&core.Handshake{
		Version: g.engine.GetCfg().VersionInfo.BuildVersion,
	}, g.GetBlockchain().ChainReader(), g.log)
	if err != nil {
		return err
	}

	if err := WriteStream(s, nodeMsg); err != nil {
		return g.logErr(err, rp, "[SendHandshake] Failed to write to stream")
	}

	g.log.Info("Handshake sent to peer", "PeerID", rpIDShort, "ClientVersion",
		nodeMsg.Version, "TotalDifficulty",
		nodeMsg.BestBlockTotalDifficulty)

	resp := &core.Handshake{}
	if err := ReadStream(s, resp); err != nil {
		return g.logErr(err, rp, "[SendHandshake] Failed to read from stream")
	}

	// Add or update peer 'last seen' timestamp
	g.PM().AddOrUpdateNode(rp)

	// Set new peer as acquainted so that
	// it will be allowed to send future messages
	g.PM().AddAcquainted(rp)

	// Add remote peer into the intro cache with a TTL of 1 hour.
	g.engine.GetIntros().AddWithExp(rp.StringID(), struct{}{}, cache.Sec(3600))

	g.log.Info("Received handshake response", "PeerID", rpIDShort,
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
			"PeerID", rpIDShort,
			"PeerChainHeight", resp.BestBlockNumber,
			"PeerChainTotalDifficulty", resp.BestBlockTotalDifficulty)
		g.log.Info("Attempting to sync blockchain with peer", "PeerID",
			rpIDShort)
		go g.SendGetBlockHashes(rp, nil)
	}

	// Update the current known best
	// remote block information
	var bestBlockInfo core.BestBlockInfo
	copier.Copy(&bestBlockInfo, resp)
	g.engine.UpdateSyncInfo(&bestBlockInfo)

	return nil
}

// OnHandshake handles incoming handshake requests
func (g *Gossip) OnHandshake(s net.Stream, rp core.Engine) error {

	msg := &core.Handshake{}
	if err := ReadStream(s, msg); err != nil {
		return g.logErr(err, rp, "[OnHandshake] Failed to read message")
	}

	g.log.Info("Received handshake", "PeerID", rp.ShortID(),
		"ClientVersion", msg.Version,
		"Height", msg.BestBlockNumber,
		"TotalDifficulty", msg.BestBlockTotalDifficulty)

	nodeMsg, err := createHandshakeMsg(&core.Handshake{
		Version: g.engine.GetCfg().VersionInfo.BuildVersion,
	}, g.GetBlockchain().
		ChainReader(), g.log)
	if err != nil {
		return err
	}

	// send back a Handshake as response
	if err := WriteStream(s, nodeMsg); err != nil {
		return g.logErr(err, rp, "[OnHandshake] Failed to send response")
	}

	// Set new peer as acquainted so that
	// it will be allowed to send future messages
	g.PM().AddAcquainted(rp)

	// Add or update peer 'last seen' timestamp
	g.PM().AddOrUpdateNode(rp)

	// Set the peer as an inbound connection
	g.PM().GetPeer(rp.StringID()).SetInbound(true)

	// Add remote peer into the intro cache with a TTL of 1 hour.
	g.engine.GetIntros().AddWithExp(rp.StringID(), struct{}{}, cache.Sec(3600))

	g.log.Info("Responded to handshake with chain state", "PeerID",
		rp.ShortID(), "ClientVersion",
		nodeMsg.Version, "TotalDifficulty",
		nodeMsg.BestBlockTotalDifficulty)

	// Compare best chain.
	// If the blockchain best block has a less
	// total difficulty, then we need to start the block sync process
	bestBlock, _ := g.GetBlockchain().ChainReader().Current()
	if bestBlock.GetHeader().GetTotalDifficulty().
		Cmp(msg.BestBlockTotalDifficulty) == -1 {
		g.log.Info("Local blockchain is behind peer",
			"ChainHeight", bestBlock.GetNumber(),
			"TotalDifficulty", bestBlock.GetHeader().GetTotalDifficulty(),
			"PeerID", rp.ShortID(),
			"PeerChainHeight", msg.BestBlockNumber,
			"PeerChainTotalDifficulty", msg.BestBlockTotalDifficulty)
		go g.SendGetBlockHashes(rp, nil)
	}

	// Update the current known best remote block information
	var bestBlockInfo core.BestBlockInfo
	copier.Copy(&bestBlockInfo, msg)
	g.engine.UpdateSyncInfo(&bestBlockInfo)
	return nil
}
