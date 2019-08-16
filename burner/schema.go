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

package burner

import (
	"fmt"

	"github.com/ellcrys/elld/util"

	"github.com/ellcrys/elld/elldb"
)

var (
	// TagLastScannedBlock is the tag for the key that stores the last scanned block
	// TODO: shorten this value
	TagLastScannedBlock = []byte("burner_lsb")

	// TagAddressUTXO is the tag for the key that stores burner address UTXOs'
	// TODO: shorten this value
	TagAddressUTXO = []byte("burner_addr_utxo")

	// TagIndexerBlock is the tag for the key that stores the last watched burner block.
	TagIndexerBlock = []byte("wb")
)

// MakeKeyIndexerBlock constructs a key for storing a blocks found by the block indexer.
func MakeKeyIndexerBlock(blockNumber int64) []byte {
	return elldb.MakeKey(
		util.EncodeNumber(uint64(blockNumber)),
		TagIndexerBlock,
	)
}

// MakeQueryKeyIndexerBlock constructs a key for querying blocks found by the block indexer.
func MakeQueryKeyIndexerBlock() []byte {
	return elldb.MakePrefix(
		TagIndexerBlock,
	)
}

// MakeKeyLastScannedBlock returns the key for storing/fetching the last scanned block
func MakeKeyLastScannedBlock(address string) []byte {
	return elldb.MakePrefix(
		TagLastScannedBlock,
		[]byte(address),
	)
}

// MakeQueryKeyUTXO returns the key for fetching all utxos
func MakeQueryKeyUTXO() []byte {
	return elldb.MakePrefix(
		TagAddressUTXO,
	)
}

// MakeQueryKeyAddressUTXO returns the key for fetching a specific utxo
// belonging to an address and transaction.
func MakeQueryKeyAddressUTXO(address, txID string, index uint32) []byte {
	return elldb.MakePrefix(
		TagAddressUTXO,
		[]byte(address),
		[]byte(txID),
		[]byte(fmt.Sprintf("%d", index)),
	)
}

// MakeQueryKeyAddressUTXOs returns the key for fetching all UTXOs
// belonging to an address.
func MakeQueryKeyAddressUTXOs(address string) []byte {
	return elldb.MakePrefix(
		TagAddressUTXO,
		[]byte(address),
	)
}

// MakeKeyAddressUTXO returns the key for storing/fetching a specific utxo
// belonging to a network, address and transaction.
func MakeKeyAddressUTXO(blockHeight int64, address, txID string, index uint32) []byte {
	return elldb.MakeKey(
		util.EncodeNumber(uint64(blockHeight)),
		TagAddressUTXO,
		[]byte(address),
		[]byte(txID),
		[]byte(fmt.Sprintf("%d", index)),
	)
}

// DocUTXO describes a UTXO
type DocUTXO struct {
	TxID  string
	Index uint32
	Value float64
}
