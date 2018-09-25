package core

import (
	"math/big"

	"github.com/ellcrys/elld/util"
)

// Block defines a block
type Block interface {
	ComputeHash() util.Hash
	Bytes() []byte
	GetHeader() Header
	SetHeader(h Header)
	WithSeal(header Header) Block
	GetNumber() uint64
	GetHash() util.Hash
	SetHash(util.Hash)
	GetTransactions() []Transaction
	GetSignature() []byte
	SetSignature(sig []byte)
	HashToHex() string
}

// Header defines a block header containing
// metadata about the block
type Header interface {
	GetNumber() uint64
	SetNumber(uint64)
	HashNoNonce() util.Hash
	Bytes() []byte
	ComputeHash() util.Hash
	GetExtra() []byte
	GetTimestamp() int64
	SetTimestamp(int64)
	GetDifficulty() *big.Int
	SetDifficulty(*big.Int)
	GetNonce() util.BlockNonce
	SetNonce(nonce util.BlockNonce)
	GetParentHash() util.Hash
	SetParentHash(util.Hash)
	GetCreatorPubKey() util.String
	Copy() Header
	SetStateRoot(util.Hash)
	GetStateRoot() util.Hash
	SetTransactionsRoot(txRoot util.Hash)
	GetTransactionsRoot() util.Hash
	GetTotalDifficulty() *big.Int
	SetTotalDifficulty(*big.Int)
}

// Account defines an interface for an account
type Account interface {
	GetAddress() util.String
	GetBalance() util.String
	SetBalance(util.String)
	GetNonce() uint64
	IncrNonce()
}

// Transaction represents a transaction
type Transaction interface {
	GetHash() util.Hash
	SetHash(util.Hash)
	Bytes() []byte
	SizeNoFee() int64
	ComputeHash() util.Hash
	ID() string
	Sign(privKey string) ([]byte, error)
	GetType() int64
	GetFrom() util.String
	SetFrom(util.String)
	GetTo() util.String
	GetValue() util.String
	SetValue(util.String)
	GetFee() util.String
	GetNonce() uint64
	GetTimestamp() int64
	GetSenderPubKey() util.String
	SetSenderPubKey(util.String)
	GetSignature() []byte
	SetSignature(sig []byte)
}
