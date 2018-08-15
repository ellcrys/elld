package blockchain

import (
	"github.com/ellcrys/elld/blockchain/common"
	"github.com/ellcrys/elld/elldb"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/util"
	"github.com/imdario/mergo"
)

// GetMeta returns the metadata of the blockchain
func (b *Blockchain) GetMeta() *core.BlockchainMeta {
	var meta core.BlockchainMeta

	result := b.db.GetByPrefix(common.MakeBlockchainMetadataKey())
	if len(result) == 0 {
		return &meta
	}

	result[0].Scan(&meta)
	return &meta
}

// updateMeta updates the metadata of this chain
func (b *Blockchain) updateMeta(upd *core.BlockchainMeta) error {
	existingMeta := b.GetMeta()
	mergo.Merge(existingMeta, upd)
	return b.db.Put([]*elldb.KVObject{
		elldb.NewKVObject(common.MakeBlockchainMetadataKey(), util.ObjectToBytes(existingMeta)),
	})
}

// updateMetaWithTx is like updateMeta except it accepts a transaction
func (b *Blockchain) updateMetaWithTx(tx elldb.Tx, upd core.BlockchainMeta) error {
	existingMeta := b.GetMeta()
	mergo.Merge(existingMeta, upd)
	return tx.Put([]*elldb.KVObject{
		elldb.NewKVObject(common.MakeBlockchainMetadataKey(), util.ObjectToBytes(existingMeta)),
	})
}
