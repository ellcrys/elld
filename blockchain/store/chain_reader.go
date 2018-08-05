package store

import (
	"github.com/ellcrys/elld/blockchain/common"
	"github.com/ellcrys/elld/wire"
)

// ChainReader provides read-only access to
// objects belonging to a single chain.
type ChainReader struct {
	chainID string
	store   Storer
}

// NewChainReader creates a ChainReader object
func NewChainReader(store Storer, chainID string) *ChainReader {
	return &ChainReader{
		chainID: chainID,
		store:   store,
	}
}

// GetBlock finds and returns a block associated with chainID.
// When 0 is passed, it should return the block with the highest number
func (r *ChainReader) GetBlock(number uint64, opts ...common.CallOp) (*wire.Block, error) {
	return r.store.GetBlock(r.chainID, number, opts...)
}

// GetBlockByHash finds and returns a block associated with chainID.
func (r *ChainReader) GetBlockByHash(hash string, opts ...common.CallOp) (*wire.Block, error) {
	return r.store.GetBlockByHash(r.chainID, hash, opts...)
}

// GetBlockHeader gets the header of a block.
// When 0 is passed, it should return the header of the block with the highest number
func (r *ChainReader) GetBlockHeader(number uint64, opts ...common.CallOp) (*wire.Header, error) {
	return r.store.GetBlockHeader(r.chainID, number, opts...)
}

// GetBlockHeaderByHash finds and returns the header of a block matching hash
func (r *ChainReader) GetBlockHeaderByHash(hash string) (*wire.Header, error) {
	return r.store.GetBlockHeaderByHash(r.chainID, hash)
}
