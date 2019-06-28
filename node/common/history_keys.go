package common

import (
	"github.com/ellcrys/mother/types"
	"github.com/ellcrys/mother/types/core"
	"github.com/ellcrys/mother/util"
)

// KeyBlock returns a key for storing the hash
// of a block received from a peer
func KeyBlock(hash string, peer core.Engine) []interface{} {
	return []interface{}{"b", hash, peer.StringID()}
}

// KeyBlock2 is like KeyBlock except it accepts a peer id
func KeyBlock2(hash string, peerID string) []interface{} {
	return []interface{}{"b", hash, peerID}
}

// KeyOrphanBlock returns a key for storing the
// hash of an orphan block received from a peer
func KeyOrphanBlock(blockHash util.Hash,
	peer core.Engine) []interface{} {
	return []interface{}{"ob", blockHash.HexStr(), peer.StringID()}
}

// KeyTx returns a key for storing the hash of a transaction
// received from a peer
func KeyTx(tx types.Transaction, peer core.Engine) []interface{} {
	return []interface{}{tx.GetID(), peer.StringID()}
}
