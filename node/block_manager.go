package node

import (
	"sync"
	"time"

	"github.com/ellcrys/elld/params"

	"github.com/ellcrys/elld/util/cache"

	"gopkg.in/oleiade/lane.v1"

	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/util/logger"
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

type processedBlock struct {

	// block is the processed block
	block types.Block

	// atSyncTime indicates that the block
	// was processed during a sync session
	atSyncTime bool
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

	// engine is the node's instance
	engine *Node

	// syncCandidate are candidate peers to
	// sync blocks with.
	syncCandidate map[string]*types.SyncPeerChainInfo

	// bestSyncCandidate is the current best sync
	// candidate to perform block synchronization with.
	bestSyncCandidate *types.SyncPeerChainInfo

	// unprocessed hold blocks that are yet to be processed
	unprocessed *lane.Deque

	// processedBlocks hold blocks that have been processed
	processedBlocks *lane.Deque

	// mined holds the hash of blocks mined by the client
	mined *cache.Cache
}

// NewBlockManager creates a new BlockManager
func NewBlockManager(node *Node) *BlockManager {
	bm := &BlockManager{
		syncMtx:         &sync.RWMutex{},
		log:             node.log,
		evt:             node.event,
		bChain:          node.bChain,
		engine:          node,
		unprocessed:     lane.NewDeque(),
		processedBlocks: lane.NewDeque(),
		mined:           cache.NewCache(100),
		syncCandidate:   make(map[string]*types.SyncPeerChainInfo),
	}
	return bm
}

// Manage handles all incoming block related events.
func (bm *BlockManager) Manage() {

	go func() {
		for evt := range bm.evt.On(core.EventNewBlock) {
			bm.processedBlocks.Append(&processedBlock{
				block:      evt.Args[0].(*core.Block),
				atSyncTime: bm.IsSyncing(),
			})
		}
	}()

	go func() {
		for evt := range bm.evt.On(core.EventOrphanBlock) {
			bm.handleOrphan(evt.Args[0].(*core.Block))
		}
	}()

	go func() {
		for evt := range bm.evt.On(core.EventProcessBlock) {
			bm.unprocessed.Append(&unprocessedBlock{
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
				for !bm.unprocessed.Empty() {
					bm.handleUnprocessedBlocks()
				}
			}
		}
	}()
}

// handleProcessedBlocks fetches blocks from the processed
// block queue to perform post-append operations
func (bm *BlockManager) handleProcessedBlocks() error {
	return nil
}

// handleUnprocessedBlocks fetches unprocessed blocks
// from the unprocessed block queue and attempts to
// append them to the blockchain.
func (bm *BlockManager) handleUnprocessedBlocks() error {

	var upb = bm.unprocessed.Shift()
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
}

// isSyncCandidate checks whether a peer is a
// valid sync candidate based on its chain state
// information.
// A peer is a valid candidate if the total difficulty
// of its best block is greater than that of the local best
// block
func (bm *BlockManager) isSyncCandidate(info *types.SyncPeerChainInfo) bool {
	localBestBlock, _ := bm.engine.GetBlockchain().ChainReader().Current()
	if localBestBlock.GetHeader().GetTotalDifficulty().Cmp(info.PeerChainTD) == -1 {
		bm.log.Info("Local blockchain is behind peer",
			"ChainHeight", localBestBlock.GetNumber(),
			"LocalTD", localBestBlock.GetHeader().GetTotalDifficulty(),
			"PeerID", info.PeerIDShort,
			"PeerChainHeight", info.PeerChainHeight,
			"PeerTD", info.PeerChainTD)
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

// sync
func (bm *BlockManager) sync() error {
	return nil
}

// IsSyncing checks whether block syncing is active.
func (bm *BlockManager) IsSyncing() bool {
	bm.syncMtx.RLock()
	defer bm.syncMtx.RUnlock()
	return bm.syncing
}

// GetSyncStat returns progress information about
// the current blockchain synchronization session.
func (bm *BlockManager) GetSyncStat() *core.SyncStateInfo {

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
