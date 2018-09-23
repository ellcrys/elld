package core

import (
	"math/big"

	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/merkletree"
	"github.com/olebedev/emitter"
)

// CallOp describes an interface to be used to define store method options
type CallOp interface {
	GetName() string
}

// Tree defines a merkle tree
type Tree interface {
	Add(item merkletree.Content)
	GetItems() []merkletree.Content
	Build() error
	Root() util.Hash
}

// ChainInfo describes a chain
type ChainInfo struct {
	ID                util.String `json:"id" msgpack:"json"`
	ParentChainID     util.String `json:"parentChainID" msgpack:"parentChainID"`
	ParentBlockNumber uint64      `json:"parentBlockNumber" msgpack:"parentBlockNumber"`
	Timestamp         int64       `json:"timestamp" msgpack:"timestamp"`
}

// TxContainer represents a container
// a container responsible for holding
// and sorting transactions
type TxContainer interface {
	ByteSize() int64
	Add(tx Transaction) bool
	Has(tx Transaction) bool
	Size() int64
	First() Transaction
	Last() Transaction
	Sort()
	IFind(predicate func(Transaction) bool) Transaction
	Remove(txs ...Transaction)
}

// TxPool represents a transactions pool
type TxPool interface {
	SetEventEmitter(ee *emitter.Emitter)
	Put(tx Transaction) error
	PutSilently(tx Transaction) error
	Has(tx Transaction) bool
	HasByHash(hash string) bool
	SenderHasTxWithSameNonce(address util.String, nonce uint64) bool
	ByteSize() int64
	Size() int64
	Container() TxContainer
}

// GenerateBlockParams represents parameters
// required for block generation.
type GenerateBlockParams struct {

	// OverrideParentHash explicitly sets the parent hash
	OverrideParentHash util.Hash

	// Transactions sets the block transaction.
	// If not provided, transactions are selected from
	// the transaction pool
	Transactions []Transaction

	// Creator sets the key of the block creator.
	// Required for setting the creator public key
	// and signing the block
	Creator *crypto.Key

	// Nonce is the special number that
	// indicates the completion of PoW
	Nonce BlockNonce

	// Difficulty represents the target
	// difficulty that the nonce satisfied
	Difficulty *big.Int

	// OverrideTotalDifficulty explicitly sets
	// the total difficulty.
	OverrideTotalDifficulty *big.Int

	// StateRoot sets the state root
	OverrideStateRoot util.Hash

	// Timestamp sets the time of block creation
	OverrideTimestamp int64

	// ChainTip explicitly sets the chain tip number.
	// This is used to alter what the tip of a chain is.
	// It is used in tests.
	OverrideChainTip uint64

	// AddFeeAlloc if set to true, will calculate the
	// miner fee reward and add an Alloc transaction
	AddFeeAlloc bool
}
