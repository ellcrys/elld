package blockchain

import (
	"encoding/json"

	"github.com/ellcrys/druid/util"
	common "github.com/ellcrys/elld/blockchain/common"
	"github.com/ellcrys/elld/database"
	"github.com/ellcrys/elld/wire"
)

// putAccount adds an account to the store
func (b *Blockchain) putAccount(blockNo uint64, chain *Chain, account *wire.Account) error {
	b.chainLock.Lock()
	defer b.chainLock.Unlock()

	// validate account

	key := common.MakeAccountKey(blockNo, chain.GetID(), account.Address)
	return chain.store.Put(key, util.StructToBytes(account))
}

// getAccount fetches an account on the provided chain
// with the matching address. The most recent version of
// the account is returned.
//
// If the account is not found in the chain, the parent chain
// and its parent is checked up to the root chain.
func (b *Blockchain) getAccount(chain *Chain, address string) (*wire.Account, error) {
	b.chainLock.RLock()
	defer b.chainLock.RUnlock()

	var result database.KVObject
	var account wire.Account
	var curChainID = chain.id

	for address != "" {
		// Initiate a look that will traverse from the chain,
		// its parents up to the root parent checking each chain
		// for the existence of the account
		var key = common.QueryAccountKey(curChainID, address)
		chain.store.GetFirstOrLast(false, key, &result)

		// If not found, check whether this chain has a parent chain
		// and if so, set the parent chain as the next to be checked and
		// also fetch the chain from the chain cache.
		if result.IsEmpty() && chain.info.ParentChainID != "" {
			curChainID = chain.info.ParentChainID
			if chain = b.chains[chain.info.ParentChainID]; chain == nil {
				break
			}
			continue
		} else if !result.IsEmpty() {
			// At this point, we found the account. Unmarshal and exit loop
			if err := json.Unmarshal(result.Value, &account); err != nil {
				return nil, err
			}
		}
		break
	}

	// If account is not set, return error
	if account == (wire.Account{}) {
		return nil, common.ErrAccountNotFound
	}

	return &account, nil
}
