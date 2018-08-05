package common

import (
	"fmt"
	"strconv"

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

// MakeAccountKey constructs a key for an account
func MakeAccountKey(blockNum uint64, chainID, address string) []byte {
	bn := strconv.FormatUint(blockNum, 10)
	return elldb.MakeKey([]byte(bn), []string{ObjectTypeChain, chainID, ObjectTypeAccount, address})
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
	key := []byte(fmt.Sprintf("%d", blockNumber))
	return elldb.MakeKey(key, []string{ObjectTypeChain, chainID, ObjectTypeBlock})
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
	bn := []byte(fmt.Sprintf("%d", blockNumber))
	return append([]byte(bn), []byte(objectType)...)
}
