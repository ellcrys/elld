package common

import (
	"encoding/json"
	"time"

	"github.com/ellcrys/elld/elldb"
	"github.com/ellcrys/elld/util"
)

// Block defines an interface for a block
type Block interface {

	// GetNumber returns the block number
	GetNumber() uint64

	// ComputeHash computes and returns the hash
	ComputeHash() util.Hash

	// GetHash returns the already computed hash
	GetHash() util.Hash
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
	Tx        elldb.Tx
	CanFinish bool
}

// GetName returns the name of the op
func (t TxOp) GetName() string {
	return "TxOp"
}

// Object represents an object that can be converted to JSON encoded byte slice
type Object interface {
	JSON() ([]byte, error)
}

// ChainInfo describes a chain
type ChainInfo struct {
	ID                string `json:"id"`
	ParentChainID     string `json:"parentChainID"`
	ParentBlockNumber uint64 `json:"parentBlockNumber"`
}

// BlockchainMeta includes information about the blockchain
type BlockchainMeta struct {
}

// JSON returns the JSON encoded equivalent
func (m *BlockchainMeta) JSON() ([]byte, error) {
	return json.Marshal(m)
}

// StateObject describes an object to be stored in a elldb.StateObject.
// Usually created after processing a Transition object.
type StateObject struct {

	// Key represents the key to use
	// to persist the object to database
	Key []byte

	// TreeKey represents the key to use
	// to add a record of this object in
	// a merkle tree
	TreeKey []byte

	// Value is the content of this state
	// object. It is written to the database
	// and the tree
	Value []byte
}
