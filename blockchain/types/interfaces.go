package types

import "github.com/ellcrys/elld/wire"

// Block defines an interface for a block
type Block interface {

	// GetNumber returns the block number
	GetNumber() uint64
}

// Store defines an interface for storing objects and metadata
// of the blockchain.
type Store interface {

	// Initializes the store
	Initialize() error

	// PutBlock adds a block to the store
	PutBlock(block Block) error

	// GetBlock finds and returns a block
	GetBlock(number int64, block Block) error

	// GetCurrentBlockHeader gets the current/tail block header
	GetBlockHeader(number int64, header *wire.Header) error

	// DropDB deletes the database
	DropDB() error

	// Close closes the store, freeing resources held.
	Close() error
}
