package types

import "github.com/ellcrys/elld/wire"

// Blockchain defines an interface for a blockchain manager
type Blockchain interface {

	// Up initializes and loads the blockchain manager
	Up() error

	// IsKnownBlock checks if a block is stored in the main or side chain or orphan
	IsKnownBlock(hash string) (bool, string, error)

	// HaveBlock checks if a block exists on the main or side chains
	HaveBlock(hash string) (bool, error)

	// GetTransaction finds and returns a transaction on the main chain
	GetTransaction(hash string) (*wire.Transaction, error)

	// ProcessBlock attempts to process and append a block to the main or side chains
	ProcessBlock(block *wire.Block) (Chain, error)
}

// Chain defines an interface for a chain
type Chain interface {

	// GetID returns the id of the chain
	GetID() string
}
