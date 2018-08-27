package common

import (
	"encoding/binary"
	"strconv"

	"github.com/ellcrys/elld/elldb"
)

var (
	// ObjectTypeAccount represents an account object
	ObjectTypeAccount = []byte("account")

	// ObjectTypeChain represents a chain object
	ObjectTypeChain = []byte("chain")

	// ObjectTypeBlock represents a block object
	ObjectTypeBlock = []byte("block")

	// ObjectTypeChainInfo represents a chain information object
	ObjectTypeChainInfo = []byte("chainInfo")

	// ObjectTypeTransaction represents a transaction object
	ObjectTypeTransaction = []byte("tx")

	// ObjectTypeBlockchainMeta represents a meta object
	ObjectTypeBlockchainMeta = []byte("blockchain-meta")

	// ObjectTypeReOrg represents a meta object
	ObjectTypeReOrg = []byte("re-org")
)

// EncodeBlockNumber serializes a block number to BigEndian
func EncodeBlockNumber(n uint64) []byte {
	var b = make([]byte, 8)
	binary.BigEndian.PutUint64(b, n)
	return b
}

// DecodeBlockNumber deserialize a block number from BigEndian
func DecodeBlockNumber(encNum []byte) uint64 {
	return binary.BigEndian.Uint64(encNum)
}

// MakeAccountKey constructs a key for storing/querying an account
func MakeAccountKey(blockNum uint64, chainID, address []byte) []byte {
	return elldb.MakeKey(
		EncodeBlockNumber(blockNum),
		ObjectTypeChain,
		chainID,
		ObjectTypeAccount,
		address,
	)
}

// MakeAccountsKey constructs a key for querying all accounts
func MakeAccountsKey(chainID []byte) []byte {
	return elldb.MakePrefix(
		ObjectTypeChain,
		chainID,
		ObjectTypeAccount,
	)
}

// QueryAccountKey constructs a key for finding account data in the store and hash tree.
func QueryAccountKey(chainID, address []byte) []byte {
	return elldb.MakePrefix(
		ObjectTypeChain,
		chainID,
		ObjectTypeAccount,
		address,
	)
}

// MakeBlockchainMetadataKey constructs a key for storing blockchain-wide metadata
func MakeBlockchainMetadataKey() []byte {
	return elldb.MakeKey(ObjectTypeBlockchainMeta)
}

// MakeBlockKey constructs a key for storing a block
func MakeBlockKey(chainID []byte, blockNumber uint64) []byte {
	return elldb.MakeKey(
		EncodeBlockNumber(blockNumber),
		ObjectTypeChain,
		chainID,
		ObjectTypeBlock,
	)
}

// MakeBlocksQueryKey constructs a key for querying all blocks
func MakeBlocksQueryKey(chainID []byte) []byte {
	return elldb.MakeKey(nil,
		ObjectTypeChain,
		chainID,
		ObjectTypeBlock,
	)
}

// MakeChainKey constructs a key for storingchain information
func MakeChainKey(chainID []byte) []byte {
	return elldb.MakeKey(chainID, ObjectTypeChainInfo)
}

// MakeChainsQueryKey constructs a key for find all chain items
func MakeChainsQueryKey() []byte {
	return elldb.MakePrefix(ObjectTypeChainInfo)
}

// MakeTxKey constructs a key for storing a transaction
func MakeTxKey(chainID []byte, blockNumber uint64, txHash []byte) []byte {
	return elldb.MakeKey(
		EncodeBlockNumber(blockNumber),
		ObjectTypeChain,
		chainID,
		ObjectTypeTransaction,
		txHash,
	)
}

// MakeTxQueryKey constructs a key for querying a transaction
func MakeTxQueryKey(chainID []byte, txHash []byte) []byte {
	return elldb.MakePrefix(
		ObjectTypeChain,
		chainID,
		ObjectTypeTransaction,
		txHash,
	)
}

// MakeTxsQueryKey constructs a key for querying all transactions in a chain
func MakeTxsQueryKey(chainID []byte) []byte {
	return elldb.MakePrefix(
		ObjectTypeChain,
		chainID,
		ObjectTypeTransaction,
	)
}

// MakeTreeKey constructs a key for recording state objects in a tree
func MakeTreeKey(blockNumber uint64, objectType []byte) []byte {
	return append(EncodeBlockNumber(blockNumber), objectType...)
}

// MakeReOrgKey constructs a key for storing reorganization info
func MakeReOrgKey(timestamp int64) []byte {
	return elldb.MakeKey(
		[]byte(strconv.FormatInt(timestamp, 10)),
		ObjectTypeReOrg,
	)
}

// MakeReOrgQueryKey constructs a key for querying reorganization objects
func MakeReOrgQueryKey() []byte {
	return elldb.MakePrefix(
		ObjectTypeReOrg,
	)
}
