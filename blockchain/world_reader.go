package blockchain

import (
	"github.com/ellcrys/elld/blockchain/common"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/util"
)

// NewWorldReader creates a new WorldReader
func (b *Blockchain) NewWorldReader() *WorldReader {
	return &WorldReader{
		bchain: b,
	}
}

// WorldReader provides functionalities for
// reading the the main and side chains. Using
// a ChainTraverser, it can search for and object
// from a chain and the chain's grand-parents.
type WorldReader struct {
	bchain *Blockchain
}

// GetAccount gets an account by the given
// address in the chain provided.
func (r *WorldReader) GetAccount(chain types.Chainer, address util.String,
	opts ...types.CallOp) (types.Account, error) {
	r.bchain.chainLock.RLock()
	defer r.bchain.chainLock.RUnlock()

	var result types.Account
	var err error

	// If a start chain is not given,
	// We will use whatever the main chain
	if chain == nil {
		// If no best chain yet, we return an error
		if r.bchain.bestChain == nil {
			return nil, core.ErrBestChainUnknown
		}

		result, err = r.bchain.bestChain.GetAccount(address, opts...)
		if err != nil {
			return nil, err
		}

		if result == nil {
			return nil, core.ErrAccountNotFound
		}

		return result, nil
	}

	// maxChainHeight is the height in the maximum
	// block number in the current chain that we are
	// allowed to access/consider. Any object associated
	// with block number greater than maxChainHeight is
	// not a valid history from the point of miner who created
	// the block.
	maxChainHeight := uint64(0)

	// Transverse the chain and its ancestors.
	err = r.bchain.NewChainTraverser().Start(chain).Query(func(ch types.Chainer) (bool, error) {
		// make a copy of the call options
		optsCopy := append([]types.CallOp{}, opts...)

		// Add a QueryBlockRange containing the max chain height
		// to the call options slice.
		if maxChainHeight > 0 {
			optsCopy = append(optsCopy, &common.OpBlockQueryRange{Max: maxChainHeight})
		}

		result, err = ch.GetAccount(address, optsCopy...)
		if err != nil {
			if err != core.ErrAccountNotFound {
				return false, err
			}
		}

		// At this point, the current chain does not have
		// the address, if this chain has a parent chain,
		// the next iteration will pass the parent to
		// this block. We need to determine the maxChainHeight
		// so we know what blocks were know to the miner at
		// the time there created their block.
		if result == nil {
			chInfo := ch.GetInfo()
			if chInfo.GetParentBlockNumber() > 0 {
				maxChainHeight = chInfo.GetParentBlockNumber()
			}
			return false, nil
		}

		return true, nil
	})

	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, core.ErrAccountNotFound
	}

	return result, nil
}
