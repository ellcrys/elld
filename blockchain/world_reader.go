package blockchain

import (
	"fmt"

	"github.com/ellcrys/elld/blockchain/common"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/util"
)

const (
	ReaderUser  = 0x1
	ReaderMiner = 0x2
)

// NewWorldReader creates a new WorldReader
func (b *Blockchain) NewWorldReader() *WorldReader {
	return &WorldReader{
		bchain: b,
	}
}

// WorldReader provides functionalities for
// reading the data store in the main and
// side chains. It is able to see the state
// of the blockchain from a user and a miner
// perspective.
//
// A user typically wants to read only object
// in the best/main chain.
//
// For miners, when a client is required to
// process a block, it must view the block
// from the perspective of the block miner
// from the block 0 to the block's parent height.
// Some nodes might have committed objects and
// extended the chain and naively reading the chain
// without the ability to capture snapshot of chain
// history will result in inconsistent state root etc
type WorldReader struct {
	bchain *Blockchain
	reader int
}

// GetAccount gets an account by the given address in the
// chain provided. If chain is not provided, the mode
// is set to user mode, otherwise it is set to ReaderMiner.
//
// In user mode, the best chain is searched.
//
// In miner mode, the chain and its ancestors
// are traversed in such a way that only blocks that were
// known in the ancestor at the time of the fork are checked.
func (r *WorldReader) GetAccount(chain core.Chainer, address util.String, opts ...core.CallOp) (core.Account, error) {
	r.bchain.chainLock.RLock()
	defer r.bchain.chainLock.RUnlock()

	var result core.Account
	var err error

	// if chain is nil, set to user mode
	if chain == nil {
		r.reader = ReaderUser
	} else {
		r.reader = ReaderMiner
	}

	if r.reader == ReaderMiner {
		goto minerMode
	} else {
		goto userMode
	}

userMode:

	// If no best chain yet, we return an error
	if r.bchain.bestChain == nil {
		return nil, fmt.Errorf("no best chain yet")
	}

	result, err = r.bchain.bestChain.GetAccount(address, opts...)
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, core.ErrAccountNotFound
	}

	return result, nil

minerMode:

	// maxChainHeight is the height in the maximum
	// block number in the current chain that we are
	// allowed to access/consider. Any object associated
	// with block number greater than maxChainHeight is
	// not a valid history of the miner whose block
	// created the chain.
	maxChainHeight := uint64(0)

	// Transverse the chain and its ancestors.
	err = r.bchain.NewChainTransverser().Start(chain).Query(func(ch core.Chainer) (bool, error) {

		// make a copy of the call options
		optsCopy := append([]core.CallOp{}, opts...)

		// Add a QueryBlockRange containing the max chain height
		// to the call options slice.
		if maxChainHeight > 0 {
			optsCopy = append(optsCopy, &common.QueryBlockRange{Max: maxChainHeight})
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
			if ch.GetInfo().ParentBlockNumber > 0 {
				maxChainHeight = ch.GetInfo().ParentBlockNumber
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
