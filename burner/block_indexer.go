package burner

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ellcrys/elld/config"

	"github.com/ellcrys/ltcd/chaincfg/chainhash"

	"github.com/ellcrys/ltcd/wire"

	"github.com/ellcrys/elld/util"

	"github.com/ellcrys/elld/elldb"
	"github.com/ellcrys/elld/util/logger"
	"github.com/olebedev/emitter"
)

var errHeaderNotFound = fmt.Errorf("header not found")

// LocalBlockHeader represents slim version of wire.BlockHeader
type LocalBlockHeader struct {
	Number        int64  `json:"number"`
	Hash          []byte `json:"hash"`
	PrevBlockHash []byte `json:"prevHash"`
}

// minStartHeight is the minimum block height to start the watcher on.
// We set this to avoid indexing blocks well before the launch of
// the network and far in the history of the burn chain
var minStartHeight = int64(0)

// BlockIndexer determines maintains a light representation of the burn
// server best chain. When a new block on the burner chain
// is mined, it stores information about the block, broadcast
// an event to inform other processes. Additionally, it will detected
// re-organization and also broadcast an event.
type BlockIndexer struct {
	*sync.Mutex
	bus        *emitter.Emitter
	db         elldb.DB
	client     RPCClient
	log        logger.Logger
	lastHeight int64
	stop       bool
	interrupt  <-chan struct{}
	ticker     *time.Ticker
	cfg        *config.EngineConfig
}

// NewBlockIndexer creates an instance of BlockWatcher
func NewBlockIndexer(cfg *config.EngineConfig, log logger.Logger, db elldb.DB, bus *emitter.Emitter,
	client RPCClient, interrupt <-chan struct{}) *BlockIndexer {
	log = log.Module("BlockIndexer")
	return &BlockIndexer{
		Mutex:     &sync.Mutex{},
		bus:       bus,
		db:        db,
		client:    client,
		interrupt: interrupt,
		log:       log,
		cfg:       cfg,
	}
}

// getLatestLocalBlock returns the most recently
// indexed block header
func (bw *BlockIndexer) getLatestLocalBlock() (*LocalBlockHeader, error) {
	var lastObj *elldb.KVObject
	key := MakeQueryKeyIndexerBlock()
	bw.db.Iterate(key, false, func(kv *elldb.KVObject) bool {
		lastObj = kv
		return true
	})
	if lastObj == nil {
		return nil, errHeaderNotFound
	}
	var header LocalBlockHeader
	if err := lastObj.Scan(&header); err != nil {
		return nil, err
	}
	return &header, nil
}

// getLocalBlock gets an indexed block header by height.
func (bw *BlockIndexer) getLocalBlock(height int64) (*LocalBlockHeader, error) {
	key := MakeKeyIndexerBlock(height)
	kv := bw.db.GetFirstOrLast(key, true)
	if kv == nil {
		return nil, errHeaderNotFound
	}
	var header LocalBlockHeader
	if err := kv.Scan(&header); err != nil {
		return nil, err
	}
	return &header, nil
}

// getStartHeight returns the block height from which to start indexing blocks.
func (bw *BlockIndexer) getStartHeight() (int64, error) {
	bw.Lock()
	defer bw.Unlock()

	// First try to find the last indexed height
	// from memory before querying the database.
	height := bw.lastHeight
	if height == 0 {
		h, err := bw.getLatestLocalBlock()
		if err != nil {
			if err != errHeaderNotFound {
				return height, err
			}
		} else {
			height = h.Number
		}
	}

	// Get the upstream chain best block height.
	_, bestBlockHeight, err := bw.client.GetBestBlock()
	if err != nil {
		return 0, fmt.Errorf("failed to fetch best block height: %s", err)
	}

	// Compare the upstream chain height with the known last
	// indexed height; If the last indexed height is greater,
	// it means for some reason the indexer previously synced
	// a much more superior chain than the current upstream
	// chain. As such, we try to correct by moving one step
	// backwards. (Note: This should never happen)
	if bestBlockHeight > 0 && int64(bestBlockHeight) < height {
		height = int64(bestBlockHeight) - 1
	}

	// If at this point, the height is less than the minimum
	// start height, we have to set the height to the minimum.
	if height < minStartHeight {
		height = minStartHeight
	}

	return height, nil
}

// detectReorg checks if a new upstream block shares a parent block
// that occupies the same height on the local block index.
func (bw *BlockIndexer) detectReorg(newBlock *wire.BlockHeader, newBlockHeight int64) (bool, error) {

	if newBlockHeight == 1 {
		return false, nil
	}

	upstreamBlockParentHeight := newBlockHeight - 1

	// Get the block header of the local block on
	// same height as the new upstream block's parent.
	prevLocalHeader, err := bw.getLocalBlock(upstreamBlockParentHeight)
	if err != nil && err != errHeaderNotFound {
		return false, err
	}

	// Here, we have found that the upstream block does not share
	// a known parent, so we can't append it. It also means that
	// the local chain has diverged some where.
	if err == errHeaderNotFound {
		errMsg := "local block index is missing a common block (%d)"
		return false, fmt.Errorf(errMsg, upstreamBlockParentHeight)
	}

	// So, if the new block's parent hash matches the hash
	// of the block with same height, then no re-org happened.
	// Otherwise, a re-org has happened.
	prevHash, _ := chainhash.NewHash(prevLocalHeader.Hash)
	if newBlock.PrevBlock.IsEqual(prevHash) {
		return false, nil
	}

	return true, nil
}

// getUpstreamBlock requests a block by height
// from the burner server blockchain
func (bw *BlockIndexer) getUpstreamBlock(height int64) (*wire.BlockHeader, error) {
	hash, err := bw.client.GetBlockHash(height)
	if err != nil {
		return nil, err
	}
	header, err := bw.client.GetBlockHeader(hash)
	if err != nil {
		return nil, err
	}
	return header, nil
}

// deleteBlocksFrom deletes a block and all of its children.
// It first checks if the starting height block exists, if so,
// it will begin to go up the chain, deleting every block it
// finds and only stop when the next block height does not exist.
func (bw *BlockIndexer) deleteBlocksFrom(height int64) error {
	startHeight := height
	for {
		key := MakeKeyIndexerBlock(startHeight)
		if bw.db.GetFirstOrLast(key, true) == nil {
			break
		}
		if err := bw.db.DeleteByPrefix(key); err != nil {
			return err
		}
		startHeight++
	}
	return nil
}

// Stop the block indexer
func (bw *BlockIndexer) Stop() {
	bw.Lock()
	defer bw.Unlock()
	if bw.stop {
		return
	}
	bw.stop = true
	bw.ticker.Stop()
}

// Begin starts the watch process
func (bw *BlockIndexer) Begin() {

	dur := time.Duration(bw.cfg.Node.BurnerBlockIndexInterval) * time.Second
	bw.ticker = time.NewTicker(dur)

	// Start a goroutine that listens for general application
	// interrupt signal. If interrupted, it stops indexer.
	go func() {
		<-bw.interrupt
		bw.Stop()
		bw.log.Info("Block indexer has been interrupted and has stopped")
	}()

	// Here we run a goroutine that waits for ticks from
	// the indexer ticker. On each tick, it attempts to
	// find new blocks and index them
	go func() {
		for range bw.ticker.C {

			// Get the block height that was last indexed
			// or is the best height to start the search from.
			lastHeight, err := bw.getStartHeight()
			if err != nil {
				bw.log.Error("Failed to determine the last indexed block height")
				return
			}

			// Start finding newer blocks until we find no more
			cursor := lastHeight + 1
			for {

				// Find the header of the block at the given height
				header, err := bw.getUpstreamBlock(cursor)
				if err != nil {
					if strings.Index(err.Error(), "Block number out of range") != -1 {
						break
					}
					bw.log.Error("Failed to fetch block header", "Err", err.Error())
					return
				}

				// Before we append the new header, we need to ensure the header's
				// parent block matches the hash of the most recent block.
				// If they do not match, it means there was a reorg in the burner
				// chain. As such, we need to delete recent block and its lineage
				// and move the cursor 1 height back.
				ok, err := bw.detectReorg(header, cursor)
				if err != nil {
					bw.log.Error("Failed to complete re-org detection", "Err", err.Error())
				} else if ok {

					reorgedHeight := cursor - 1
					if err := bw.deleteBlocksFrom(reorgedHeight); err != nil {
						bw.log.Error("failed to delete invalidated block(s)", "Err", err)
						return
					}

					bw.log.Debug("[REORG DETECTED]. Deleting the block and its lineage",
						"Height", reorgedHeight)

					// Set the cursor back to the same height where the invalid block (and lineage)
					// was removed. We do this so, we can re-index the latest block from the
					// upstream chain.
					cursor = reorgedHeight

					// Emit a reorg event for other processes to react
					bw.bus.Emit(EventInvalidLocalBlock, reorgedHeight)

					continue
				}

				hash := header.BlockHash()
				localBlockHeader := &LocalBlockHeader{
					Number:        cursor,
					Hash:          hash[:],
					PrevBlockHash: header.PrevBlock[:],
				}

				// Save the block
				key := MakeKeyIndexerBlock(cursor)
				kv := elldb.NewKVObject(key, util.ObjectToBytes(localBlockHeader))
				if err := bw.db.Put([]*elldb.KVObject{kv}); err != nil {
					bw.log.Error("Failed to stored block header", "Err", err.Error())
					return
				}

				bw.log.Debug("Found a burner chain block", "Height", cursor, "Hash", hash.String())

				// Emit an EventNewBlock
				bw.bus.Emit(EventNewBlock, localBlockHeader)

				bw.Lock()
				bw.lastHeight = cursor
				bw.Unlock()

				cursor++
			}
		}
	}()
}
