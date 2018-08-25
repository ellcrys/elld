package store

import (
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/util"
)

// ChainRead provides read-only access to
// objects belonging to a single chain.
type ChainRead struct {
	ch core.Chainer
}

// NewChainReader creates a ChainReader object
func NewChainReader(ch core.Chainer) *ChainRead {
	return &ChainRead{
		ch: ch,
	}
}

// GetID gets the chain ID
func (r *ChainRead) GetID() util.String {
	return r.GetID()
}

// GetParent returns a chain reader to the parent.
// Returns nil if chain has no parent.
func (r *ChainRead) GetParent() core.ChainReader {
	if ch := r.ch.GetParent(); ch != nil {
		return ch.ChainReader()
	}
	return nil
}

// GetBlock finds and returns a block associated with chainID.
// When 0 is passed, it should return the block with the highest number
func (r *ChainRead) GetBlock(number uint64, opts ...core.CallOp) (core.Block, error) {
	return r.ch.GetBlock(number, opts...)
}

// GetBlockByHash finds and returns a block associated with chainID.
func (r *ChainRead) GetBlockByHash(hash util.Hash, opts ...core.CallOp) (core.Block, error) {
	return r.ch.GetStore().GetBlockByHash(hash, opts...)
}

// GetHeader gets the header of a block.
// When 0 is passed, it should return the header of the block with the highest number
func (r *ChainRead) GetHeader(number uint64, opts ...core.CallOp) (core.Header, error) {
	return r.ch.GetStore().GetHeader(number, opts...)
}

// GetHeaderByHash finds and returns the header of a block matching hash
func (r *ChainRead) GetHeaderByHash(hash util.Hash, opts ...core.CallOp) (core.Header, error) {
	return r.ch.GetStore().GetHeaderByHash(hash, opts...)
}

// Current gets the current block at the tip of the chain
func (r *ChainRead) Current(opts ...core.CallOp) (core.Block, error) {
	return r.ch.GetStore().Current(opts...)
}
