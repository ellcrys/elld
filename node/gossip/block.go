package gossip

import (
	"fmt"

	"github.com/ellcrys/elld/util/cache"
	"github.com/jinzhu/copier"

	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/params"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/util"
	net "github.com/libp2p/go-libp2p-net"
)

func makeBlockHistoryKey(hash string, peer core.Engine) []interface{} {
	return []interface{}{"b", hash, peer.StringID()}
}

func makeOrphanBlockHistoryKey(blockHash util.Hash,
	peer core.Engine) []interface{} {
	return []interface{}{"ob", blockHash.HexStr(), peer.StringID()}
}

// RelayBlock sends a given block to remote peers.
// The block is encapsulated in a BlockBody message.
func (g *GossipManager) RelayBlock(block types.Block, remotePeers []core.Engine) error {

	g.log.Debug("Relaying block to peer(s)", "BlockNo", block.GetNumber(),
		"NumPeers", len(remotePeers))

	sent := 0
	for _, peer := range remotePeers {

		historyKey := makeBlockHistoryKey(block.GetHashAsHex(), peer)

		if g.engine.GetHistory().HasMulti(historyKey...) {
			continue
		}

		s, c, err := g.NewStream(peer, config.Versions.BlockBody)
		if err != nil {
			g.logConnectErr(err, peer, "[RelayBlock] Failed to connect to peer")
			continue
		}
		defer c()
		defer s.Close()

		var blockBody core.BlockBody
		copier.Copy(&blockBody, block)
		if err := WriteStream(s, blockBody); err != nil {
			s.Reset()
			g.logErr(err, peer, "[RelayBlock] Failed to write to peer")
			continue
		}

		g.engine.GetHistory().AddMulti(cache.Sec(600), historyKey...)

		sent++
	}

	g.log.Debug("Finished relaying block", "BlockNo",
		block.GetNumber(), "NumPeersSentTo", sent)

	return nil
}

// OnBlockBody handles incoming BlockBody message.
// BlockBody messages contain information about a
// block. It will attempt to process the received
// block.
func (g *GossipManager) OnBlockBody(s net.Stream, rp core.Engine) error {

	defer s.Close()

	blockBody := &core.BlockBody{}
	if err := ReadStream(s, &blockBody); err != nil {
		s.Reset()
		return g.logErr(err, rp, "[OnBlockBody] Failed to read")
	}

	var block core.Block
	copier.Copy(&block, blockBody)
	block.SetBroadcaster(rp)

	g.log.Info("Received a block", "BlockNo", block.GetNumber(),
		"Difficulty", block.GetHeader().GetDifficulty())

	historyKey := makeBlockHistoryKey(block.GetHashAsHex(), rp)
	if g.engine.GetHistory().HasMulti(historyKey...) {
		return nil
	}

	if _, err := g.GetBlockchain().ProcessBlock(&block); err != nil {
		g.engine.GetEventEmitter().Emit(EventBlockProcessed, &block, err)
		return err
	}

	g.engine.GetEventEmitter().Emit(EventBlockProcessed, &block, nil)
	g.engine.GetHistory().AddMulti(cache.Sec(600), historyKey...)

	return nil
}

// RequestBlock sends a RequestBlock message to remote peer.
// A RequestBlock message includes information about a
// specific block. It will attempt to process the requested
// block after receiving it from the remote peer.
// The block's validation context is set to ContextBlockSync
// which cause the transactions to not be required to exist
// in the transaction pool.
func (g *GossipManager) RequestBlock(rp core.Engine, blockHash util.Hash) error {

	historyKey := makeOrphanBlockHistoryKey(blockHash, rp)
	if g.engine.GetHistory().HasMulti(historyKey...) {
		return nil
	}

	s, c, err := g.NewStream(rp, config.Versions.RequestBlock)
	if err != nil {
		return g.logConnectErr(err, rp, "[RequestBlock] Failed to connect to peer")
	}
	defer c()
	defer s.Reset()

	msg := &core.RequestBlock{Hash: blockHash.HexStr()}
	if err := WriteStream(s, msg); err != nil {
		s.Reset()
		return g.logErr(err, rp, "[RequestBlock] Failed to write to peer")
	}

	var blockBody core.BlockBody
	if err := ReadStream(s, &blockBody); err != nil {
		s.Reset()
		return g.logErr(err, rp, "[RequestBlock] Failed to read")
	}

	var block core.Block
	copier.Copy(&block, blockBody)
	block.SetBroadcaster(rp)
	block.SetValidationContexts(types.ContextBlockSync)
	if _, err := g.GetBlockchain().ProcessBlock(&block); err != nil {
		g.log.Debug("Unable to process block", "Err", err)
		g.engine.GetEventEmitter().Emit(EventBlockProcessed, &block, err)
		return err
	}

	g.engine.GetHistory().AddMulti(cache.Sec(600), historyKey...)

	return nil
}

// OnRequestBlock handles RequestBlock message.
// A RequestBlock message includes information
// a bout a block that a remote node needs.
func (g *GossipManager) OnRequestBlock(s net.Stream, rp core.Engine) error {

	defer s.Close()

	msg := &core.RequestBlock{}
	if err := ReadStream(s, msg); err != nil {
		s.Reset()
		return g.logErr(err, rp, "[OnRequestBlock] Failed to read")
	}

	g.log.Debug("Received request for block",
		"RequestedBlockHash", util.StrToHash(msg.Hash).SS())

	if msg.Hash == "" {
		s.Reset()
		err := fmt.Errorf("Invalid RequestBlock message: empty 'Hash' field")
		g.log.Debug(err.Error(), "PeerID", rp.ShortID())
		return err
	}

	var block types.Block

	// decode the hex into a util.Hash
	blockHash, err := util.HexToHash(msg.Hash)
	if err != nil {
		s.Reset()
		g.log.Debug("Invalid hash supplied in requestblock message",
			"PeerID", rp.ShortID(), "Hash", msg.Hash)
		return err
	}

	// find the block
	block, err = g.GetBlockchain().GetBlockByHash(blockHash)
	if err != nil {
		if err != core.ErrBlockNotFound {
			s.Reset()
			g.log.Error(err.Error())
			return err
		}
		s.Reset()
		g.log.Debug("Requested block is not found", "PeerID", rp.ShortID(),
			"Hash", util.StrToHash(msg.Hash).SS())
		return err
	}

	var blockBody core.BlockBody
	copier.Copy(&blockBody, block)
	if err := WriteStream(s, blockBody); err != nil {
		s.Reset()
		g.logErr(err, rp, "[OnRequestBlock] Failed to write")
	}

	return nil
}

// SendGetBlockHashes sends a GetBlockHashes message to
// the remotePeer asking for block hashes beginning from
// a block they share in common. The local peer sends the
// remote peer a list of hashes (locators) while the
// remote peer use the locators to find the highest
// block height they share in common, then it collects
// and sends block hashes after the chosen shared block.
//
// If the locators is not provided via the locator argument,
// they will be collected from the main chain.
func (g *GossipManager) SendGetBlockHashes(rp core.Engine, locators []util.Hash) (*core.BlockHashes, error) {

	rpID := rp.ShortID()
	g.log.Debug("Requesting block headers", "PeerID", rpID)

	s, c, err := g.NewStream(rp, config.Versions.GetBlockHashes)
	if err != nil {
		return nil, g.logConnectErr(err, rp, "[SendGetBlockHashes] Failed to connect")
	}
	defer c()
	defer s.Close()

	if len(locators) == 0 {
		locators, err = g.GetBlockchain().GetLocators()
		if err != nil {
			g.log.Error("failed to get locators", "Err", err)
			return nil, err
		}
	}

	msg := core.GetBlockHashes{
		Locators:  locators,
		MaxBlocks: params.MaxGetBlockHashes,
	}

	if err := WriteStream(s, msg); err != nil {
		return nil, g.logErr(err, rp, "[SendGetBlockHashes] Failed to write")
	}

	g.engine.GetEventEmitter().Emit(EventRequestedBlockHashes,
		msg.Locators, msg.MaxBlocks)

	// Read the return block hashes
	var blockHashes core.BlockHashes
	if err := ReadStream(s, &blockHashes); err != nil {
		return nil, g.logErr(err, rp, "[SendGetBlockHashes] Failed to read")
	}

	g.engine.GetEventEmitter().Emit(EventReceivedBlockHashes)
	g.log.Info("Successfully requested block headers", "PeerID", rpID, "NumLocators",
		len(msg.Locators))

	return &blockHashes, nil
}

// OnGetBlockHashes processes a core.GetBlockHashes request.
// It will attempt to find a chain it shares in common using
// the locator block hashes provided in the message.
//
// If it does not find a chain that is shared with the remote
// chain, it will assume the chains are not off same network
// and as such send an empty block hash response.
//
// If it finds that the remote peer has a chain that is
// not the same as its main chain (a branch), it will
// send block hashes starting from the root parent block (oldest
// ancestor) which exists on the main chain.
func (g *GossipManager) OnGetBlockHashes(s net.Stream, rp core.Engine) error {

	defer s.Close()

	// Read the message
	msg := &core.GetBlockHashes{}
	if err := ReadStream(s, msg); err != nil {
		return g.logErr(err, rp, "[OnGetBlockHashes] Failed to read")
	}

	var blockHashes = core.BlockHashes{}
	var startBlock types.Block
	var blockCursor uint64
	var locatorChain types.ChainReader
	var locatorHash util.Hash

	// Using the provided locator hashes, find a chain
	// where one of the locator block exists. Expects the
	// order of the locator to begin with the highest
	// tip block hash of the remote node
	for _, hash := range msg.Locators {
		locatorChain = g.GetBlockchain().GetChainReaderByHash(hash)
		if locatorChain != nil {
			locatorHash = hash
			break
		}
	}

	// Since we didn't find any common chain,
	// we will assume the node does not share
	// any similarity with the local peer's network
	// as such return nothing
	if locatorChain == nil {
		blockHashes = core.BlockHashes{}
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
		g.log.Debug("Found locator chain root", "HasStartBlock", startBlock != nil)
	} else {
		startBlock, _ = locatorChain.GetBlockByHash(locatorHash)
		g.log.Debug("Found locator block", "HasStartBlock", startBlock != nil)
	}

	// This should only be true when chain tree
	// structure has been corrupted on disk.
	if startBlock == nil {
		g.log.Warn("Could not get the sync start block. " +
			"Possible chain tree corruption.")
		return nil
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
		g.logErr(err, rp, "[OnGetBlockHashes] Failed to write")
		return err
	}

	return nil
}

// SendGetBlockBodies sends a GetBlockBodies message
// requesting for whole bodies of a collection blocks.
func (g *GossipManager) SendGetBlockBodies(rp core.Engine, hashes []util.Hash) (*core.BlockBodies, error) {

	rpID := rp.ShortID()
	g.log.Debug("Requesting block bodies", "PeerID", rpID, "NumHashes", len(hashes))

	s, c, err := g.NewStream(rp, config.Versions.GetBlockBodies)
	if err != nil {
		return nil, g.logConnectErr(err, rp, "[SendGetBlockBodies] Failed to connect")
	}
	defer c()
	defer s.Close()

	// do nothing if no hash is given
	if len(hashes) == 0 {
		return &core.BlockBodies{}, nil
	}

	msg := core.GetBlockBodies{
		Hashes: hashes,
	}

	// write to the stream
	if err := WriteStream(s, msg); err != nil {
		return nil, g.logErr(err, rp, "[SendGetBlockBodies] Failed to write")
	}

	// Read the return block bodies
	var blockBodies core.BlockBodies
	if err := ReadStream(s, &blockBodies); err != nil {
		return nil, g.logErr(err, rp, "[SendGetBlockBodies] Failed to read")
	}

	g.engine.GetEventEmitter().Emit(EventBlockBodiesProcessed)

	return &blockBodies, nil
}

// OnGetBlockBodies handles GetBlockBodies requests
func (g *GossipManager) OnGetBlockBodies(s net.Stream, rp core.Engine) error {
	defer s.Close()

	// Read the message
	msg := &core.GetBlockBodies{}
	if err := ReadStream(s, msg); err != nil {
		return g.logErr(err, rp, "[OnGetBlockBodies] Failed to read")
	}

	var bestChain = g.GetBlockchain().ChainReader()
	var blockBodies = new(core.BlockBodies)
	for _, hash := range msg.Hashes {
		block, err := bestChain.GetBlockByHash(hash)
		if err != nil {
			if err != core.ErrBlockNotFound {
				g.log.Error("Failed fetch block body of a given hash", "Err", err,
					"Hash", hash)
				return err
			}
			continue
		}
		var blockBody core.BlockBody
		copier.Copy(&blockBody, block)
		blockBodies.Blocks = append(blockBodies.Blocks, &blockBody)
	}

	// send the block bodies
	if err := WriteStream(s, blockBodies); err != nil {
		g.logErr(err, rp, "[OnGetBlockBodies] Failed to write")
		return err
	}

	return nil
}
