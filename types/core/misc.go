package core

import (
	"math/big"

	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/merkletree"
	"github.com/vmihailenco/msgpack"
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

// BlockchainMeta includes information about the blockchain
type BlockchainMeta struct {
}

// JSON returns the JSON encoded equivalent
func (m *BlockchainMeta) JSON() ([]byte, error) {
	return msgpack.Marshal(m)
}

// GenerateBlockParams represents parameters
// required for block generation.
type GenerateBlockParams struct {
	OverrideParentHash      util.Hash
	Transactions            []Transaction
	Creator                 *crypto.Key
	Nonce                   BlockNonce
	Difficulty              *big.Int
	OverrideTotalDifficulty *big.Int
	OverrideStateRoot       util.Hash
	OverrideTimestamp       int64
	OverrideChainTip        uint64
}
