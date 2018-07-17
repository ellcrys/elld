package blockchain

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/ellcrys/elld/util"

	"github.com/imdario/mergo"

	"github.com/ellcrys/elld/blockchain/common"
	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/database"
	"github.com/ellcrys/elld/util/logger"
	"github.com/ellcrys/elld/wire"
)

// Chain represents a chain of blocks
type Chain struct {

	// id represents the identifier of this chain
	id string

	// parentBlock represents the block from which this chain was formed.
	// A chain that is not a subtree of another chain will have this set to nil.
	parentBlock *wire.Block

	// cfg includes configuration parameters of the client
	cfg *config.EngineConfig

	// chainLock is used to synchronize access to fields that are
	// accessed concurrently
	chainLock *sync.RWMutex

	// store provides functionalities for storing objects
	store common.Store

	// log is used for logging
	log logger.Logger

	// stateTree is used to store the hashes of all objects in the chain
	// such that the transition of objects from one state to another
	// is deterministically verifiable.
	stateTree *HashTree
}

// NewChain creates an instance of a chain. It will create metadata object for the
// chain if not exists. It will return error if it is unable to do so.
func NewChain(id string, store common.Store, cfg *config.EngineConfig, log logger.Logger) (*Chain, error) {

	chain := new(Chain)
	chain.id = id
	chain.cfg = cfg
	chain.store = store
	chain.chainLock = &sync.RWMutex{}
	chain.log = log
	chain.stateTree = NewHashTree(id, store)

	if err := chain.initMetadata(); err != nil {
		return nil, err
	}

	return chain, nil
}

// NewChainWithTx is like NewChain except it accepts a database transaction
func NewChainWithTx(tx database.Tx, id string, store common.Store, cfg *config.EngineConfig, log logger.Logger) (*Chain, error) {

	chain := new(Chain)
	chain.id = id
	chain.cfg = cfg
	chain.store = store
	chain.chainLock = &sync.RWMutex{}
	chain.log = log
	chain.stateTree = NewHashTree(id, store)

	if err := chain.initMetadataWithTx(tx); err != nil {
		return nil, err
	}

	return chain, nil
}

func (c *Chain) setParentBlock(b *wire.Block) {
	c.parentBlock = b
}

// initMetadata creates a metadata object for keep track of the
// state of this chain over time. Useful for caching computed
// values that are expensive to recompute.
func (c *Chain) initMetadata() error {
	var meta = c.getMeta()
	if meta == nil {
		if err := c.updateMeta(&common.ChainMeta{}); err != nil {
			return fmt.Errorf("failed to update chain metadata: %s", err)
		}
	}
	return nil
}

// initMetadataWithTx is like initMetadata except it accepts database transaction object
func (c *Chain) initMetadataWithTx(tx database.Tx) error {
	var meta = c.getMeta()
	if meta == nil {
		if err := c.updateMetaWithTx(tx, &common.ChainMeta{}); err != nil {
			return fmt.Errorf("failed to update chain metadata: %s", err)
		}
	}
	return nil
}

// init initializes a new chain. If there is no committed genesis block,
// it adds a genesis block to the store if it does not have one.
func (c *Chain) init(genesisBlockJSON string) error {
	var err error
	var genBlock = &wire.Block{}

	// attempt to fetch the genesis block. If it does not exists, we must create it
	err = c.store.GetBlock(c.id, 1, &wire.Block{})
	if err != nil {

		if err != common.ErrBlockNotFound {
			c.chainLock.Unlock()
			return fmt.Errorf("failed to check genesis block existence: %s", err)
		}

		// Unmarshal the genesis block JSON data to wire.Block
		genBlock, err = wire.BlockFromString(genesisBlockJSON)
		if err != nil {
			return fmt.Errorf("failed to unmarshal genesis block data: %s", err)
		}

		// 	// ensure it passes validation
		// 	if err := genBlock.Validate(); err != nil {
		// 		return fmt.Errorf("genesis block failed validation: %s", err)
		// 	}

		// 	// verify the block
		// 	if err := wire.BlockVerify(genBlock); err != nil {
		// 		return fmt.Errorf("genesis block signature is not valid: %s", err)
		// 	}

		// Save the genesis block
		if err := c.store.PutBlock(c.id, genBlock); err != nil {
			return fmt.Errorf("failed to commit genesis block to store. %s", err)
		}
	}

	// genesisKey := crypto.NewKeyFromIntSeed(127465328937663)
	// fmt.Println(genesisKey.PubKey().Base58())

	return nil
}

// getTipHeader returns the header of the highest block on this chain
func (c *Chain) getTipHeader() (*wire.Header, error) {
	var h wire.Header
	if err := c.store.GetBlockHeader(c.id, 0, &h); err != nil {
		return nil, err
	}
	return &h, nil
}

// getMatureTickets returns mature ticket transactions in the
// last n blocks.
// This method is safe for concurrent calls.
func (c *Chain) getMatureTickets(nLastBlocks uint64) (mTxs []*wire.Transaction, err error) {

	c.chainLock.RLock()
	defer c.chainLock.RUnlock()

	// get the current block header. We need the block height.
	curBlockHeader, err := c.getTipHeader()
	if err != nil {
		return nil, err
	}

	var startBlock uint64
	var endBlock = uint64(curBlockHeader.Number)

	// Set startBlock to 1 if the chain height is less or equal to nLastBlocks specified
	if curBlockHeader.Number <= nLastBlocks {
		startBlock = 1
	} else if curBlockHeader.Number > nLastBlocks {
		startBlock = curBlockHeader.Number - nLastBlocks + 1
	}

	// Find blocks within this range and get the matured endorsers
	for i := startBlock; i <= endBlock; i++ {

		var block wire.Block
		if err := c.store.GetBlock(c.id, startBlock, &block); err != nil {
			return nil, err
		}

		// find endorser ticket transactions and collect the
		// matured and non-revoked ones.
		for _, tx := range block.Transactions {
			if tx.Type == wire.TxTypeEndorserTicketCreate &&
				(curBlockHeader.Number-block.Header.Number) >= uint64(c.cfg.Consensus.NumBlocksForTicketMaturity) {
				mTxs = append(mTxs, tx)
			}
		}
	}

	return

}

// hasBlock checks if a block with the provided hash exists on this chain
func (c *Chain) hasBlock(hash string) (bool, error) {
	c.chainLock.RLock()
	defer c.chainLock.RUnlock()

	var header wire.Header
	if err := c.store.GetBlockHeaderByHash(c.id, hash, &header); err != nil {
		if err != common.ErrBlockNotFound {
			return false, err
		}
	}

	return header != wire.NilHeader, nil
}

// getBlockHeaderByHash returns the header of a block that matches the hash on this chain
func (c *Chain) getBlockHeaderByHash(hash string) (*wire.Header, error) {
	c.chainLock.RLock()
	defer c.chainLock.RUnlock()

	var header wire.Header
	if err := c.store.GetBlockHeaderByHash(c.id, hash, &header); err != nil {
		return nil, err
	}

	return &header, nil
}

// getBlockByHash fetches a block by its hash
func (c *Chain) getBlockByHash(hash string) (*wire.Block, error) {
	c.chainLock.RLock()
	defer c.chainLock.RUnlock()
	var block wire.Block
	if err := c.store.GetBlockByHash(c.id, hash, &block); err != nil {
		return nil, err
	}
	return &block, nil
}

// appendBlock adds a block to the tail of the chain. It returns
// error if the previous block hash in the header is not the hash
// of the current block. If there is no block on the chain yet,
// then we assume this to be the first block of a fork.
func (c *Chain) appendBlock(block *wire.Block) error {
	tx := c.store.NewTx()
	if err := c.appendBlockWithTx(tx, block); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

// appendBlockWithTx is like appendBlock except it accepts a transaction object
func (c *Chain) appendBlockWithTx(tx database.Tx, block *wire.Block) error {
	c.chainLock.Lock()
	defer c.chainLock.Unlock()

	var curBlock wire.Block
	if err := c.store.GetBlockWithTx(tx, c.id, 0, &curBlock); err != nil {
		if err != common.ErrBlockNotFound {
			return err
		}
	}

	if curBlock.Header != nil && curBlock.Hash != block.Header.ParentHash {
		return fmt.Errorf("unable to append block: parent hash does not match the hash of the current block")
	}

	return c.store.PutBlockWithTx(tx, c.id, block)
}

// getMeta returns the metadata of this chain
func (c *Chain) getMeta() *common.ChainMeta {
	var result database.KVObject
	c.store.GetFirstOrLast(false, common.MakeChainMetadataKey(c.id), &result)
	var meta common.ChainMeta
	json.Unmarshal(result.Value, &meta)
	return &meta
}

// updateMeta updates the metadata of this chain
func (c *Chain) updateMeta(upd *common.ChainMeta) error {
	existingMeta := c.getMeta()
	mergo.Merge(existingMeta, upd)
	return c.store.Put(common.MakeChainMetadataKey(c.id), util.ObjectToBytes(existingMeta))
}

// updateMeta updates the metadata of this chain
func (c *Chain) updateMetaWithTx(tx database.Tx, upd *common.ChainMeta) error {
	existingMeta := c.getMeta()
	mergo.Merge(existingMeta, upd)
	return c.store.PutWithTx(tx, common.MakeChainMetadataKey(c.id), util.ObjectToBytes(existingMeta))
}
