package blockchain

import (
	"fmt"
	"sync"

	"github.com/ellcrys/elld/blockchain/common"
	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/database"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/util/logger"
	"github.com/ellcrys/elld/wire"
)

// Chain represents a chain of blocks
type Chain struct {

	// id represents the identifier of this chain
	id string

	// parentBlock represents the block from which this chain is formed.
	// A chain that is not a subtree of another chain will have this set to nil.
	parentBlock *wire.Block

	// info holds information about the chain
	info *common.ChainInfo

	// cfg includes configuration parameters of the client
	cfg *config.EngineConfig

	// chainLock is used to synchronize access to fields that are
	// accessed concurrently
	chainLock *sync.RWMutex

	// store provides functionalities for storing objects
	store common.Store

	// log is used for logging
	log logger.Logger
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
	return chain
}

// init initializes a new chain with a block. The block becomes the "genesis"
// block of the chain if no block already exists or attempts to append the block
// to the current block at the time.
func (c *Chain) init(block string) error {
	b, err := wire.BlockFromString(block)
	if err != nil {
		return fmt.Errorf("failed to unmarshal block data: %s", err)
	}
	return c.append(b)
}

// getTipHeader returns the header of the highest block on this chain
func (c *Chain) getTipHeader(opts ...common.CallOp) (*wire.Header, error) {
	c.chainLock.RLock()
	defer c.chainLock.RUnlock()

	h, err := c.store.GetBlockHeader(c.id, 0, opts...)
	if err != nil {
		return nil, err
	}

	return h, nil
}

// height returns the height of this chain. The height can
// be deduced by fetching the number of the most recent block
// added to the chain.
func (c *Chain) height(opts ...common.CallOp) (uint64, error) {
	tip, err := c.getTipHeader(opts...)
	if err != nil {
		if err != common.ErrBlockNotFound {
			return 0, err
		}
		return 0, nil
	}
	return tip.GetNumber(), nil
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

// append adds a block to the tail of the chain. It returns
// error if the previous block hash in the header is not the hash
// of the current block and if the difference between the chain tip
// and the candidate block number is not 1. If there is no block on
// the chain yet, then we assume this to be the first block of a fork.
func (c *Chain) append(candidate *wire.Block, opts ...common.CallOp) error {
	c.chainLock.Lock()
	defer c.chainLock.Unlock()

	var err error
	var txOp = common.GetTxOp(c.store, opts...)

	// ensure it passes validation
	// if err := block.Validate(); err != nil {
	// 	if canFinish {
	// 		tx.Rollback()
	// 	}
	// 	return fmt.Errorf("genesis block failed validation: %s", err)
	// }

	// // verify the block
	// if err := wire.BlockVerify(block); err != nil {
	// 	if canFinish {
	// 		tx.Rollback()
	// 	}
	// 	return fmt.Errorf("genesis block signature is not valid: %s", err)
	// }

	// Get the current block at the tip of the chain.
	// Continue if no error or no block currently exist on the chain.
	chainTip, err := c.store.GetBlock(c.id, 0, common.TxOp{Tx: txOp.Tx})
	if err != nil {
		if err != common.ErrBlockNotFound {
			if txOp.CanFinish {
				txOp.Tx.Rollback()
			}
			return err
		}
	}

	// If the difference between the tip's block number and the new block number
	// is not 1, then this new block will not satisfy the serial numbering of blocks.
	if chainTip != nil && (candidate.GetNumber()-chainTip.GetNumber()) != 1 {
		if txOp.CanFinish {
			txOp.Tx.Rollback()
		}
		return fmt.Errorf(fmt.Sprintf("unable to append: candidate block number {%d} "+
			"is not the expected block number {expected=%d}", candidate.GetNumber(), chainTip.GetNumber()+1))
	}

	// If we found the current chainTip and its hash does not correspond with the
	// hash of the block we are trying to append, then we return an error.
	if chainTip != nil && chainTip.Hash != candidate.Header.ParentHash {
		if txOp.CanFinish {
			txOp.Tx.Rollback()
		}
		return fmt.Errorf("unable to append block: parent hash does not match the hash of the current block")
	}

	return c.store.PutBlock(c.id, candidate, txOp)
}

// NewStateTree creates a new tree. When backLinked is set to true
// the root of the previous block is included as the first entry to the tree.
func (c *Chain) NewStateTree(noBackLink bool, opts ...common.CallOp) (*Tree, error) {

	var prevRoot []byte

	// Get the root of the block at the tip. If no block was found, it means the chain is empty.
	// In this case, if the chain has a parent block, we use the parent block stateRoot.
	tipHeader, err := c.getTipHeader(opts...)
	if err != nil {
		if err != common.ErrBlockNotFound {
			return nil, err
		}
		if c.parentBlock != nil {
			prevRoot, err = util.FromHex(c.parentBlock.Header.StateRoot)
			if err != nil {
				return nil, fmt.Errorf("failed to decode parent state root")
			}
		}
	} else {
		// Decode the state root to byte equivalent
		prevRoot, err = util.FromHex(tipHeader.StateRoot)
		if err != nil {
			return nil, fmt.Errorf("failed to decode chain tip state root")
		}
	}

	// Create the new tree and seed it by adding the root
	// of the previous block as the first item. This ensures
	// cryptographic, backward linkage with previous blocks. No need do
	// this if no tip block or noBackLink is true
	tree := NewTree()
	if !noBackLink && len(prevRoot) > 0 {
		tree.Add(TreeItem(prevRoot))
	}

	return tree, nil
}

// putTransactions stores the provided transactions
// under the namespace of the chain.
func (c *Chain) putTransactions(txs []*wire.Transaction, opts ...common.CallOp) error {
	c.chainLock.Lock()
	defer c.chainLock.Unlock()

	var txOp = common.GetTxOp(c.store, opts...)

	for i, tx := range txs {
		txKey := common.MakeTxKey(c.id, tx.ID())
		if err := txOp.Tx.Put([]*database.KVObject{database.NewKVObject(txKey, util.ObjectToBytes(tx))}); err != nil {
			if txOp.CanFinish {
				txOp.Tx.Rollback()
			}
			return fmt.Errorf("index %d: %s", i, err)
		}
	}

	if txOp.CanFinish {
		return txOp.Tx.Commit()
	}

	return nil
}
