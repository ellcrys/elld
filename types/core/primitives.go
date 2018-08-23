package core

import (
	"encoding/binary"
	"fmt"
	"math/big"

	"github.com/ellcrys/elld/util"
	"github.com/fatih/structs"
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

// MapFieldsToHex takes a struct and converts
// selected field types to hex string. It returns
// a map. It will panic if obj is not a map/struct.
func MapFieldsToHex(obj interface{}) interface{} {

	if obj == nil {
		return obj
	}

	var m map[string]interface{}

	// if not struct, we assume it is a map
	if structs.IsStruct(obj) {
		s := structs.New(obj)
		s.TagName = "json"
		m = s.Map()
	} else {
		m = obj.(map[string]interface{})
	}

	for k, v := range m {
		switch _v := v.(type) {
		case BlockNonce:
			m[k] = util.BytesToHash(_v[:]).HexStr()
		case util.Hash:
			m[k] = _v.HexStr()
		case *big.Int:
			m[k] = fmt.Sprintf("0x%x", _v)
		case []byte:
			m[k] = util.BytesToHash(_v).HexStr()
		case map[string]interface{}:
			m[k] = MapFieldsToHex(_v)
		case []interface{}:
			for i, item := range _v {
				_v[i] = MapFieldsToHex(item)
			}
		}
	}

	return m
}
