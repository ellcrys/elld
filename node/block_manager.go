package node

import (
	"fmt"
	"sync"
	"time"

	"github.com/ellcrys/elld/params"

	"github.com/ellcrys/elld/node/common"

	"github.com/ellcrys/elld/util/cache"

	"gopkg.in/oleiade/lane.v1"

	"github.com/ellcrys/elld/miner"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/util/logger"
	"github.com/jinzhu/copier"
	"github.com/olebedev/emitter"
	"github.com/shopspring/decimal"
)

// unprocessedBlock represents a block the requires processing.
type unprocessedBlock struct {

	// block is the block that needs to be processed
	block types.Block

	// done is a channel to send the processing error status
	done chan error
}

// BlockManager is responsible for handling
// incoming, mined or processed blocks in a
// concurrency safe way.
type BlockManager struct {

	// evt is the global event emitter
	evt *emitter.Emitter

	// syncMtx is a mutex used during block sync.
	syncMtx *sync.RWMutex

	// syncing indicates that block syncing is in progress
	syncing bool

	// log is the logger used by this module
	log logger.Logger

	// bChain is the blockchain manager
	bChain types.Blockchain

	// miner is CPU miner
	miner *miner.Miner

	// engine is the node's instance
	engine *Node

	// syncCandidate are candidate peers to
	// sync blocks with.
	syncCandidate map[string]*types.SyncPeerChainInfo

	// bestSyncCandidate is the current best sync
	// candidate to perform block synchronization with.
	bestSyncCandidate *types.SyncPeerChainInfo

	// unprocessedBlocks hold blocks that are yet to be processed
	unprocessedBlocks *lane.Deque

	// processedBlocks hold blocks that have been processed
	processedBlocks *lane.Deque

	// mined holds the hash of blocks mined by the client
	mined *cache.Cache
}

// NewBlockManager creates a new BlockManager
func NewBlockManager(node *Node) *BlockManager {
	bm := &BlockManager{
		syncMtx:           &sync.RWMutex{},
		log:               node.log,
		evt:               node.event,
		bChain:            node.bChain,
		engine:            node,
		unprocessedBlocks: lane.NewDeque(),
		processedBlocks:   lane.NewDeque(),
		mined:             cache.NewCache(100),
		syncCandidate:     make(map[string]*types.SyncPeerChainInfo),
	}
	return bm
}

// SetMiner sets a reference of the CPU miner
func (bm *BlockManager) SetMiner(m *miner.Miner) {
	bm.miner = m
}

// Manage handles all incoming block related events.
func (bm *BlockManager) Manage() {

	go func() {
		for evt := range bm.evt.On(core.EventFoundBlock) {
			b := evt.Args[0].(*miner.FoundBlock)
			bm.mined.Add(b.Block.GetHash().HexStr(), struct{}{})
			errCh := evt.Args[1].(chan error)
			bm.unprocessedBlocks.Append(&unprocessedBlock{
				block: b.Block,
				done:  errCh,
			})
		}
	}()

	go func() {
		for evt := range bm.evt.On(core.EventNewBlock) {
			bm.processedBlocks.Append(evt.Args[0].(*core.Block))
		}
	}()

	go func() {
		for evt := range bm.evt.On(core.EventOrphanBlock) {
			bm.handleOrphan(evt.Args[0].(*core.Block))
		}
	}()

	go func() {
		for evt := range bm.evt.On(core.EventProcessBlock) {
			bm.unprocessedBlocks.Append(&unprocessedBlock{
				block: evt.Args[0].(*core.Block),
			})
		}
	}()

	go func() {
		for evt := range bm.evt.On(core.EventPeerChainInfo) {
			peerChainInfo := evt.Args[0].(*types.SyncPeerChainInfo)
			if bm.isSyncCandidate(peerChainInfo) {
				bm.addSyncCandidate(peerChainInfo)
				if bm.sync() == nil {
					bm.log.Info("Block synchronization complete")
				}
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(params.QueueProcessorInterval)
		for {
			select {
			case <-ticker.C:
				for !bm.processedBlocks.Empty() {
					bm.handleProcessedBlocks()
				}
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(params.QueueProcessorInterval)
		for {
			select {
			case <-ticker.C:
				for !bm.unprocessedBlocks.Empty() {
					bm.handleUnprocessedBlocks()
				}
			}
		}
	}()
}

// handleProcessedBlocks fetches blocks from the processed
// block queue to perform post-append operations such as
// updating clearing processed transactions from the pool,
// restarting miners and relaying the processed block.
func (bm *BlockManager) handleProcessedBlocks() error {

	var b = bm.processedBlocks.Shift()
	if b == nil {
		return nil
	}

	// Remove the blocks transactions from the pool.
	bm.engine.txsPool.Remove(b.(*core.Block).GetTransactions()...)

	// Restart miner workers if some block was created
	// by another peer. The miner typically manages its
	// restart when it finds a block. But when a remote
	// peer finds a block, we need to force the miner to restart.
	if !bm.mined.Has(b.(*core.Block).GetHash().HexStr()) {
		bm.miner.RestartWorkers()
	}

	// Relay the block to peers.
	if b.(*core.Block).GetNumber() > 1 {
		bm.engine.Gossip().BroadcastBlock(b.(*core.Block),
			bm.engine.PM().GetConnectedPeers())
	}

	return nil
}

// handleUnprocessedBlocks fetches unprocessed blocks
// from the unprocessed block queue and attempts to
// append them to the blockchain.
func (bm *BlockManager) handleUnprocessedBlocks() error {

	var upb = bm.unprocessedBlocks.Shift()
	if upb == nil {
		return nil
	}

	b := upb.(*unprocessedBlock).block
	errCh := upb.(*unprocessedBlock).done
	_, err := bm.bChain.ProcessBlock(b)
	if err != nil {
		go bm.evt.Emit(core.EventBlockProcessed, b, err)
		bm.log.Debug("Failed to process block", "Err", err.Error())

		if errCh != nil {
			errCh <- err
		}

		return err
	}

	go bm.evt.Emit(core.EventBlockProcessed, b, nil)
	bm.log.Info("Block has been processed",
		"BlockNo", b.GetNumber(),
		"BlockHash", b.GetHash().SS())

	if errCh != nil {
		errCh <- err
	}

	return nil
}

// handleOrphan sends a RequestBlock message to
// the originator of an orphaned block.
func (bm *BlockManager) handleOrphan(b *core.Block) {
	parentHash := b.GetHeader().GetParentHash()
	bm.log.Debug("Requesting orphan parent block from broadcaster",
		"BlockNo", b.GetNumber(),
		"ParentBlockHash", parentHash.SS())
	bm.engine.gossipMgr.RequestBlock(b.Broadcaster, parentHash)
}

// isSyncCandidate checks whether a peer is a
// valid sync candidate based on its chain state
// information. A peer is a valid candidate if
// the total difficulty of its best block is
// greater that of the local best block.
func (bm *BlockManager) isSyncCandidate(info *types.SyncPeerChainInfo) bool {
	localBestBlock, _ := bm.engine.GetBlockchain().ChainReader().Current()
	if localBestBlock.GetHeader().GetTotalDifficulty().Cmp(info.PeerChainTD) == -1 {
		bm.log.Info("Local blockchain is behind peer",
			"ChainHeight", localBestBlock.GetNumber(),
			"TotalDifficulty", localBestBlock.GetHeader().GetTotalDifficulty(),
			"PeerID", info.PeerIDShort,
			"PeerChainHeight", info.PeerChainHeight,
			"PeerChainTotalDifficulty", info.PeerChainTD)
		return true
	}
	return false
}

// pickBestSyncCandidate returns the best block synchronization
// candidate. The best candidate is the one with the highest
// total difficulty.
// Note: Not thread safe.
func (bm *BlockManager) pickBestSyncCandidate() *types.SyncPeerChainInfo {
	var bestCandidate *types.SyncPeerChainInfo
	for _, candidate := range bm.syncCandidate {
		if bestCandidate == nil {
			bestCandidate = candidate
			continue
		}
		if bestCandidate.PeerChainTD.
			Cmp(candidate.PeerChainTD) == -1 {
			bestCandidate = candidate
		}
	}
	return bestCandidate
}

// sync starts sync sessions with the available candidates
// starting with the best candidate. The best candidate is
// the one with the highest total difficulty. It continues
// to sync with the best candidate until it is completely
// in sync with it or fails to connect to it.
//
// If there is a failure in connection or a failure in
// requesting for sync objects, the candidate is removed
// and synchronization is restarted.
func (bm *BlockManager) sync() error {

	var blockBodies *core.BlockBodies
	var blockHashes *core.BlockHashes
	var syncStatus *core.SyncStateInfo
	var err error

	if bm.IsSyncing() {
		return fmt.Errorf("syncing")
	}

	bm.syncMtx.Lock()
	if len(bm.syncCandidate) == 0 {
		bm.syncMtx.Unlock()
		return nil
	}

	bm.syncing = true

	// Choose the best candidate peer and
	// set it as the current sync peer
	bm.bestSyncCandidate = bm.pickBestSyncCandidate()
	bm.syncMtx.Unlock()

	var peer = bm.engine.peerManager.GetPeer(bm.bestSyncCandidate.PeerID)
	if peer == nil {
		err := fmt.Errorf("best candidate not found in peer list")
		bm.log.Debug(err.Error(), "PeerID", bm.bestSyncCandidate.PeerID)
		delete(bm.syncCandidate, bm.bestSyncCandidate.PeerID)
		goto resync
	}

	// Request block hashes from the peer
	blockHashes, err = bm.engine.gossipMgr.SendGetBlockHashes(peer, nil,
		bm.bestSyncCandidate.LastBlockSent)
	if err != nil {
		bm.log.Debug("Failed to get block hashes", "Err", err.Error())
		delete(bm.syncCandidate, bm.bestSyncCandidate.PeerID)
		goto resync
	}

	// Request for block bodies
	blockBodies, err = bm.engine.gossipMgr.SendGetBlockBodies(peer, blockHashes.Hashes)
	if err != nil {
		bm.log.Debug("Failed to get block bodies", "Err", err.Error())
		delete(bm.syncCandidate, bm.bestSyncCandidate.PeerID)
		goto resync
	}

	bm.log.Debug("Received block bodies",
		"PeerID", bm.bestSyncCandidate.PeerID,
		"NumBlockBodies", len(blockBodies.Blocks))

	// Attempt to append the block bodies to the blockchain
	for _, bb := range blockBodies.Blocks {
		var block core.Block
		copier.Copy(&block, bb)

		// Set core.ContextBlockSync to inform the block
		// process to validate the block as synced block
		// and set the broadcaster
		block.SetValidationContexts(types.ContextBlockSync)
		block.SetBroadcaster(peer)
		bm.bestSyncCandidate.LastBlockSent = block.GetHash()

		hk := common.KeyBlock2(block.GetHashAsHex(), bm.bestSyncCandidate.PeerID)
		bm.engine.GetHistory().AddMulti(cache.Sec(600), hk...)

		// Process the block
		bm.unprocessedBlocks.Append(&unprocessedBlock{
			block: &block,
		})
	}

	// Let's check if the candidate is still a viable
	// sync candidate. If it is not, remove it as a
	// sync candidate and proceed to starting the sync
	// process with another peer.
	if !bm.isSyncCandidate(bm.bestSyncCandidate) {
		delete(bm.syncCandidate, bm.bestSyncCandidate.PeerID)
		goto resync
	}

	syncStatus = bm.GetSyncStateInfo()
	if syncStatus != nil {
		bm.log.Info("Current synchronization status",
			"TargetTD", syncStatus.TargetTD,
			"CurTD", syncStatus.CurrentTD,
			"TargetChainHeight", syncStatus.TargetChainHeight,
			"CurChainHeight", syncStatus.CurrentChainHeight,
			"Progress(%)", syncStatus.ProgressPercent)
	}

resync:
	bm.syncMtx.Lock()
	bm.syncing = false
	bm.bestSyncCandidate = nil
	bm.syncMtx.Unlock()
	bm.sync()

	return nil
}

// IsSyncing checks whether block syncing is active.
func (bm *BlockManager) IsSyncing() bool {
	bm.syncMtx.RLock()
	defer bm.syncMtx.RUnlock()
	return bm.syncing
}

// GetSyncStateInfo returns progress information about
// the current blockchain synchronization session.
func (bm *BlockManager) GetSyncStateInfo() *core.SyncStateInfo {

	if !bm.IsSyncing() || bm.bestSyncCandidate == nil {
		return nil
	}

	var syncState = &core.SyncStateInfo{}

	// Get the current local best chain
	localBestBlock, _ := bm.engine.GetBlockchain().ChainReader().Current()
	syncState.TargetTD = bm.bestSyncCandidate.PeerChainTD
	syncState.TargetChainHeight = bm.bestSyncCandidate.PeerChainHeight
	syncState.CurrentTD = localBestBlock.GetHeader().GetTotalDifficulty()
	syncState.CurrentChainHeight = localBestBlock.GetNumber()

	// compute progress percentage based
	// on block height differences
	pct := float64(100) * (float64(syncState.CurrentChainHeight) /
		float64(syncState.TargetChainHeight))
	syncState.ProgressPercent, _ = decimal.NewFromFloat(pct).
		Round(1).Float64()

	return syncState
}

// addSyncCandidate adds a sync candidate.
// If the candidate already exists, it updates
// it only if the update candidate has a greater
// total difficulty
func (bm *BlockManager) addSyncCandidate(candidate *types.SyncPeerChainInfo) {
	bm.syncMtx.Lock()

	existing, ok := bm.syncCandidate[candidate.PeerID]
	if ok {
		if existing.PeerChainTD.Cmp(candidate.PeerChainTD) == 0 {
			bm.syncMtx.Unlock()
			return
		}
		candidate.LastBlockSent = existing.LastBlockSent
		bm.log.Debug("Updated sync candidate", "PeerID", candidate.PeerID)
	} else {
		bm.log.Debug("Added new sync candidate", "PeerID", candidate.PeerID)
	}

	bm.syncCandidate[candidate.PeerID] = candidate
	bm.syncMtx.Unlock()
}
