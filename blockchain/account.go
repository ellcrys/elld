package blockchain

import (
	"github.com/ellcrys/elld/blockchain/common"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/wire"
)

// putAccount adds an account to the store
func (b *Blockchain) putAccount(blockNo uint64, chain *Chain, account *wire.Account) error {
	b.chainLock.Lock()
	defer b.chainLock.Unlock()
	return chain.CreateAccount(blockNo, account)
}

// getAccount fetches an account on the provided chain
// with the matching address. The most recent version of
// the account is returned.
//
// If the account is not found in the chain, the parent chain
// and its parent is checked up to the root chain.
func (b *Blockchain) getAccount(chain common.Chainer, address util.String, opts ...common.CallOp) (*wire.Account, error) {
	b.chainLock.RLock()
	defer b.chainLock.RUnlock()

	var account *wire.Account
	var err error

	for address != "" {
		// Initiate a search that will traverse from the chain,
		// its parents up to the root parent checking each chain
		// for the existence of the account
		if account, err = chain.GetAccount(address, opts...); err != nil {
			if err != common.ErrAccountNotFound {
				return nil, err
			}
		}

		// If not found, check whether this chain has a parent chain
		// and if so, set the parent chain as the next to be checked and
		// also fetch the chain from the chain cache.
		if account == nil && chain.GetParentInfo().ParentChainID != "" {
			if chain = b.chains[chain.GetParentInfo().ParentChainID]; chain == nil {
				break
			}
			continue
		}
		break
	}

	if account == nil {
		return nil, common.ErrAccountNotFound
	}

	return account, nil
}
