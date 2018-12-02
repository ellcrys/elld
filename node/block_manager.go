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
}

// NewBlockManager creates a new BlockManager
func NewBlockManager(bChain types.Blockchain, evt *emitter.Emitter, log logger.Logger) *BlockManager {
	bm := &BlockManager{
		evt:    evt,
		log:    log,
		bChain: bChain,
	}
	return bm
}

// SetMiner sets a reference of the CPU miner
func (bm *BlockManager) SetMiner(m *miner.Miner) {
	bm.miner = m
}

// Handle handles all incoming block related events.
func (bm *BlockManager) Handle() {
	go func() {
		for {
			evt := <-bm.evt.Once("*")
			switch evt.OriginalTopic {
			case core.EventFoundBlock:
				errCh := evt.Args[1].(chan error)
				errCh <- bm.handleMined(evt.Args[0].(*miner.FoundBlock))
			case core.EventNewBlock:
				bm.handleAppendedBlock(evt.Args[0].(*core.Block))
			}
		}
	}()
}

// handle mined block by attempting to append it to
// the main chain.
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

// handleAppendedBlock handles an event about a block
// that got appended to the main chain.
// It will restart the miner if the miner workers
func (bm *BlockManager) handleAppendedBlock(b types.Block) {
	bm.miner.RestartWorkers()
}
