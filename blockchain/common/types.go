package common

import (
	"encoding/json"
	"time"

	"github.com/ellcrys/elld/elldb"
	"github.com/ellcrys/elld/wire"
)

// OrphanBlock represents an orphan block
type OrphanBlock struct {
	Block      *wire.Block
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

// Chainer defines an interface for Chain
type Chainer interface {

	// NewStateTree returns a new tree
	NewStateTree(noBackLink bool, opts ...CallOp) (*Tree, error)

	// GetTipHeader gets the header of the tip block
	GetTipHeader(opts ...CallOp) (*wire.Header, error)

	// GetID gets the chain ID
	GetID() string

	// GetBlock gets a block in the chain
	GetBlock(uint64) (*wire.Block, error)

	// GetParentBlock gets the chain's parent block if it has one
	GetParentBlock() *wire.Block

	// GetParentInfo gets the chain's parent information
	GetParentInfo() *ChainInfo
}

// Blockchain defines an interface for a blockchain manager
type Blockchain interface {

	// Up initializes and loads the blockchain manager
	Up() error

	// GetBestChain gets the chain that is currently considered the main chain
	GetBestChain() Chainer

	// IsKnownBlock checks if a block is stored in the main or side chain or orphan
	IsKnownBlock(hash string) (bool, string, error)

	// HaveBlock checks if a block exists on the main or side chains
	HaveBlock(hash string) (bool, error)

	// GetTransaction finds and returns a transaction on the main chain
	GetTransaction(hash string) (*wire.Transaction, error)

	// ProcessBlock attempts to process and append a block to the main or side chains
	ProcessBlock(*wire.Block) (Chainer, error)

	// GenerateBlock creates a new block for a target chain.
	// The Chain is specified by passing to ChainOp.
	GenerateBlock(*GenerateBlockParams, ...CallOp) (*wire.Block, error)
}
