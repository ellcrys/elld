package node

import (
	"fmt"
	"sync"
	"time"

	"github.com/ellcrys/elld/miner"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/util/logger"
	"github.com/fatih/color"
	"github.com/jinzhu/copier"
	"github.com/olebedev/emitter"
	"github.com/shopspring/decimal"
)

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

	// txPool is the transaction pool
	txPool types.TxPool

	// peerMgr is the client peer manager
	engine *Node

	// syncCandidate are candidate peers to
	// sync blocks with.
	syncCandidate map[string]*types.SyncPeerChainInfo

	// bestSyncCandidate is the current best sync
	// candidate to perform block synchronization with.
	bestSyncCandidate *types.SyncPeerChainInfo
}

// NewBlockManager creates a new BlockManager
func NewBlockManager(node *Node) *BlockManager {
	bm := &BlockManager{
		syncMtx:       &sync.RWMutex{},
		log:           node.log,
		evt:           node.event,
		bChain:        node.bChain,
		engine:        node,
		syncCandidate: make(map[string]*types.SyncPeerChainInfo),
	}
	return bm
}

// SetMiner sets a reference of the CPU miner
func (bm *BlockManager) SetMiner(m *miner.Miner) {
	bm.miner = m
}

// SetTxPool sets a reference of the transaction pool
func (bm *BlockManager) SetTxPool(tp types.TxPool) {
	bm.txPool = tp
}

// Handle handles all incoming block related events.
func (bm *BlockManager) Handle() {
	for {
		evt := <-bm.evt.Once("*")
		switch evt.OriginalTopic {

		case core.EventFoundBlock:
			errCh := evt.Args[1].(chan error)
			go func(errCh chan error) {
				errCh <- bm.handleMined(evt.Args[0].(*miner.FoundBlock))
			}(errCh)

		case core.EventNewBlock:
			go bm.handleAppendedBlock(evt.Args[0].(*core.Block))

		case core.EventOrphanBlock:
			go bm.handleOrphan(evt.Args[0].(*core.Block))

		case core.EventPeerChainInfo:
			peerChainInfo := evt.Args[0].(*types.SyncPeerChainInfo)
			if bm.isSyncCandidate(peerChainInfo) {
				bm.addSyncCandidate(peerChainInfo)
				go func() {
					bm.sync()
					bm.log.Info("Block synchronization complete")
				}()
			}
		}
	}
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

// handleMined attempts to append a block with a valid
// PoW to the block chain
func (bm *BlockManager) handleMined(fb *miner.FoundBlock) error {

	_, err := bm.bChain.ProcessBlock(fb.Block)
	if err != nil {
		bm.log.Warn("Failed to process block", "Err", err.Error())
		return err
	}

	bm.log.Info(color.GreenString("New block mined"),
		"Number", fb.Block.GetNumber(),
		"Difficulty", fb.Block.GetHeader().GetDifficulty(),
		"TotalDifficulty", fb.Block.GetHeader().GetTotalDifficulty(),
		"PoW Time", time.Since(fb.Started))

	return nil
}

// relayAppendedBlock a block to connected peers
func (bm *BlockManager) relayAppendedBlock(b types.Block) {
	if b.GetNumber() > 1 {
		bm.engine.Gossip().RelayBlock(b, bm.engine.PM().GetConnectedPeers())
	}
}

// handleAppendedBlock handles an event about a block
// that got appended to the main chain.
func (bm *BlockManager) handleAppendedBlock(b types.Block) {

	// Remove the blocks transactions from the pool.
	bm.txPool.Remove(b.GetTransactions()...)

	// Restart miner workers.
	bm.miner.RestartWorkers()

	// Relay the block to peers.
	bm.relayAppendedBlock(b)
}

// isSyncCandidate checks whether a peer is a
// valid sync candidate based on its chain state
// information. A peer is a valid candidate if
// the total difficulty of its best block is
// greater that of the local best block.
func (bm *BlockManager) isSyncCandidate(info *types.SyncPeerChainInfo) bool {
	localBestBlock, _ := bm.engine.GetBlockchain().ChainReader().Current()
	if localBestBlock.GetHeader().GetTotalDifficulty().Cmp(info.PeerChainTotalDifficulty) == -1 {
		bm.log.Info("Local blockchain is behind peer",
			"ChainHeight", localBestBlock.GetNumber(),
			"TotalDifficulty", localBestBlock.GetHeader().GetTotalDifficulty(),
			"PeerID", info.PeerIDShort,
			"PeerChainHeight", info.PeerChainHeight,
			"PeerChainTotalDifficulty", info.PeerChainTotalDifficulty)
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
		if bestCandidate.PeerChainTotalDifficulty.
			Cmp(candidate.PeerChainTotalDifficulty) == -1 {
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
		return nil
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
	blockHashes, err = bm.engine.gossipMgr.SendGetBlockHashes(peer, nil)
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

	// Attempt to append the block bodies to the blockchain
	for _, bb := range blockBodies.Blocks {

		var block core.Block
		copier.Copy(&block, bb)

		// Set core.ContextBlockSync to inform the block
		// process to validate the block as synced block
		// and set the broadcaster
		block.SetValidationContexts(types.ContextBlockSync)
		block.SetBroadcaster(peer)

		// Process the block
		_, err := bm.engine.GetBlockchain().ProcessBlock(&block)
		if err != nil {
			bm.log.Debug("Unable to process block", "Err", err)
			goto resync
		}
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

	// Wait a few seconds before fetching more
	// sync objects from this candidate.
	time.Sleep(5 * time.Second)

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
	syncState.TargetTD = bm.bestSyncCandidate.PeerChainTotalDifficulty
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

	if c, ok := bm.syncCandidate[candidate.PeerID]; ok {
		if c.PeerChainTotalDifficulty.Cmp(candidate.PeerChainTotalDifficulty) >= 1 {
			bm.syncMtx.Unlock()
			return
		}
	}

	bm.log.Debug("Added new sync candidate", "PeerID", candidate.PeerID)
	bm.syncCandidate[candidate.PeerID] = candidate
	bm.syncMtx.Unlock()
}
