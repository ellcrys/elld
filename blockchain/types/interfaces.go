package types

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
	GetBlock(number uint64, block Block) error

	// Close closes the store, freeing resources held.
	Close() error
}
