package node

import (
	"time"

	"github.com/ellcrys/elld/miner"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/util/logger"
	"github.com/fatih/color"
	"github.com/olebedev/emitter"
)

// BlockManager is responsible for handling
// incoming, mined or processed blocks in a
// concurrency safe way.
type BlockManager struct {

	// evt is the global event emitter
	evt *emitter.Emitter

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
}

// NewBlockManager creates a new BlockManager
func NewBlockManager(node *Node) *BlockManager {
	bm := &BlockManager{
		log:    node.log,
		evt:    node.event,
		bChain: node.bChain,
		engine: node,
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
	go func() {
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
			}
		}
	}()
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
