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
		bChain: b,
	}
}

// WorldReader provides functionalities for
// reading the the main and side chains. Using
// a ChainTraverser, it can search for and object
// from a chain and the chain's grand-parents.
type WorldReader struct {
	bChain *Blockchain
}

// GetAccount gets an account by the given address.
// If chain is nil, it will check in the main chain.
// If chain is provided, it will check within the chain
// or traverse the chain's parent and ancestors to find
// the account.
func (r *WorldReader) GetAccount(chain types.Chainer, address util.String,
	opts ...types.CallOp) (types.Account, error) {

	var result types.Account
	var err error

	// If a start chain is not given, we will use the main chain
	if chain == nil {
		r.bChain.chl.RLock()
		mainChain := r.bChain.bestChain
		r.bChain.chl.RUnlock()

		// If no best chain yet, we return an error
		if mainChain == nil {
			return nil, core.ErrBestChainUnknown
		}

		result, err = mainChain.GetAccount(address, opts...)
		if err != nil {
			return nil, err
		}

		if result == nil {
			return nil, core.ErrAccountNotFound
		}

		return result, nil
	}

	// maxChainHeight is the maximum block number
	// in the current chain that we are
	// allowed to access/consider. Any object associated
	// with block number greater than maxChainHeight is
	// not a valid history from the point of miner who created
	// the block.
	maxChainHeight := uint64(0)

	// Transverse the chain and its ancestors.
	err = r.bChain.NewChainTraverser().Start(chain).Query(func(ch types.Chainer) (bool, error) {
		// make a copy of the call options
		optsCopy := append([]types.CallOp{}, opts...)

		// If a max chain height is set, add a QueryBlockRange.Max value
		// which will prevent searching beyond the max height on the chain.
		if maxChainHeight > 0 {
			optsCopy = append(optsCopy, &common.OpBlockQueryRange{Max: maxChainHeight})
		}

		// Attempt to find the account on this chain.
		// If not found, return false and error if an error occurs
		result, err = ch.GetAccount(address, optsCopy...)
		if err != nil {
			if err != core.ErrAccountNotFound {
				return false, err
			}
		}

		// At this point, the current chain does not have
		// the address, if this chain has a parent chain,
		// it will be the next to be searched;
		// We need to determine the maxChainHeight
		// so we know what blocks were known to the
		// miner at the time they created the block.
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
