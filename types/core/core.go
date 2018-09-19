package core

import (
	"github.com/ellcrys/elld/elldb"
	"github.com/ellcrys/elld/util"
	"github.com/olebedev/emitter"
)

// Chainer (a.k.a Chains) defines an interface for accessing
// mutating and managing a collection of blocks
type Chainer interface {

	// GetStore returns the store
	GetStore() ChainStorer

	// NewStateTree returns a new tree
	NewStateTree(opts ...CallOp) (Tree, error)

	// Current gets the header of the tip block
	Current(opts ...CallOp) (Header, error)

	// GetID gets the chain ID
	GetID() util.String

	// GetBlock gets a block in the chain
	GetBlock(uint64, ...CallOp) (Block, error)

	// GetParentBlock gets the chain's parent block if it has one
	GetParentBlock() Block

	// GetInfo gets the chain's parent information
	GetInfo() *ChainInfo

	// CreateAccount creates an account on a target block
	CreateAccount(targetBlockNum uint64, account Account, opts ...CallOp) error

	// GetAccount gets an account
	GetAccount(address util.String, opts ...CallOp) (Account, error)

	// PutTransactions stores a collection of transactions
	PutTransactions(txs []Transaction, blockNumber uint64, opts ...CallOp) error

	// GetTransaction gets a transaction by hash
	GetTransaction(hash util.Hash, opts ...CallOp) (Transaction, error)

	// ChainReader gets a chain reader for this chain
	ChainReader() ChainReader

	// GetRoot fetches the root block of this chain. If the chain
	// has more than one parents/ancestors, it will traverse
	// the parents to return the root parent block.
	GetRoot() Block
}

// Blockchain defines an interface for a blockchain manager
type Blockchain interface {

	// Up initializes and loads the blockchain manager
	Up() error

	// GetBestChain gets the chain that is currently considered the main chain
	GetBestChain() Chainer

	// IsKnownBlock checks if a block is stored in the main or side chain or orphan
	IsKnownBlock(hash util.Hash) (bool, string, error)

	// HaveBlock checks whether we have a block matching the hash in any of the known chains
	HaveBlock(hash util.Hash) (bool, error)

	// GetTransaction finds and returns a transaction on the main chain
	GetTransaction(util.Hash, ...CallOp) (Transaction, error)

	// ProcessBlock attempts to process and append a block to the main or side chains
	ProcessBlock(Block) (ChainReader, error)

	// Generate creates a new block for a target chain.
	// The Chain is specified by passing to ChainOp.
	Generate(*GenerateBlockParams, ...CallOp) (Block, error)

	// ChainReader gets a Reader for reading the main chain
	ChainReader() ChainReader

	// GetChainsReader gets chain reader for all known chains
	GetChainsReader() (readers []ChainReader)

	// SetDB sets the database
	SetDB(elldb.DB)

	// OrphanBlocks gets a reader for the orphan cache
	OrphanBlocks() CacheReader

	// GetEventEmitter gets the event emitter
	GetEventEmitter() *emitter.Emitter

	// GetBlock finds a block in any chain with a matching
	// block number and hash.
	GetBlock(number uint64, hash util.Hash) (Block, error)

	// GetBlockByHash finds a block in any chain with a matching hash.
	GetBlockByHash(hash util.Hash, opts ...CallOp) (Block, error)

	// GetChainReaderByHash returns a chain reader to a chain
	// where a block with the given hash exists
	GetChainReaderByHash(hash util.Hash) ChainReader

	// SetGenesisBlock sets the genesis block
	SetGenesisBlock(block Block)

	// GetTxPool gets the transaction pool
	GetTxPool() TxPool

	// CreateAccount creates an account that is associated with
	// the given block number and chain.
	CreateAccount(blockNo uint64, chain Chainer, account Account) error

	// GetAccount gets an account
	GetAccount(address util.String, opts ...CallOp) (Account, error)

	// GetAccountNonce gets the nonce of an account
	GetAccountNonce(address util.String, opts ...CallOp) (uint64, error)
}

// BlockMaker defines an interface providing the
// necessary functions to create new blocks
type BlockMaker interface {

	// Generate creates a new block for a target chain.
	// The Chain is specified by passing to ChainOp.
	Generate(*GenerateBlockParams, ...CallOp) (Block, error)

	// ChainReader gets a Reader for reading the main chain
	ChainReader() ChainReader

	// ProcessBlock attempts to process and append a block to the main or side chains
	ProcessBlock(Block) (ChainReader, error)

	// IsMainChain checks whether a chain is the main chain
	IsMainChain(ChainReader) bool
}

// ChainReader defines an interface for reading a chain
type ChainReader interface {

	// GetID gets the chain ID
	GetID() util.String

	// GetBlock finds and returns a block associated with chainID.
	// When 0 is passed, it should return the block with the highest number
	GetBlock(number uint64, opts ...CallOp) (Block, error)

	// GetBlockByHash finds and returns a block associated with chainID.
	GetBlockByHash(hash util.Hash, opts ...CallOp) (Block, error)

	// GetHeader gets the header of a block.
	// When 0 is passed, it should return the header of the block with the highest number
	GetHeader(number uint64, opts ...CallOp) (Header, error)

	// GetHeaderByHash finds and returns the header of a block matching hash
	GetHeaderByHash(hash util.Hash, opts ...CallOp) (Header, error)

	// Current gets the current block at the tip of the chain
	Current(opts ...CallOp) (Block, error)

	// GetParent returns a chain reader to the parent chain.
	// Returns nil if chain has no parent.
	GetParent() ChainReader

	// GetParentBlock returns the parent block
	GetParentBlock() Block

	// GetRoot fetches the root block of this chain. If the chain
	// has more than one parents/ancestors, it will traverse
	// the parents to return the root parent block.
	GetRoot() Block
}

// CacheReader provides an interface for reading the orphan cache
type CacheReader interface {

	// Len gets the number of orphans
	Len() int

	// Hash checks whether an item exists in the cache
	Has(key interface{}) bool

	// Get gets an item from the cache
	Get(key interface{}) interface{}
}
