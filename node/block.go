package node

import (
	"fmt"

	"github.com/ellcrys/elld/types/core/objects"
	"github.com/ellcrys/elld/util/cache"
	"github.com/jinzhu/copier"

	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/params"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/wire"
	net "github.com/libp2p/go-libp2p-net"
)

// BlockHash represents a hash of a block
// sent by a remote peer
type BlockHash struct {
	Hash        util.Hash
	Broadcaster types.Engine
}

func makeBlockHistoryKey(hash string, peer types.Engine) []interface{} {
	return []interface{}{"b", hash, peer.StringID()}
}

func makeOrphanBlockHistoryKey(blockHash util.Hash,
	peer types.Engine) []interface{} {
	return []interface{}{"ob", blockHash.HexStr(), peer.StringID()}
}

// RelayBlock sends a given block
// to remote peers wrapped as the
// only block in a BlockBodies
func (g *Gossip) RelayBlock(block core.Block,
	remotePeers []types.Engine) error {

	g.log.Debug("Relaying block to peer(s)", "BlockNo", block.GetNumber(),
		"NumPeers", len(remotePeers))

	sent := 0
	for _, peer := range remotePeers {

		historyKey := makeBlockHistoryKey(block.HashToHex(), peer)

		// check if we have an history of
		// sending or receiving this block
		// from this remote peer.
		// If yes, do not relay
		if g.engine.history.HasMulti(historyKey...) {
			continue
		}

		s, c, err := g.NewStream(peer, config.BlockBodyVersion)
		if err != nil {
			g.log.Error("Block message failed. failed to connect to peer",
				"Err", err, "PeerID", peer.ShortID())
			continue
		}
		defer c()
		defer s.Close()

		var blockBody wire.BlockBody
		copier.Copy(&blockBody, block)
		if err := WriteStream(s, blockBody); err != nil {
			s.Reset()
			g.log.Error("Block message failed. failed to write to stream",
				"Err", err, "PeerID", peer.ShortID())
			continue
		}

		g.PM().UpdatePeerTime(peer)
		g.engine.history.AddMulti(cache.Sec(600), historyKey...)

		sent++
	}

	g.log.Info("Finished relaying block", "BlockNo",
		block.GetNumber(), "NumPeersSentTo", sent)

	return nil
}

// OnBlockBody handles incoming block message
func (g *Gossip) OnBlockBody(s net.Stream) {

	defer s.Close()
	remotePeer := NewRemoteNode(util.RemoteAddressFromStream(s), g.engine)
	rpID := remotePeer.ShortID()

	blockBody := &wire.BlockBody{}
	if err := ReadStream(s, &blockBody); err != nil {
		s.Reset()
		g.log.Error("Failed to read block message",
			"Err", err, "PeerID", rpID)
		return
	}

	g.PM().UpdatePeerTime(remotePeer)

	var block objects.Block
	copier.Copy(&block, blockBody)
	block.SetBroadcaster(remotePeer)

	g.log.Info("Received a block",
		"BlockNo", block.GetNumber(),
		"Difficulty", block.GetHeader().GetDifficulty())

	// check if we have an history about this
	// block with the remote peer, if no,
	// process the block.
	historyKey := makeBlockHistoryKey(block.HashToHex(), remotePeer)
	if g.engine.history.HasMulti(historyKey...) {
		return
	}

	// Add the transaction to the transaction
	// pool and wait for error response
	if _, err := g.GetBlockchain().ProcessBlock(&block); err != nil {
		g.engine.event.Emit(EventBlockProcessed, &block, err)
		return
	}

	g.engine.event.Emit(EventBlockProcessed, &block, nil)

	// add transaction to the history cache
	// using the key we created earlier
	g.engine.history.AddMulti(cache.Sec(600), historyKey...)
}

// RequestBlock sends a RequestBlock
// message to remotePeer
func (g *Gossip) RequestBlock(remotePeer types.Engine,
	blockHash util.Hash) error {

	historyKey := makeOrphanBlockHistoryKey(blockHash, remotePeer)

	// check if we have an history of sending
	// or receiving this request from this
	// remote peer. If yes, do not relay
	if g.engine.history.HasMulti(historyKey...) {
		return nil
	}

	s, c, err := g.NewStream(remotePeer, config.RequestBlockVersion)
	if err != nil {
		g.log.Error("RequestBlock message failed. failed to connect to peer",
			"Err", err, "PeerID", remotePeer.ShortID())
		return err
	}
	defer c()
	defer s.Close()

	if err := WriteStream(s, &wire.RequestBlock{
		Hash: blockHash.HexStr(),
	}); err != nil {
		s.Reset()
		g.log.Error("RequestBlock message failed. failed to write to stream",
			"Err", err, "PeerID", remotePeer.ShortID())
		return err
	}

	g.PM().UpdatePeerTime(remotePeer)
	g.engine.history.AddMulti(cache.Sec(600), historyKey...)

	return nil
}

// OnRequestBlock handles RequestBlock message
func (g *Gossip) OnRequestBlock(s net.Stream) {

	defer s.Close()
	remotePeer := NewRemoteNode(util.RemoteAddressFromStream(s), g.engine)
	rpID := remotePeer.ShortID()

	// read the message
	requestBlock := &wire.RequestBlock{}
	if err := ReadStream(s, requestBlock); err != nil {
		s.Reset()
		g.log.Error("Failed to read requestblock message",
			"Err", err, "PeerID", rpID)
		return
	}

	g.PM().UpdatePeerTime(remotePeer)
	g.log.Debug("Received request for block",
		"RequestedBlockHash", util.StrToHash(requestBlock.Hash).SS())

	// If the hash and number fields
	// are not set, do not proceed
	// and log error
	if requestBlock.Hash == "" && requestBlock.Number == 0 {
		s.Reset()
		g.log.Warn("Invalid requestblock message: "+
			"Empty 'Hash' and 'Number' fields", "PeerID", rpID)
		return
	}

	// The hash field is mandatory
	if requestBlock.Hash == "" {
		s.Reset()
		g.log.Warn("Invalid requestblock message: "+
			"Empty 'Hash'", "PeerID", rpID)
		return
	}

	var block core.Block
	if requestBlock.Hash != "" && requestBlock.Number > 0 {

		// decode the hex into a util.Hash
		blockHash, err := util.HexToHash(requestBlock.Hash)
		if err != nil {
			s.Reset()
			g.log.Warn("Invalid hash supplied in "+
				"requestblock message",
				"PeerID", rpID, "Hash",
				util.StrToHash(requestBlock.Hash).SS())
			return
		}

		// find the block by number and hash
		block, err = g.GetBlockchain().GetBlock(requestBlock.Number,
			blockHash)
		if err != nil {
			if err != core.ErrBlockNotFound {
				s.Reset()
				g.log.Warn("Failed to find block described "+
					"in requestblock message",
					"PeerID", rpID,
					"Hash", util.StrToHash(requestBlock.Hash).SS())
				return
			}
			s.Reset()
			g.log.Debug("Block is currently unknown",
				"PeerID", rpID,
				"Hash", util.StrToHash(requestBlock.Hash).SS())
			return
		}
	}

	// If a block number is not provided,
	// then we need to find the block by
	// just the hash
	if requestBlock.Number == 0 {

		// decode the hex into a util.Hash
		blockHash, err := util.HexToHash(requestBlock.Hash)
		if err != nil {
			s.Reset()
			g.log.Debug("Invalid hash supplied in "+
				"requestblock message",
				"PeerID", rpID, "Hash", requestBlock.Hash)
			return
		}

		// find the block by number and hash
		block, err = g.GetBlockchain().GetBlockByHash(blockHash)
		if err != nil {
			if err != core.ErrBlockNotFound {
				s.Reset()
				g.log.Debug("Failed to find block described "+
					"in requestblock message",
					"PeerID", rpID, "Hash", requestBlock.Hash)
				return
			}
			s.Reset()
			g.log.Debug("Block is currently unknown",
				"PeerID", rpID,
				"Hash", util.StrToHash(requestBlock.Hash).SS())
			return
		}
	}

	// relay the block to only the remote peer
	if err := g.RelayBlock(block, []types.Engine{remotePeer}); err != nil {
		s.Reset()
		g.log.Error("Failed to relay block requested "+
			"in requestblock message",
			"PeerID", rpID,
			"Hash", requestBlock.Hash)
		return
	}
}

// SendGetBlockHashes sends a GetBlockHashes message to
// the remotePeer asking for block hashes beginning from
// a block they share in common. The local peer sends the
// remote peer a list of hashes (locators) on its main chain
// while the remote peer uses the locators to find the highest
// block they share in common, then it collects and sends
// block hashes after the selected shared block.
//
// If the locators is not provided via the locator argument,
// they will be collected from the main chain.
func (g *Gossip) SendGetBlockHashes(remotePeer types.Engine,
	locators []util.Hash) error {

	rpID := remotePeer.ShortID()
	g.log.Debug("Requesting block headers", "PeerID", rpID)

	s, c, err := g.NewStream(remotePeer, config.GetBlockHashesVersion)
	if err != nil {
		g.log.Error("GetBlockHashes message failed. Failed "+
			"to connect to peer", "Err", err, "PeerID", rpID)
		return err
	}
	defer c()
	defer s.Close()

	g.engine.setSyncing(true)

	if len(locators) == 0 {
		locators, err = g.GetBlockchain().GetLocators()
		if err != nil {
			g.log.Error("GetBlockHashes message failed. "+
				"Failed to gather locators", "Err", err)
			return err
		}
	}

	msg := wire.GetBlockHashes{
		Locators:  locators,
		MaxBlocks: params.MaxGetBlockHeader,
	}

	if err := WriteStream(s, msg); err != nil {
		errMsg := fmt.Errorf("GetBlockHashes message failed. " +
			"Failed to write to stream")
		g.log.Error(errMsg.Error(), "Err", err, "PeerID", rpID)
		return fmt.Errorf("%s: %s", errMsg, err)
	}

	g.engine.event.Emit(EventRequestedBlockHashes,
		msg.Locators, msg.MaxBlocks)

	// Read the return block hashes
	var blockHashes wire.BlockHashes
	if err := ReadStream(s, &blockHashes); err != nil {
		errMsg := fmt.Errorf("GetBlockHashes message failed. " +
			"Failed to read stream")
		g.log.Error(errMsg.Error(), "Err", err, "PeerID", rpID)
		return fmt.Errorf("%s: %s", errMsg, err)
	}

	g.PM().UpdatePeerTime(remotePeer)

	// add all the hashes to the hash queue
	for _, h := range blockHashes.Hashes {
		g.engine.blockHashQueue.Append(&BlockHash{
			Hash:        h,
			Broadcaster: remotePeer,
		})
	}

	g.engine.event.Emit(EventReceivedBlockHashes)
	g.log.Info("Successfully requested block headers",
		"PeerID", rpID, "NumLocators", len(msg.Locators))

	return nil
}

// OnGetBlockHashes processes a wire.GetBlockHashes request.
// It will attempt to find a chain it shares in common using
// the locator block hashes provided in the message.
//
// If it does not find a chain that is shared with the remote
// chain, it will assume the chains are not off same network
// and as such send an empty block hash response.
//
// If it finds that the remote peer has a chain that is
// not the same as its main chain (a side branch), it will
// send block hashes starting from the root parent block (oldest
// ancestor) which exists on the main chain.
func (g *Gossip) OnGetBlockHashes(s net.Stream) {
	defer s.Close()
	remotePeer := NewRemoteNode(util.RemoteAddressFromStream(s), g.engine)
	rpID := remotePeer.ShortID()

	// Read the message
	msg := &wire.GetBlockHashes{}
	if err := ReadStream(s, msg); err != nil {
		g.log.Error("Failed to read block message",
			"Err", err, "PeerID", rpID)
		return
	}

	g.PM().UpdatePeerTime(remotePeer)

	var blockHashes = wire.BlockHashes{}
	var startBlock core.Block
	var blockCursor uint64
	var locatorChain core.ChainReader
	var locatorHash util.Hash

	// Using the provided locator hashes, find a chain
	// where one of the locator block exists. Expects the
	// order of the locator to begin with the highest
	// tip block hash of the remote node
	for _, hash := range msg.Locators {
		locatorChain = g.GetBlockchain().GetChainReaderByHash(hash)
		locatorHash = hash
		if locatorChain != nil {
			break
		}
	}

	// Since we didn't find any common chain,
	// we will assume the node does not share
	// any similarity with the local peer's network
	// as such return nothing
	if locatorChain == nil {
		blockHashes = wire.BlockHashes{}
		goto send
	}

	// Check whether the locator's chain is the main
	// chain. If it is not, we need to get the root
	// parent block from which the chain (and its parent)
	// sprouted from. Otherwise, get the locator block
	// and use as the start block.
	if mainChain := g.GetBlockchain().GetBestChain(); mainChain.GetID() !=
		locatorChain.GetID() {
		startBlock = locatorChain.GetRoot()
	} else {
		startBlock, _ = locatorChain.GetBlockByHash(locatorHash)
	}

	// Fetch block hashes starting from the block
	// after the start block
	blockCursor = startBlock.GetNumber() + 1
	for int64(len(blockHashes.Hashes)) <= msg.MaxBlocks {
		block, err := g.GetBlockchain().ChainReader().GetBlock(blockCursor)
		if err != nil {
			if err != core.ErrBlockNotFound {
				g.log.Error("Failed to fetch block header", "Err", err)
			}
			break
		}
		blockHashes.Hashes = append(blockHashes.Hashes, block.GetHash())
		blockCursor++
	}

send:
	if err := WriteStream(s, blockHashes); err != nil {
		errMsg := fmt.Errorf("Failed to write BlockHeader message to stream")
		g.log.Error(errMsg.Error(), "Err", err, "PeerID", rpID)
		return
	}
}

// SendGetBlockBodies sends a GetBlockBodies message
// requesting for whole bodies of a collection blocks.
func (g *Gossip) SendGetBlockBodies(remotePeer types.Engine,
	hashes []util.Hash) error {

	rpID := remotePeer.ShortID()
	g.log.Debug("Requesting block bodies",
		"PeerID", rpID,
		"NumHashes", len(hashes))

	s, c, err := g.NewStream(remotePeer, config.GetBlockBodiesVersion)
	if err != nil {
		g.log.Error("GetBlockBodies message failed. "+
			"Failed to connect to peer", "Err", err, "PeerID", rpID)
		return fmt.Errorf("GetBlockBodies message failed. "+
			"Failed to connect to peer: %s", err)
	}
	defer c()
	defer s.Close()

	// do nothing if no hash is given
	if len(hashes) == 0 {
		return nil
	}

	msg := wire.GetBlockBodies{
		Hashes: hashes,
	}

	// write to the stream
	if err := WriteStream(s, msg); err != nil {
		g.log.Error("GetBlockBodies message failed. "+
			"Failed to write to stream", "Err", err, "PeerID", rpID)
		return fmt.Errorf("GetBlockBodies message failed. "+
			"Failed to write to stream: %s", err)
	}

	// Read the return block bodies
	var blockBodies wire.BlockBodies
	if err := ReadStream(s, &blockBodies); err != nil {
		g.log.Error("Unable to retrieve BlockBodies. "+
			"Failed to read stream", "Err", err, "PeerID", rpID)
		return fmt.Errorf("Unable to retrieve BlockBodies. "+
			"Failed to read stream: %s", err)
	}

	g.PM().UpdatePeerTime(remotePeer)
	g.log.Info("Received block bodies",
		"NumBlocks", len(blockBodies.Blocks))

	// attempt to append the blocks to the blockchain
	for _, bb := range blockBodies.Blocks {
		var block objects.Block
		copier.Copy(&block, bb)

		// Add an history that prevents other routines from
		// relaying this same block to the remote peer.
		historyKey := makeBlockHistoryKey(block.HashToHex(), remotePeer)
		g.engine.history.AddMulti(cache.Sec(600), historyKey...)

		// set the broadcaster and process the block
		block.SetBroadcaster(remotePeer)
		_, err := g.GetBlockchain().ProcessBlock(&block)
		if err != nil {
			g.log.Debug("Unable to append block", "Err", err)
			g.engine.event.Emit(EventBlockProcessed, &block, err)
		} else {
			g.engine.event.Emit(EventBlockProcessed, &block, nil)
		}
	}

	// get sync status
	syncStatus := g.engine.getSyncStateInfo()
	if syncStatus != nil {
		g.log.Info("Current synchronization status",
			"TargetTD", syncStatus.TargetTD,
			"CurTD", syncStatus.CurrentTD,
			"TargetChainHeight", syncStatus.TargetChainHeight,
			"CurChainHeight", syncStatus.CurrentChainHeight,
			"Progress(%)", syncStatus.ProgressPercent)
	}

	// Update the sync status
	g.engine.updateSyncInfo(nil)

	g.engine.event.Emit(EventBlockBodiesProcessed)

	return nil
}

// OnGetBlockBodies handles GetBlockBodies requests
func (g *Gossip) OnGetBlockBodies(s net.Stream) {
	remotePeer := NewRemoteNode(util.RemoteAddressFromStream(s), g.engine)
	rpID := remotePeer.ShortID()

	// Read the message
	msg := &wire.GetBlockBodies{}
	if err := ReadStream(s, msg); err != nil {
		g.log.Error("Failed to read block message", "Err", err, "PeerID", rpID)
		return
	}

	g.PM().UpdatePeerTime(remotePeer)

	var bestChain = g.GetBlockchain().ChainReader()
	var blockBodies = new(wire.BlockBodies)
	for _, hash := range msg.Hashes {
		block, err := bestChain.GetBlockByHash(hash)
		if err != nil {
			if err != core.ErrBlockNotFound {
				g.log.Error("Failed fetch block body of a given hash",
					"Err", err, "Hash", hash)
				return
			}
			continue
		}
		var blockBody wire.BlockBody
		copier.Copy(&blockBody, block)
		blockBodies.Blocks = append(blockBodies.Blocks, &blockBody)
	}

	// send the block bodies
	if err := WriteStream(s, blockBodies); err != nil {
		errMsg := fmt.Errorf("Failed to write BlockBodies message to stream")
		g.log.Error(errMsg.Error(), "Err", err, "PeerID", rpID)
		return
	}
}
