// This file provides methods for constructing database
// objects describing various state values like accounts,
// chains, transactions etc.
//
// When defining a new method to store a new object type,
// ensure to store the block number as the 'key' and every
// other identifiers as prefixes. We can also store any
// numeric value that might be relied upon for ordering
// in the 'key' field (e.g timestamp). The 'key' value must
// be encoded as a big-endian.

package common

import (
	"github.com/ellcrys/elld/elldb"
	"github.com/ellcrys/elld/util"
)

var (
	// TagAccount represents an account object
	TagAccount = []byte("a")

	// TagChain represents a chain object
	TagChain = []byte("c")

	// TagBlock represents a block object
	TagBlock = []byte("b")

	// TagChainInfo represents a chain information object
	TagChainInfo = []byte("i")

	// TagTransaction represents a transaction object
	TagTransaction = []byte("t")

	// TagReOrg represents a meta object
	TagReOrg = []byte("r")
)

// MakeKeyAccount constructs a key for storing an account.
// Prefixes: tag_chain + chain ID + tag_account + address +
// block number (big endian)
func MakeKeyAccount(blockNum uint64, chainID, address []byte) []byte {
	return elldb.MakeKey(
		util.EncodeNumber(blockNum),
		TagChain,
		chainID,
		TagAccount,
		address,
	)
}

// MakeQueryKeyAccounts constructs a key for querying all
// accounts in a given chain.
// Prefixes: tag_chain + chain ID + tag_account
func MakeQueryKeyAccounts(chainID []byte) []byte {
	return elldb.MakePrefix(
		TagChain,
		chainID,
		TagAccount,
	)
}

// MakeQueryKeyAccount constructs a key for
// finding account data in the store and
// hash tree.
// Prefixes: tag_chain + chain ID + tag_account + address
func MakeQueryKeyAccount(chainID, address []byte) []byte {
	return elldb.MakePrefix(
		TagChain,
		chainID,
		TagAccount,
		address,
	)
}

// MakeKeyBlock constructs a key for storing a block.
// Prefixes: tag_chain + chain ID + tag_block +
// block number (big endian)
func MakeKeyBlock(chainID []byte, blockNumber uint64) []byte {
	return elldb.MakeKey(
		util.EncodeNumber(blockNumber),
		TagChain,
		chainID,
		TagBlock,
	)
}

// MakeQueryKeyBlocks constructs a key for querying
// all blocks in given chain.
// Prefixes: tag_chain + chain ID + tag_block
func MakeQueryKeyBlocks(chainID []byte) []byte {
	return elldb.MakePrefix(
		TagChain,
		chainID,
		TagBlock,
	)
}

// MakeKeyChain constructs a key for storing chain
// information.
// Prefixes: tag_chain_info + chain ID
func MakeKeyChain(chainID []byte) []byte {
	return elldb.MakePrefix(
		TagChainInfo,
		chainID,
	)
}

// MakeQueryKeyChains constructs a key to
// find all chain information.
// Prefixes: tag_chain_info
func MakeQueryKeyChains() []byte {
	return elldb.MakePrefix(TagChainInfo)
}

// MakeKeyTransaction constructs a key for storing a transaction.
// Prefixes: tag_chain + chain ID + tag_transaction +
// transaction hash + block number (big endian)
func MakeKeyTransaction(chainID []byte, blockNumber uint64, txHash []byte) []byte {
	return elldb.MakeKey(
		util.EncodeNumber(blockNumber),
		TagChain,
		chainID,
		TagTransaction,
		txHash,
	)
}

// MakeTxQueryKey constructs a key for querying a transaction.
// Prefixes: tag_chain + chain ID + tag_transaction +
// transaction hash
func MakeTxQueryKey(chainID []byte, txHash []byte) []byte {
	return elldb.MakePrefix(
		TagChain,
		chainID,
		TagTransaction,
		txHash,
	)
}

// MakeQueryKeyTransactions constructs a key for
// querying all transactions in a chain.
// Prefixes: tag_chain + chain ID + tag_transaction
func MakeQueryKeyTransactions(chainID []byte) []byte {
	return elldb.MakePrefix(
		TagChain,
		chainID,
		TagTransaction,
	)
}

// MakeTreeKey constructs a key for
// recording state objects in a tree.
// Combination: block number (big endian) + object type
func MakeTreeKey(blockNumber uint64, objectType []byte) []byte {
	return append(util.EncodeNumber(blockNumber), objectType...)
}

// MakeKeyReOrg constructs a key for storing
// reorganization info.
// Prefixes: tag_reOrg + timestamp (big endian)
func MakeKeyReOrg(timestamp int64) []byte {
	return elldb.MakeKey(util.EncodeNumber(uint64(timestamp)),
		TagReOrg,
	)
}

// MakeQueryKeyReOrg constructs a key for
// querying reorganization objects.
// Prefixes: tag_reOrg
func MakeQueryKeyReOrg() []byte {
	return elldb.MakePrefix(
		TagReOrg,
	)
}
