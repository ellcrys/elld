package common

import (
	"fmt"
	"strconv"

	"github.com/ellcrys/elld/database"
)

// MakeAccountKey constructs a key for persisting account data in the store and hash tree.
func MakeAccountKey(blockNum uint64, chainID, address string) []byte {
	bn := strconv.FormatUint(blockNum, 10)
	return database.MakeKey([]byte(bn), []string{chainID, "account", address})
}

// QueryAccountKey constructs a key for finding account data in the store and hash tree.
func QueryAccountKey(chainID, address string) []byte {
	return database.MakePrefix([]string{chainID, "account", address})
}

// MakeChainMetadataKey constructs a key for storing chain metadata
func MakeChainMetadataKey(chainID string) []byte {
	return database.MakeKey([]byte(chainID), []string{"meta", "chain"})
}

// MakeBlockchainMetadataKey constructs a key for storing blockchain-wide metadata
func MakeBlockchainMetadataKey() []byte {
	return database.MakeKey([]byte("_"), []string{"meta", "blockchain"})
}

// MakeBlockKey constructs a key for storing a block
func MakeBlockKey(chainID string, blockNo uint64) []byte {
	key := []byte(fmt.Sprintf("%d", blockNo))
	return database.MakeKey(key, []string{"block", chainID, "number"})
}

// MakeBlockHashKey constructs a key for storing a block using the block hash
func MakeBlockHashKey(chainID string, blockHash string) []byte {
	key := []byte(blockHash)
	return database.MakeKey(key, []string{"block", chainID, "hash"})
}

// MakeChainKey constructs a key for storing a chain
func MakeChainKey(chainID string) []byte {
	return database.MakeKey([]byte(chainID), []string{"chain"})
}
