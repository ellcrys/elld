package common

import (
	"encoding/json"
	"time"

	"github.com/ellcrys/elld/database"
	"github.com/ellcrys/elld/wire"
)

// Block defines an interface for a block
type Block interface {

	// GetNumber returns the block number
	GetNumber() uint64

	// ComputeHash computes and returns the hash
	ComputeHash() string

	// GetHash returns the already computed hash
	GetHash() string
}

// OrphanBlock represents an orphan block
type OrphanBlock struct {
	Block      Block
	Expiration time.Time
}

// CallOp describes an interface to be used to define store method options
type CallOp interface {
	GetName() string
}

// TxOp defines a method option for passing external transactions
type TxOp struct {
	Tx        database.Tx
	CanFinish bool
}

// GetName returns the name of the op
func (t TxOp) GetName() string {
	return "TxOp"
}

// Store defines an interface for storing objects and metadata
// of the blockchain.
type Store interface {

	// PutBlock adds a block to the store
	PutBlock(chainID string, block *wire.Block, opts ...CallOp) error

	// GetBlock finds and returns a block associated with chainID.
	// When 0 is passed, it should return the block with the highest number
	GetBlock(chainID string, number uint64, opts ...CallOp) (*wire.Block, error)

	// GetBlockByHash finds and returns a block associated with chainID.
	GetBlockByHash(chainID string, hash string, opts ...CallOp) (*wire.Block, error)

	// GetBlockHeader gets the header of a block.
	// When 0 is passed, it should return the header of the block with the highest number
	GetBlockHeader(chainID string, number uint64) (*wire.Header, error)

	// GetBlockHeaderByHash finds and returns the header of a block matching hash
	GetBlockHeaderByHash(chainID string, hash string) (*wire.Header, error)

	// Put store an arbitrary value
	Put(key []byte, value []byte, opts ...CallOp) error

	// Get retrieves an arbitrary value
	Get(key []byte, result *[]*database.KVObject)

	// GetFirstOrLast returns the first or last object matching the key.
	// Set first to true to return the first or false for last.
	GetFirstOrLast(first bool, key []byte, result *database.KVObject)

	// NewTx creates and returns a transaction
	NewTx() database.Tx
}

// Object represents an object that can be converted to JSON encoded byte slice
type Object interface {
	JSON() ([]byte, error)
}

// ChainInfo describes a chain
type ChainInfo struct {
	ID           string `json:"id"`
	ParentNumber uint64 `json:"parentNumber"`
}

// BlockchainMeta includes information about the blockchain
type BlockchainMeta struct {
}

// JSON returns the JSON encoded equivalent
func (m *BlockchainMeta) JSON() ([]byte, error) {
	return json.Marshal(m)
}

// StateObject describes an object to be stored in a database.StateObject.
// Usually created after processing a Transition object.
type StateObject struct {
	Key   []byte
	Value []byte
}
