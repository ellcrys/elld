package blockchain

import (
	"fmt"
	"sync"

	"github.com/ellcrys/elld/blockchain/common"
	"github.com/ellcrys/elld/blockchain/store"
	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/elldb"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/util/logger"
	"github.com/ellcrys/elld/wire"
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
// Implements common.Chainer
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
	store common.ChainStorer

	// log is used for logging
	log logger.Logger
}

// NewChain creates an instance of a chain. It will create metadata object for the
// chain if not exists. It will return error if it is unable to do so.
func NewChain(id string, db elldb.DB, cfg *config.EngineConfig, log logger.Logger) *Chain {
	chain := new(Chain)
	chain.id = id
	chain.cfg = cfg
	chain.store = store.New(db, chain.id)
	chain.chainLock = &sync.RWMutex{}
	chain.log = log
	chain.info = &common.ChainInfo{
		ID: id,
	}
	return chain
}

// GetID returns the id of the chain
func (c *Chain) GetID() string {
	return c.id
}

// GetParentBlock gets the chain's parent block if it has one
func (c *Chain) GetParentBlock() *wire.Block {
	return c.parentBlock
}

// GetParentInfo gets the parent info
func (c *Chain) GetParentInfo() *common.ChainInfo {
	return c.info
}

// GetBlock fetches a block by its number
func (c *Chain) GetBlock(number uint64) (*wire.Block, error) {
	c.chainLock.RLock()
	defer c.chainLock.RUnlock()

	b, err := c.store.GetBlock(number)
	if err != nil {
		return nil, err
	}

	return b, nil
}

// Current returns the header of the highest block on this chain
func (c *Chain) Current(opts ...common.CallOp) (*wire.Header, error) {
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
func (c *Chain) height(opts ...common.CallOp) (uint64, error) {
	tip, err := c.Current(opts...)
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

	h, err := c.store.GetHeaderByHash(hash)
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

	h, err := c.store.GetHeaderByHash(hash)
	if err != nil {
		return nil, err
	}

	return h, nil
}

// getBlockByHash fetches a block by its hash
func (c *Chain) getBlockByHash(hash string) (*wire.Block, error) {
	c.chainLock.RLock()
	defer c.chainLock.RUnlock()

	block, err := c.store.GetBlockByHash(hash)
	if err != nil {
		return nil, err
	}

	return block, nil
}

// CreateAccount creates an account on a target block
func (c *Chain) CreateAccount(targetBlockNum uint64, account *wire.Account, opts ...common.CallOp) error {
	c.chainLock.Lock()
	defer c.chainLock.Unlock()
	return c.store.CreateAccount(targetBlockNum, account, opts...)
}

// GetAccount gets an account
func (c *Chain) GetAccount(address string, opts ...common.CallOp) (*wire.Account, error) {
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
func (c *Chain) append(candidate *wire.Block, opts ...common.CallOp) error {
	c.chainLock.Lock()
	defer c.chainLock.Unlock()

	var err error
	var txOp = common.GetTxOp(c.store, opts...)

	// Get the current block at the tip of the chain.
	// Continue if no error or no block currently exist on the chain.
	chainTip, err := c.store.Current(common.TxOp{Tx: txOp.Tx})
	if err != nil {
		if err != common.ErrBlockNotFound {
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
	if chainTip != nil && chainTip.Hash != candidate.Header.ParentHash {
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
func (c *Chain) NewStateTree(noBackLink bool, opts ...common.CallOp) (*common.Tree, error) {

	var prevRoot util.Hash

	// Get the root of the block at the tip. If no block was found, it means the chain is empty.
	// In this case, if the chain has a parent block, we use the parent block stateRoot.
	if !noBackLink {
		tipHeader, err := c.Current(opts...)
		if err != nil {
			if err != common.ErrBlockNotFound {
				return nil, err
			}
			if c.parentBlock != nil {
				prevRoot = c.parentBlock.Header.StateRoot
			}
		} else {
			prevRoot = tipHeader.StateRoot
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
func (c *Chain) PutTransactions(txs []*wire.Transaction, opts ...common.CallOp) error {
	return c.store.PutTransactions(txs, opts...)
}

// GetTransaction gets a transaction by hash
func (c *Chain) GetTransaction(hash string) *wire.Transaction {
	return c.store.GetTransaction(hash)
}
