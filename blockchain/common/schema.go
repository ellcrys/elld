package common

import (
	"github.com/vmihailenco/msgpack"

	"github.com/ellcrys/elld/elldb"
)

const (
	// ObjectTypeAccount represents an account object
	ObjectTypeAccount = "account"

	// ObjectTypeChain represents a chain object
	ObjectTypeChain = "chain"

	// ObjectTypeBlock represents a block object
	ObjectTypeBlock = "block"

	// ObjectTypeChainInfo represents a chain information object
	ObjectTypeChainInfo = "chainInfo"

	// ObjectTypeTransaction represents a transaction object
	ObjectTypeTransaction = "tx"

	// ObjectTypeMeta represents a meta object
	ObjectTypeMeta = "meta"
)

// EncodeBlockNumber serializes a block number
func EncodeBlockNumber(n uint64) []byte {
	b, _ := msgpack.Marshal(n)
	return b
}

// DecodeBlockNumber deserializes a block number
func DecodeBlockNumber(encNum []byte) uint64 {
	var bn uint64
	msgpack.Unmarshal(encNum, &bn)
	return bn
}

// MakeAccountKey constructs a key for an account
func MakeAccountKey(blockNum uint64, chainID, address string) []byte {
	return elldb.MakeKey(EncodeBlockNumber(blockNum), []string{ObjectTypeChain, chainID, ObjectTypeAccount, address})
}

// QueryAccountKey constructs a key for finding account data in the store and hash tree.
func QueryAccountKey(chainID, address string) []byte {
	return elldb.MakePrefix([]string{ObjectTypeChain, chainID, ObjectTypeAccount, address})
}

// MakeBlockchainMetadataKey constructs a key for storing blockchain-wide metadata
func MakeBlockchainMetadataKey() []byte {
	return elldb.MakeKey([]byte("_"), []string{ObjectTypeMeta, "blockchain"})
}

// MakeBlockKey constructs a key for storing a block
func MakeBlockKey(chainID string, blockNumber uint64) []byte {
	return elldb.MakeKey(EncodeBlockNumber(blockNumber), []string{ObjectTypeChain, chainID, ObjectTypeBlock})
}

// MakeBlocksQueryKey constructs a key for storing a block
func MakeBlocksQueryKey(chainID string) []byte {
	return elldb.MakeKey(nil, []string{ObjectTypeChain, chainID, ObjectTypeBlock})
}

// MakeChainKey constructs a key for storing a chain information
func MakeChainKey(chainID string) []byte {
	return elldb.MakeKey([]byte(chainID), []string{ObjectTypeChainInfo})
}

// MakeChainsQueryKey constructs a key for find all chain items
func MakeChainsQueryKey() []byte {
	return elldb.MakePrefix([]string{ObjectTypeChainInfo})
}

// MakeTxKey constructs a key for storing a transaction
func MakeTxKey(chainID, txHash string) []byte {
	return elldb.MakeKey(nil, []string{ObjectTypeChain, chainID, ObjectTypeTransaction, txHash})
}

// MakeTreeKey constructs a key for recording state objects in a tree
func MakeTreeKey(blockNumber uint64, objectType string) []byte {
	return append(EncodeBlockNumber(blockNumber), []byte(objectType)...)
}
