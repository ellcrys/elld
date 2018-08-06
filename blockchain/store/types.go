package store

import (
	"github.com/ellcrys/elld/blockchain/common"
	"github.com/ellcrys/elld/elldb"
	"github.com/ellcrys/elld/wire"
)

// Reader defines an interface for reading a chain
type Reader interface {

	// GetBlock finds and returns a block associated with chainID.
	// When 0 is passed, it should return the block with the highest number
	GetBlock(number uint64, opts ...common.CallOp) (*wire.Block, error)

	// GetBlockByHash finds and returns a block associated with chainID.
	GetBlockByHash(hash string, opts ...common.CallOp) (*wire.Block, error)

	// GetBlockHeader gets the header of a block.
	// When 0 is passed, it should return the header of the block with the highest number
	GetBlockHeader(number uint64, opts ...common.CallOp) (*wire.Header, error)

	// GetBlockHeaderByHash finds and returns the header of a block matching hash
	GetBlockHeaderByHash(hash string) (*wire.Header, error)
}

// Storer defines an interface for storing objects
type Storer interface {

	// PutBlock adds a block to the store
	PutBlock(chainID string, block *wire.Block, opts ...common.CallOp) error

	// GetBlock finds and returns a block associated with chainID.
	// When 0 is passed, it should return the block with the highest number
	GetBlock(chainID string, number uint64, opts ...common.CallOp) (*wire.Block, error)

	// GetBlockByHash finds and returns a block associated with chainID.
	GetBlockByHash(chainID string, hash string, opts ...common.CallOp) (*wire.Block, error)

	// GetBlockHeader gets the header of a block.
	// When 0 is passed, it should return the header of the block with the highest number
	GetBlockHeader(chainID string, number uint64, opts ...common.CallOp) (*wire.Header, error)

	// GetBlockHeaderByHash finds and returns the header of a block matching hash
	GetBlockHeaderByHash(chainID string, hash string) (*wire.Header, error)

	// Put store an arbitrary value
	Put(key []byte, value []byte, opts ...common.CallOp) error

	// Get retrieves an arbitrary value
	Get(key []byte, result *[]*elldb.KVObject)

	// GetFirstOrLast returns the first or last object matching the key.
	// Set first to true to return the first or false for last.
	GetFirstOrLast(first bool, key []byte, result *elldb.KVObject)

	// NewTx creates and returns a transaction
	NewTx() (elldb.Tx, error)
}
