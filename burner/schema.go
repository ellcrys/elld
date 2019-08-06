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
	TagLastScannedBlock = []byte("burner_lsb")

	// TagAddressUTXO is the tag for the key that stores burner address UTXOs
	TagAddressUTXO = []byte("burner_addr_utxo")
)

// MakeKeyLastScannedBlock returns the key for storing/fetching the last scanned block
func MakeKeyLastScannedBlock(netID, address string) []byte {
	return elldb.MakePrefix(
		TagLastScannedBlock,
		[]byte(netID),
		[]byte(address),
	)
}

// MakeQueryKeyAddressUTXO returns the key for fetching a specific utxo
// belonging to a network, address and transaction.
func MakeQueryKeyAddressUTXO(netID, address, txHash string, index uint32) []byte {
	return elldb.MakePrefix(
		TagAddressUTXO,
		[]byte(netID),
		[]byte(address),
		[]byte(txHash),
		[]byte(fmt.Sprintf("%d", index)),
	)
}

// MakeQueryKeyAddressUTXOs returns the key for fetching all UTXOs
// belonging to a network and address.
func MakeQueryKeyAddressUTXOs(netID, address string) []byte {
	return elldb.MakePrefix(
		TagAddressUTXO,
		[]byte(netID),
		[]byte(address),
	)
}

// MakeKeyAddressUTXO returns the key for storing/fetching a specific utxo
// belonging to a network, address and transaction.
func MakeKeyAddressUTXO(blockHeight int64, netID, address, txHash string, index uint32) []byte {
	return elldb.MakeKey(
		util.EncodeNumber(uint64(blockHeight)),
		TagAddressUTXO,
		[]byte(netID),
		[]byte(address),
		[]byte(txHash),
		[]byte(fmt.Sprintf("%d", index)),
	)
}

// DocUTXO describes a UTXO
type DocUTXO struct {
	TxHash      string
	Index       uint32
	PkScriptStr string
	Value       float64
}
