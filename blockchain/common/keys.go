package common

import (
	"fmt"
	"strconv"

	"github.com/ellcrys/elld/database"
)

// MakeAccountKey constructs a key for an account
func MakeAccountKey(blockNum uint64, chainID, address string) []byte {
	bn := strconv.FormatUint(blockNum, 10)
	return database.MakeKey([]byte(bn), []string{chainID, "account", address})
}

// QueryAccountKey constructs a key for finding account data in the store and hash tree.
func QueryAccountKey(chainID, address string) []byte {
	return database.MakePrefix([]string{chainID, "account", address})
}

// MakeBlockchainMetadataKey constructs a key for storing blockchain-wide metadata
func MakeBlockchainMetadataKey() []byte {
	return database.MakeKey([]byte("_"), []string{"meta", "blockchain"})
}

// MakeBlockKey constructs a key for storing a block
func MakeBlockKey(chainID string, blockNumber uint64) []byte {
	key := []byte(fmt.Sprintf("%d", blockNumber))
	return database.MakeKey(key, []string{"chain", chainID, "block"})
}

// MakeBlocksQueryKey constructs a key for storing a block
func MakeBlocksQueryKey(chainID string) []byte {
	return database.MakeKey(nil, []string{"chain", chainID, "block"})
}

// MakeChainKey constructs a key for storing a chain information
func MakeChainKey(chainID string) []byte {
	return database.MakeKey([]byte(chainID), []string{"chainInfo"})
}

// MakeChainsQueryKey constructs a key for find all chain items
func MakeChainsQueryKey() []byte {
	return database.MakePrefix([]string{"chainInfo"})
}
