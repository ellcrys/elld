package blockchain

import (
	"fmt"
	"sync"
	"time"

	"github.com/ellcrys/elld/types/core"

	"github.com/ellcrys/elld/blockchain/common"
	"github.com/ellcrys/elld/blockchain/store"
	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/elldb"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/util/logger"
)

// ChainOp defines a method option for passing a chain object
type ChainOp struct {
	Chain *Chain
}

// GetName returns the name of the op
func (t ChainOp) GetName() string {
	return "ChainOp"
}

// Chain represents a chain of blocks
// Implements core.Chainer
type Chain struct {

	// id represents the identifier of this chain
	id util.String

	// parentBlock represents the block from which this chain is formed.
	// A chain that is not a subtree of another chain will have this set to nil.
	parentBlock core.Block

	// info holds information about the chain
	info *core.ChainInfo

	// cfg includes configuration parameters of the client
	cfg *config.EngineConfig

	// chainLock is used to synchronize access to fields that are
	// accessed concurrently
	chainLock *sync.RWMutex

	// store provides functionalities for storing objects
	store store.ChainStorer

	// log is used for logging
	log logger.Logger
}

// NewChain creates an instance of a chain. It will create metadata object for the
// chain if not exists. It will return error if it is unable to do so.
func NewChain(id util.String, db elldb.DB, cfg *config.EngineConfig, log logger.Logger) *Chain {
	chain := new(Chain)
	chain.id = id
	chain.cfg = cfg
	chain.store = store.New(db, chain.id)
	chain.chainLock = &sync.RWMutex{}
	chain.log = log
	chain.info = &core.ChainInfo{
		ID:        id,
		Timestamp: time.Now().UnixNano(),
	}
	return chain
}

// GetID returns the id of the chain
func (c *Chain) GetID() util.String {
	return c.id
}

// ChainReader gets a chain reader for this chain
func (c *Chain) ChainReader() core.ChainReader {
	return store.NewChainReader(c.store, c.id)
}

// GetParentBlock gets the chain's parent block if it has one
func (c *Chain) GetParentBlock() core.Block {
	return c.parentBlock
}

// GetParentInfo gets the parent info
func (c *Chain) GetParentInfo() *core.ChainInfo {
	return c.info
}

// GetBlock fetches a block by its number
func (c *Chain) GetBlock(number uint64) (core.Block, error) {
	c.chainLock.RLock()
	defer c.chainLock.RUnlock()

	b, err := c.store.GetBlock(number)
	if err != nil {
		return nil, err
	}

	return b, nil
}

// Current returns the header of the highest block on this chain
func (c *Chain) Current(opts ...core.CallOp) (core.Header, error) {
	c.chainLock.RLock()
	defer c.chainLock.RUnlock()

	h, err := c.store.GetHeader(0, opts...)
	if err != nil {
		return nil, err
	}

	return h, nil
}

// height returns the height of this chain. The height can
// be deduced by fetching the number of the most recent block
// added to the chain.
func (c *Chain) height(opts ...core.CallOp) (uint64, error) {
	tip, err := c.Current(opts...)
	if err != nil {
		if err != core.ErrBlockNotFound {
			return 0, err
		}
		return 0, nil
	}
	return tip.GetNumber(), nil
}

// hasBlock checks if a block with the provided hash exists on this chain
func (c *Chain) hasBlock(hash util.Hash) (bool, error) {
	c.chainLock.RLock()
	defer c.chainLock.RUnlock()

	h, err := c.store.GetHeaderByHash(hash)
	if err != nil {
		if err != core.ErrBlockNotFound {
			return false, err
		}
	}

	return h != nil, nil
}

// getBlockHeaderByHash returns the header of a block that matches the hash on this chain
func (c *Chain) getBlockHeaderByHash(hash util.Hash) (core.Header, error) {
	c.chainLock.RLock()
	defer c.chainLock.RUnlock()

	h, err := c.store.GetHeaderByHash(hash)
	if err != nil {
		return nil, err
	}

	return h, nil
}

// getBlockByHash fetches a block by hash
func (c *Chain) getBlockByHash(hash util.Hash) (core.Block, error) {
	c.chainLock.RLock()
	defer c.chainLock.RUnlock()
	return c.store.GetBlockByHash(hash)
}

// getBlockByNumberAndHash fetches a block by number and hash
func (c *Chain) getBlockByNumberAndHash(number uint64, hash util.Hash) (core.Block, error) {
	c.chainLock.RLock()
	defer c.chainLock.RUnlock()
	return c.store.GetBlockByNumberAndHash(number, hash)
}

// CreateAccount creates an account on a target block
func (c *Chain) CreateAccount(targetBlockNum uint64, account core.Account, opts ...core.CallOp) error {
	c.chainLock.Lock()
	defer c.chainLock.Unlock()
	return c.store.CreateAccount(targetBlockNum, account, opts...)
}

// GetAccount gets an account
func (c *Chain) GetAccount(address util.String, opts ...core.CallOp) (core.Account, error) {
	c.chainLock.RLock()
	defer c.chainLock.RUnlock()
	return c.store.GetAccount(address, opts...)
}

// append adds a block to the tail of the chain. It returns
// error if the previous block hash in the header is not the hash
// of the current block and if the difference between the chain tip
// and the candidate block number is not 1. If there is no block on
// the chain yet, then we assume this to be the first block of a fork.
//
// The caller is expected to validate the block before call.
func (c *Chain) append(candidate core.Block, opts ...core.CallOp) error {
	c.chainLock.Lock()
	defer c.chainLock.Unlock()

	var err error
	var txOp = common.GetTxOp(c.store, opts...)

	// Get the current block at the tip of the chain.
	// Continue if no error or no block currently exist on the chain.
	chainTip, err := c.store.Current(common.TxOp{Tx: txOp.Tx})
	if err != nil {
		if err != core.ErrBlockNotFound {
			txOp.Rollback()
			return err
		}
	}

	// If the difference between the tip's block number and the new block number
	// is not 1, then this new block will not satisfy the serial numbering of blocks.
	if chainTip != nil && (candidate.GetNumber()-chainTip.GetNumber()) != 1 {
		txOp.Rollback()
		return fmt.Errorf(fmt.Sprintf("unable to append: candidate block number {%d} "+
			"is not the expected block number {expected=%d}", candidate.GetNumber(), chainTip.GetNumber()+1))
	}

	// If we found the current chainTip and its hash does not correspond with the
	// hash of the block we are trying to append, then we return an error.
	if chainTip != nil && !chainTip.GetHash().Equal(candidate.GetHeader().GetParentHash()) {
		txOp.Rollback()
		return fmt.Errorf("unable to append block: parent hash does not match the hash of the current block")
	}

	return c.store.PutBlock(candidate, txOp)
}

// NewStateTree creates a new tree seeded with the state root of
// the chain's tip block or the chain's parent block for side chains that have
// no block but are children of a block in a parent block.
//
// When backLinked is set to false the tree is not seeded with the state root
// of the previous tip block or chain parent block.
func (c *Chain) NewStateTree(noBackLink bool, opts ...core.CallOp) (core.Tree, error) {

	var prevRoot util.Hash

	// Get the root of the block at the tip. If no block was found, it means the chain is empty.
	// In this case, if the chain has a parent block, we use the parent block stateRoot.
	if !noBackLink {
		tipHeader, err := c.Current(opts...)
		if err != nil {
			if err != core.ErrBlockNotFound {
				return nil, err
			}
			if c.parentBlock != nil {
				prevRoot = c.parentBlock.GetHeader().GetStateRoot()
			}
		} else {
			prevRoot = tipHeader.GetStateRoot()
			if err != nil {
				return nil, fmt.Errorf("failed to decode chain tip state root")
			}
		}
	}

	// Create the new tree and seed it by adding the root
	// of the previous state root. No need to do
	// this if we have not determined the previous state root.
	tree := common.NewTree()
	if !prevRoot.IsEmpty() {
		tree.Add(common.TreeItem(prevRoot.Bytes()))
	}

	return tree, nil
}

// PutTransactions stores a collection of transactions in the chain
func (c *Chain) PutTransactions(txs []core.Transaction, blockNumber uint64, opts ...core.CallOp) error {
	return c.store.PutTransactions(txs, blockNumber, opts...)
}

// GetTransaction gets a transaction by hash
func (c *Chain) GetTransaction(hash util.Hash) core.Transaction {
	return c.store.GetTransaction(hash)
}

// removeBlock deletes a block and all objects
// associated to it such as transactions, accounts etc.
func (c *Chain) removeBlock(number uint64) error {
	return nil
}
