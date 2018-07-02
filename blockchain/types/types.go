package types

import "github.com/ellcrys/elld/wire"

// Block defines an interface for a block
type Block interface {

	// GetNumber returns the block number
	GetNumber() uint64

	// GetHash returns the hash
	GetHash() string
}

// Store defines an interface for storing objects and metadata
// of the blockchain.
type Store interface {

	// GetMetadata returns the store's metadata
	GetMetadata(*Meta) error

	// UpdateMetadata updates the store's metadata
	UpdateMetadata(*Meta) error

	// PutBlock adds a block to the store
	PutBlock(block Block) error

	// GetBlock finds and returns a block
	// When 0 is passed, it should return the block with the highest number
	GetBlock(number uint64, block Block) error

	// GetBlockByHash finds and returns a block
	GetBlockByHash(hash string, block Block) error

	// GetCurrentBlockHeader gets the current/tail block header.
	// When 0 is passed, it should return the header of the block with the highest number
	GetBlockHeader(number uint64, header *wire.Header) error
}

// Meta includes information about a store
type Meta struct {
	CurrentBlockNumber uint64 `json:"curBlock"` // The number of the current block
}
