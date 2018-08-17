package core

import (
	"encoding/binary"
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
	SetChainReader(cr ChainReader)
	GetChainReader() ChainReader
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
	GetNonce() BlockNonce
	SetNonce(nonce BlockNonce)
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
}

// Transaction represents a transaction
type Transaction interface {
	GetHash() util.Hash
	Bytes() []byte
	ComputeHash() util.Hash
	ID() string
	Sign(privKey string) ([]byte, error)
	GetType() int64
	GetFrom() util.String
	SetFrom(util.String)
	GetTo() util.String
	GetValue() util.String
	GetFee() util.String
	GetNonce() int64
	GetTimestamp() int64
	GetSenderPubKey() util.String
	GetSignature() []byte
}

// A BlockNonce is a 64-bit hash which proves (combined with the
// mix-hash) that a sufficient amount of computation has been carried
// out on a block.
type BlockNonce [8]byte

// EmptyBlockNonce is a BlockNonce with no values
var EmptyBlockNonce = BlockNonce([8]byte{})

// EncodeNonce converts the given integer to a block nonce.
func EncodeNonce(i uint64) BlockNonce {
	var n BlockNonce
	binary.BigEndian.PutUint64(n[:], i)
	return n
}

// Uint64 returns the integer value of a block nonce.
func (n BlockNonce) Uint64() uint64 {
	return binary.BigEndian.Uint64(n[:])
}

// MarshalText encodes n as a hex string with 0x prefix.
func (n BlockNonce) MarshalText() string {
	return util.ToHex(n[:])
}
