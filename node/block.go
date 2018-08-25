package node

import (
	"context"
	"fmt"

	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/node/histcache"
	"github.com/ellcrys/elld/params"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/types/core/objects"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/wire"
	"github.com/k0kubun/pp"
	net "github.com/libp2p/go-libp2p-net"
)

func makeBlockHistoryKey(block core.Block, peer types.Engine) histcache.MultiKey {
	return []interface{}{"b", block.HashToHex(), peer.StringID()}
}

func makeOrphanBlockHistoryKey(blockHash util.Hash, peer types.Engine) histcache.MultiKey {
	return []interface{}{"ob", blockHash.HexStr(), peer.StringID()}
}

// RelayBlock sends a given block to remote peers.
func (g *Gossip) RelayBlock(block core.Block, remotePeers []types.Engine) error {

	g.log.Debug("Relaying block to peer(s)", "BlockNo", block.GetNumber(), "NumPeers", len(remotePeers))

	sent := 0
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
			g.log.Error("Block message failed. failed to connect to peer", "Err", err, "PeerID", peer.ShortID())
			continue
		}
		defer s.Close()

		// write to the stream
		if err := writeStream(s, block); err != nil {
			s.Reset()
			g.log.Error("Block message failed. failed to write to stream", "Err", err, "PeerID", peer.ShortID())
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
	rpID := remotePeer.ShortID()

	// read the message
	block := &objects.Block{}
	if err := readStream(s, block); err != nil {
		s.Reset()
		g.log.Error("Failed to read block message", "Err", err, "PeerID", rpID)
		return
	}

	// set the broadcaster
	block.SetBroadcaster(remotePeer)

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

// RequestBlock sends a RequestBlock message to remotePeer
func (g *Gossip) RequestBlock(remotePeer types.Engine, blockHash util.Hash) error {

	historyKey := makeOrphanBlockHistoryKey(blockHash, remotePeer)

	// check if we have an history of sending or receiving this request
	// from this remote peer. If yes, do not relay
	if g.engine.History().Has(historyKey) {
		return nil
	}

	// create a stream to the remote peer
	s, err := g.newStream(context.Background(), remotePeer, config.RequestBlockVersion)
	if err != nil {
		g.log.Error("RequestBlock message failed. failed to connect to peer", "Err", err, "PeerID", remotePeer.ShortID())
		return err
	}
	defer s.Close()

	// write to the stream
	if err := writeStream(s, &wire.RequestBlock{
		Hash: blockHash.HexStr(),
	}); err != nil {
		s.Reset()
		g.log.Error("RequestBlock message failed. failed to write to stream", "Err", err, "PeerID", remotePeer.ShortID())
		return err
	}

	// add new history
	g.engine.History().Add(historyKey)

	return nil
}

// OnRequestBlock handles RequestBlock message
func (g *Gossip) OnRequestBlock(s net.Stream) {

	defer s.Close()
	remotePeer := NewRemoteNode(util.FullRemoteAddressFromStream(s), g.engine)
	rpID := remotePeer.ShortID()

	// read the message
	requestBlock := &wire.RequestBlock{}
	if err := readStream(s, requestBlock); err != nil {
		s.Reset()
		g.log.Error("Failed to read requestblock message", "Err", err, "PeerID", rpID)
		return
	}

	g.log.Debug("Received request for block", "RequestedBlockHash", requestBlock.Hash)

	// If the hash and number fields are not set,
	// do not proceed and log error
	if requestBlock.Hash == "" && requestBlock.Number == 0 {
		s.Reset()
		g.log.Warn("Invalid requestblock message: Empty 'Hash' and 'Number' fields", "PeerID", rpID)
		return
	}

	// The hash field is mandatory
	if requestBlock.Hash == "" {
		s.Reset()
		g.log.Warn("Invalid requestblock message: Empty 'Hash'", "PeerID", rpID)
		return
	}

	var block core.Block

	if requestBlock.Hash != "" && requestBlock.Number > 0 {

		// decode the hex into a util.Hash
		blockHash, err := util.HexToHash(requestBlock.Hash)
		if err != nil {
			s.Reset()
			g.log.Warn("Invalid hash supplied in requestblock message", "PeerID", rpID, "Hash", requestBlock.Hash)
			return
		}

		// find the block by number and hash
		block, err = g.engine.bchain.GetBlock(requestBlock.Number, blockHash)
		if err != nil {
			if err != core.ErrBlockNotFound {
				s.Reset()
				g.log.Warn("Failed to find block described in requestblock message", "PeerID", rpID, "Hash", requestBlock.Hash)
				return
			}
			s.Reset()
			g.log.Debug("Block is currently unknown", "PeerID", rpID, "Hash", requestBlock.Hash)
			return
		}
	}

	// If a block number is not provided, then we need to
	// find the block by just the hash
	if requestBlock.Number == 0 {

		// decode the hex into a util.Hash
		blockHash, err := util.HexToHash(requestBlock.Hash)
		if err != nil {
			s.Reset()
			g.log.Warn("Invalid hash supplied in requestblock message", "PeerID", rpID, "Hash", requestBlock.Hash)
			return
		}

		// find the block by number and hash
		block, err = g.engine.bchain.GetBlockByHash(blockHash)
		if err != nil {
			if err != core.ErrBlockNotFound {
				s.Reset()
				g.log.Warn("Failed to find block described in requestblock message", "PeerID", rpID, "Hash", requestBlock.Hash)
				return
			}
			s.Reset()
			g.log.Debug("Block is currently unknown", "PeerID", rpID, "Hash", requestBlock.Hash)
			return
		}
	}

	// relay the block to only the remote peer
	if err := g.RelayBlock(block, []types.Engine{remotePeer}); err != nil {
		s.Reset()
		g.log.Error("Failed to relay block requested in requestblock message", "PeerID", rpID, "Hash", requestBlock.Hash)
		return
	}
}

// SendGetBlockHeaders sends a GetBlockHeaders message to
// remotePeer asking for headers of blocks after the provided
// hash.
func (g *Gossip) SendGetBlockHeaders(remotePeer types.Engine) error {

	rpID := remotePeer.ShortID()
	g.log.Debug("Requesting block headers", "PeerID", rpID)

	// create a stream to the remote peer
	s, err := g.newStream(context.Background(), remotePeer, config.GetBlockHeaders)
	if err != nil {
		errMsg := fmt.Errorf("GetBlockHeaders message failed. Failed to connect to peer")
		g.log.Error(errMsg.Error(), "Err", err, "PeerID", rpID)
		return fmt.Errorf("%s: %s", errMsg, err)
	}
	defer s.Close()

	bestBlock, _ := g.engine.bchain.ChainReader().Current()
	msg := wire.GetBlockHeaders{
		Hash:      bestBlock.GetHash(),
		MaxBlocks: params.MaxGetBlockHeader,
	}

	// write to the stream
	if err := writeStream(s, msg); err != nil {
		s.Reset()
		errMsg := fmt.Errorf("GetBlockHeaders message failed. Failed to write to stream")
		g.log.Error(errMsg.Error(), "Err", err, "PeerID", rpID)
		return fmt.Errorf("%s: %s", errMsg, err)
	}

	g.log.Info("Successfully requested block headers", "PeerID", rpID, "Locator", msg.Hash)

	return nil
}

// OnGetBlockHeaders processes a wire.GetBlockHeaders request.
// It will find the given locator hash in its main chain
// and return headers of subsequent blocks after the locator up
// to the maximum block limit specified.
func (g *Gossip) OnGetBlockHeaders(s net.Stream) {

	defer s.Close()
	remotePeer := NewRemoteNode(util.FullRemoteAddressFromStream(s), g.engine)
	rpID := remotePeer.ShortID()

	// read the message
	msg := &wire.GetBlockHeaders{}
	if err := readStream(s, msg); err != nil {
		s.Reset()
		g.log.Error("Failed to read block message", "Err", err, "PeerID", rpID)
		return
	}

	// get a chain reader to the chain
	// where the locator block hash exist in.
	// If we are unable to find the locator in any
	// known chain, we send an empty BlockHashes message
	locatorChain := g.engine.bchain.GetChainReaderByHash(msg.Hash)
	if locatorChain == nil {
		if err := writeStream(s, msg); err != nil {
			s.Reset()
			errMsg := fmt.Errorf("GetBlockHeaders message failed. Failed to write to stream")
			g.log.Error(errMsg.Error(), "Err", err, "PeerID", rpID)
			// return fmt.Errorf("%s: %s", errMsg, err)
			return
		}
	}

	pp.Println(msg)
}
