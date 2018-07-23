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
	key := common.MakeAccountKey(blockNo, chain.id, account.Address)
	return chain.store.Put(key, util.StructToBytes(account))
}

// GetAccount fetches an account on the provided chain
// with the matching address. The most recent version of
// the account is returned.
func (b *Blockchain) GetAccount(chain *Chain, address string) (*wire.Account, error) {
	b.chainLock.RLock()
	defer b.chainLock.RUnlock()

	// make an account key, then query the database for an account
	// matching the key
	var account wire.Account
	var key = common.QueryAccountKey(chain.id, address)
	var result database.KVObject
	chain.store.GetFirstOrLast(false, key, &result)
	if len(result.Key) == 0 {
		return nil, common.ErrAccountNotFound
	}

	return &account, json.Unmarshal(result.Value, &account)
}
