package store

import (
	"github.com/ellcrys/elld/blockchain/common"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/wire"
)

// ChainRead provides read-only access to
// objects belonging to a single chain.
type ChainRead struct {
	chainID util.String
	store   common.ChainStorer
}

// NewChainReader creates a ChainReader object
func NewChainReader(store common.ChainStorer, chainID util.String) *ChainRead {
	return &ChainRead{
		chainID: chainID,
		store:   store,
	}
}

// GetBlock finds and returns a block associated with chainID.
// When 0 is passed, it should return the block with the highest number
func (r *ChainRead) GetBlock(number uint64, opts ...common.CallOp) (*wire.Block, error) {
	return r.store.GetBlock(number, opts...)
}

// GetBlockByHash finds and returns a block associated with chainID.
func (r *ChainRead) GetBlockByHash(hash util.Hash, opts ...common.CallOp) (*wire.Block, error) {
	return r.store.GetBlockByHash(hash, opts...)
}

// GetHeader gets the header of a block.
// When 0 is passed, it should return the header of the block with the highest number
func (r *ChainRead) GetHeader(number uint64, opts ...common.CallOp) (*wire.Header, error) {
	return r.store.GetHeader(number, opts...)
}

// GetHeaderByHash finds and returns the header of a block matching hash
func (r *ChainRead) GetHeaderByHash(hash util.Hash) (*wire.Header, error) {
	return r.store.GetHeaderByHash(hash)
}

// Current gets the current block at the tip of the chain
func (r *ChainRead) Current(opts ...common.CallOp) (*wire.Block, error) {
	return r.store.Current(opts...)
}
