package node

import (
	"context"
	"fmt"
	"time"

	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/util/logger"
	"github.com/ellcrys/elld/wire"
	net "github.com/libp2p/go-libp2p-net"
)

// createHandshakeMsg creates an Handshake message
func createHandshakeMsg(bestChain core.ChainReader, log logger.Logger) (*wire.Handshake, error) {

	msg := &wire.Handshake{
		SubVersion: config.ClientVersion,
	}

	// determine the best block and the total
	// difficulty of the block. Add these data to
	// the handshake message.
	bestBlock, err := bestChain.GetBlock(0)
	if err != nil {
		log.Error("Handshake failed. Failed to determine best block", "Err", err)
		return nil, fmt.Errorf("handshake failed: failed to determine best block: %s", err)
	}

	msg.BestBlockHash = bestBlock.GetHash()
	msg.BestBlockTotalDifficulty = bestBlock.GetHeader().GetTotalDifficulty()

	return msg, nil
}

// SendHandshake sends an introduction message to a peer
func (g *Gossip) SendHandshake(remotePeer types.Engine) error {

	remotePeerIDShort := remotePeer.ShortID()

	g.log.Info("Sending handshake to peer", "PeerID", remotePeerIDShort)

	// create stream
	s, err := g.newStream(context.Background(), remotePeer, config.HandshakeVersion)
	if err != nil {
		g.log.Debug("Handshake failed. failed to connect to peer", "Err", err, "PeerID", remotePeerIDShort)
		return fmt.Errorf("handshake failed. failed to connect to peer. %s", err)
	}
	defer s.Close()

	engineHandshakeMsg, err := createHandshakeMsg(g.engine.bchain.ChainReader(), g.log)
	if err != nil {
		return err
	}

	// write to the stream
	if err := writeStream(s, engineHandshakeMsg); err != nil {
		s.Reset()
		g.log.Debug("Handshake failed. failed to write to stream", "Err", err, "PeerID", remotePeerIDShort)
		return fmt.Errorf("handshake failed. failed to write to stream")
	}

	g.log.Info("Sent handshake with current main chain state", "PeerID",
		remotePeerIDShort, "SubVersion",
		engineHandshakeMsg.SubVersion, "TotalDifficulty",
		engineHandshakeMsg.BestBlockTotalDifficulty)

	// receive handshake message from the remote peer.
	resp := &wire.Handshake{}
	if err := readStream(s, resp); err != nil {
		s.Reset()
		g.log.Debug("Failed to read handshake response", "Err", err, "PeerID", remotePeerIDShort)
		return fmt.Errorf("failed to read handshake response")
	}

	// update the timestamp of the peer
	remotePeer.SetTimestamp(time.Now())
	g.PM().AddOrUpdatePeer(remotePeer)

	g.log.Info("Handshake was successful", "PeerID", remotePeerIDShort, "SubVersion", resp.SubVersion)
	g.log.Info("Received handshake response", "PeerID",
		remotePeerIDShort, "SubVersion",
		resp.SubVersion, "TotalDifficulty",
		resp.BestBlockTotalDifficulty)

	return nil
}

// OnHandshake handles incoming handshake request
func (g *Gossip) OnHandshake(s net.Stream) {

	defer s.Close()

	remotePeer := NewRemoteNode(util.FullRemoteAddressFromStream(s), g.Engine())
	remotePeerIDShort := remotePeer.ShortID()

	// In non-production mode, ensure wire from public addresses are ignored
	if !g.Engine().ProdMode() && !util.IsDevAddr(remotePeer.IP) {
		s.Reset()
		g.log.Debug("In development mode, we cannot interact with peers with public IP",
			"Addr", remotePeer.GetMultiAddr(), "Msg", "Handshake")
		return
	}

	// read the message from the stream
	remotePeerMsg := &wire.Handshake{}
	if err := readStream(s, remotePeerMsg); err != nil {
		s.Reset()
		g.log.Error("failed to read handshake message", "Err", err, "PeerID", remotePeerIDShort)
		return
	}

	g.log.Info("Received handshake", "PeerID",
		remotePeerIDShort, "SubVersion",
		remotePeerMsg.SubVersion, "TotalDifficulty",
		remotePeerMsg.BestBlockTotalDifficulty)

	engineHandshakeMsg, err := createHandshakeMsg(g.engine.bchain.ChainReader(), g.log)
	if err != nil {
		return
	}

	// send back a Handshake as response
	if err := writeStream(s, engineHandshakeMsg); err != nil {
		s.Reset()
		g.log.Error("failed to send handshake response", "Err", err)
		return
	}

	// update the remote peer's timestamp and add it to the peer manager's list
	remotePeer.Timestamp = time.Now()
	g.PM().AddOrUpdatePeer(remotePeer)

	g.log.Info("Returned handshake with current main chain state", "PeerID",
		remotePeerIDShort, "SubVersion",
		engineHandshakeMsg.SubVersion, "TotalDifficulty",
		engineHandshakeMsg.BestBlockTotalDifficulty)

	// compare best chain.
	// If the blockchain best block has a less
	// total difficulty, then need to start the block sync process
	bestBlock, _ := g.engine.bchain.ChainReader().Current()
	if bestBlock.GetHeader().GetTotalDifficulty().Cmp(remotePeerMsg.BestBlockTotalDifficulty) == -1 {
		go g.SendGetBlockHeaders(remotePeer)
	}
}
