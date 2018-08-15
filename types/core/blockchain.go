package core

import (
	"github.com/ellcrys/elld/util"
)

// Chainer defines an interface for Chain
type Chainer interface {

	// NewStateTree returns a new tree
	NewStateTree(noBackLink bool, opts ...CallOp) (Tree, error)

	// Current gets the header of the tip block
	Current(opts ...CallOp) (Header, error)

	// GetID gets the chain ID
	GetID() util.String

	// GetBlock gets a block in the chain
	GetBlock(uint64) (Block, error)

	// GetParentBlock gets the chain's parent block if it has one
	GetParentBlock() Block

	// GetParentInfo gets the chain's parent information
	GetParentInfo() *ChainInfo

	// CreateAccount creates an account on a target block
	CreateAccount(targetBlockNum uint64, account Account, opts ...CallOp) error

	// GetAccount gets an account
	GetAccount(address util.String, opts ...CallOp) (Account, error)

	// PutTransactions stores a collection of transactions
	PutTransactions(txs []Transaction, opts ...CallOp) error

	// GetTransaction gets a transaction by hash
	GetTransaction(hash util.Hash) Transaction
}

// Blockchain defines an interface for a blockchain manager
type Blockchain interface {

	// Up initializes and loads the blockchain manager
	Up() error

	// GetBestChain gets the chain that is currently considered the main chain
	GetBestChain() Chainer

	// IsKnownBlock checks if a block is stored in the main or side chain or orphan
	IsKnownBlock(hash util.Hash) (bool, string, error)

	// HaveBlock checks if a block exists on the main or side chains
	HaveBlock(hash util.Hash) (bool, error)

	// GetTransaction finds and returns a transaction on the main chain
	GetTransaction(hash util.Hash) (Transaction, error)

	// ProcessBlock attempts to process and append a block to the main or side chains
	ProcessBlock(Block) (ChainReader, error)

	// Generate creates a new block for a target chain.
	// The Chain is specified by passing to ChainOp.
	Generate(*GenerateBlockParams, ...CallOp) (Block, error)

	// ChainReader gets a Reader for reading the main chain
	ChainReader() ChainReader

	// GetChainsReader gets chain reader for all known chains
	GetChainsReader() (readers []ChainReader)
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
	GetHeaderByHash(hash util.Hash) (Header, error)

	// Current gets the current block at the tip of the chain
	Current(opts ...CallOp) (Block, error)
}
