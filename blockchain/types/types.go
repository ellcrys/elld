package types

import (
	"encoding/json"

	"github.com/ellcrys/elld/wire"
)

// Block defines an interface for a block
type Block interface {

	// GetNumber returns the block number
	GetNumber() uint64

	// ComputeHash returns the hash
	ComputeHash() string
}

// Store defines an interface for storing objects and metadata
// of the blockchain.
type Store interface {

	// GetMetadata returns the store's metadata
	GetMetadata(name string, m Object) error

	// UpdateMetadata updates the store's metadata
	UpdateMetadata(name string, m Object) error

	// PutBlock adds a block to the store
	PutBlock(chainID string, block Block) error

	// GetBlock finds and returns a block associated with chainID.
	// When 0 is passed, it should return the block with the highest number
	GetBlock(chainID string, number uint64, block Block) error

	// GetBlockByHash finds and returns a block associated with chainID.
	GetBlockByHash(chainID string, hash string, block Block) error

	// GetBlockHeader gets the header of a block.
	// When 0 is passed, it should return the header of the block with the highest number
	GetBlockHeader(chainID string, number uint64, header *wire.Header) error

	// GetBlockHeaderByHash finds and returns the header of a block matching hash
	GetBlockHeaderByHash(chainID string, hash string, header *wire.Header) error

	// Put store an arbitrary value
	Put(key []byte, value []byte) error

	// Get retrieves an arbitrary value
	Get(key []byte, result interface{}) error
}

// Object represents an object that can be converted to JSON encoded byte slice
type Object interface {
	JSON() ([]byte, error)
}

// ChainMeta includes information about a chain
type ChainMeta struct {
	CurrentBlockNumber uint64 `json:"curBlock"` // The number of the current block
}

// JSON returns the JSON encoded equivalent
func (m *ChainMeta) JSON() ([]byte, error) {
	return json.Marshal(m)
}

// BlockchainMeta includes information about the blockchain
type BlockchainMeta struct {
	Chains []string `json:"chains"` // Contains the ID of all known chains
}

// JSON returns the JSON encoded equivalent
func (m *BlockchainMeta) JSON() ([]byte, error) {
	return json.Marshal(m)
}
