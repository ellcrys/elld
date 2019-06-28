package node

import (
	"sync"

	"github.com/ellcrys/mother/util/cache"

	"gopkg.in/oleiade/lane.v1"

	"github.com/ellcrys/mother/types"
	"github.com/ellcrys/mother/util/logger"
	"github.com/olebedev/emitter"
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

}
