package common

import (
	"github.com/ellcrys/elld/elldb"
	"github.com/ellcrys/elld/wire"
)

// ChainReader defines an interface for reading a chain
type ChainReader interface {

	// GetBlock finds and returns a block associated with chainID.
	// When 0 is passed, it should return the block with the highest number
	GetBlock(number uint64, opts ...CallOp) (*wire.Block, error)

	// GetBlockByHash finds and returns a block associated with chainID.
	GetBlockByHash(hash string, opts ...CallOp) (*wire.Block, error)

	// GetHeader gets the header of a block.
	// When 0 is passed, it should return the header of the block with the highest number
	GetHeader(number uint64, opts ...CallOp) (*wire.Header, error)

	// GetHeaderByHash finds and returns the header of a block matching hash
	GetHeaderByHash(hash string) (*wire.Header, error)

	// Current gets the current block at the tip of the chain
	Current(opts ...CallOp) (*wire.Block, error)
}

// ChainStorer defines an interface for storing objects
type ChainStorer interface {

	// PutBlock adds a block to the store
	PutBlock(block *wire.Block, opts ...CallOp) error

	// GetBlock finds and returns a block in the chain.
	// When 0 is passed, it should return the block with the highest number
	GetBlock(number uint64, opts ...CallOp) (*wire.Block, error)

	// GetBlockByHash finds and returns a block associated with chainID.
	GetBlockByHash(hash string, opts ...CallOp) (*wire.Block, error)

	// GetHeader gets the header of a block.
	// When 0 is passed, it should return the header of the block with the highest number
	GetHeader(number uint64, opts ...CallOp) (*wire.Header, error)

	// GetHeaderByHash finds and returns the header of a block matching hash
	GetHeaderByHash(hash string, opts ...CallOp) (*wire.Header, error)

	// GetTransaction gets a transaction (by hash) belonging to the chain
	GetTransaction(hash string, opts ...CallOp) *wire.Transaction

	// CreateAccount creates an account on a target block
	CreateAccount(targetBlockNum uint64, account *wire.Account, opts ...CallOp) error

	// GetAccount gets an account
	GetAccount(address string, opts ...CallOp) (*wire.Account, error)

	// PutTransactions stores a collection of transactions
	PutTransactions(txs []*wire.Transaction, opts ...CallOp) error

	// Current gets the current block at the tip of the chain
	Current(opts ...CallOp) (*wire.Block, error)

	// NewTx creates and returns a transaction
	NewTx() (elldb.Tx, error)
}
