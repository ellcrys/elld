package blockchain

import (
	"fmt"
	"sync"

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
func NewChain(id string, store common.Store, cfg *config.EngineConfig, log logger.Logger) *Chain {
	chain := new(Chain)
	chain.id = id
	chain.cfg = cfg
	chain.store = store
	chain.chainLock = &sync.RWMutex{}
	chain.log = log
	chain.stateTree = NewHashTree(id, store)
	return chain
}

func (c *Chain) setParentBlock(b *wire.Block) {
	c.parentBlock = b
}

// init initializes a new chain with a block. The block becomes the "genesis"
// block of the chain if no block already exists or attempts to append the block
// to the current block at the time.
func (c *Chain) init(block string) error {

	// Unmarshal the genesis block JSON data to wire.Block
	genBlock, err := wire.BlockFromString(block)
	if err != nil {
		return fmt.Errorf("failed to unmarshal genesis block data: %s", err)
	}

	return c.appendBlock(genBlock)
}

// getTipHeader returns the header of the highest block on this chain
func (c *Chain) getTipHeader() (*wire.Header, error) {
	h, err := c.store.GetBlockHeader(c.id, 0)
	if err != nil {
		return nil, err
	}
	return h, nil
}

// hasBlock checks if a block with the provided hash exists on this chain
func (c *Chain) hasBlock(hash string) (bool, error) {
	c.chainLock.RLock()
	defer c.chainLock.RUnlock()

	h, err := c.store.GetBlockHeaderByHash(c.id, hash)
	if err != nil {
		if err != common.ErrBlockNotFound {
			return false, err
		}
	}

	return h != nil, nil
}

// getBlockHeaderByHash returns the header of a block that matches the hash on this chain
func (c *Chain) getBlockHeaderByHash(hash string) (*wire.Header, error) {
	c.chainLock.RLock()
	defer c.chainLock.RUnlock()

	h, err := c.store.GetBlockHeaderByHash(c.id, hash)
	if err != nil {
		return nil, err
	}

	return h, nil
}

// getBlockByHash fetches a block by its hash
func (c *Chain) getBlockByHash(hash string) (*wire.Block, error) {
	c.chainLock.RLock()
	defer c.chainLock.RUnlock()

	block, err := c.store.GetBlockByHash(c.id, hash)
	if err != nil {
		return nil, err
	}

	return block, nil
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

	// ensure it passes validation
	// 	if err := genBlock.Validate(); err != nil {
	// 		return fmt.Errorf("genesis block failed validation: %s", err)
	// 	}

	// 	// verify the block
	// 	if err := wire.BlockVerify(genBlock); err != nil {
	// 		return fmt.Errorf("genesis block signature is not valid: %s", err)
	// 	}

	// Get the current block at the tip of the chain.
	// Continue if no error or no block currently exist on the chain.
	curBlock, err := c.store.GetBlock(c.id, 0, common.TxOp{Tx: tx})
	if err != nil {
		if err != common.ErrBlockNotFound {
			return err
		}
	}

	// If we found the current curBlock and its hash does not correspond with the
	// hash of the block we are trying to append, then we return an error.
	if curBlock != nil && curBlock.Hash != block.Header.ParentHash {
		return fmt.Errorf("unable to append block: parent hash does not match the hash of the current block")
	}

	return c.store.PutBlock(c.id, block, common.TxOp{Tx: tx})
}
