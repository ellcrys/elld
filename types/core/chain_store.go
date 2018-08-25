package core

import (
	"github.com/ellcrys/elld/elldb"
	"github.com/ellcrys/elld/util"
)

// ChainStorer defines an interface that provides
// every functionality need to mutate or query a 
// chain. 
type ChainStorer interface {

	// PutBlock adds a block to the store
	PutBlock(block Block, opts ...CallOp) error

	// GetBlock finds and returns a block in the chain.
	// When 0 is passed, it should return the block with the highest number
	GetBlock(number uint64, opts ...CallOp) (Block, error)

	// GetBlockByHash finds a block by its hash
	GetBlockByHash(hash util.Hash, opts ...CallOp) (Block, error)

	// GetBlockByNumberAndHash finds by number and hash
	GetBlockByNumberAndHash(number uint64, hash util.Hash, opts ...CallOp) (Block, error)

	// GetHeader gets the header of a block.
	// When 0 is passed, it should return the header of the block with the highest number
	GetHeader(number uint64, opts ...CallOp) (Header, error)

	// GetHeaderByHash finds and returns the header of a block matching hash
	GetHeaderByHash(hash util.Hash, opts ...CallOp) (Header, error)

	// GetTransaction gets a transaction (by hash) belonging to the chain
	GetTransaction(hash util.Hash, opts ...CallOp) Transaction

	// CreateAccount creates an account on a target block
	CreateAccount(targetBlockNum uint64, account Account, opts ...CallOp) error

	// GetAccount gets an account
	GetAccount(address util.String, opts ...CallOp) (Account, error)

	// PutTransactions stores a collection of transactions
	PutTransactions(txs []Transaction, blockNumber uint64, opts ...CallOp) error

	// Current gets the current block at the tip of the chain
	Current(opts ...CallOp) (Block, error)

	// Delete deletes objects
	Delete(key []byte, opts ...CallOp) error

	// NewTx creates and returns a transaction
	NewTx() (elldb.Tx, error)

	// DB gets the database
	DB() elldb.DB
}
