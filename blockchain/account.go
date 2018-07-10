package blockchain

import (
	"encoding/json"
	"fmt"

	"github.com/ellcrys/elld/database"
	"github.com/ellcrys/elld/wire"
)

// ErrAccountNotFound refers to a missing account
var ErrAccountNotFound = fmt.Errorf("account not found")

// makeAccountKey constructs a key for persisting account data in
// the store and hash tree.
func makeAccountKey(chainID, address string) []byte {
	return database.MakePrefix([]string{chainID, "account", address})
}

// makeAccountKey constructs a key for persisting account data in
// the store and hash tree.
func queryAccountKey(chainID, address string) []byte {
	return database.MakePrefix([]string{chainID, "account", address})
}

// updateOrCreateAccount will update an account or create it if it does not
// currently exists on the main chain. It uses the Address field of the account
// for querying or for construction of the account key during creation.
// func (b *Blockchain) updateOrCreateAccount(account *wire.Account) error {

// 	// validate the account.
// 	if err := wire.ValidateAccount(account); err != nil {
// 		return err
// 	}

// 	// find the account by its address
// 	account, err := b.GetAccount(account.Address)
// 	if err != nil {
// 		if err != ErrAccountNotFound {
// 			return err
// 		}
// 	}

// 	b.lock.Lock()
// 	defer b.lock.Unlock()

// 	// get the header of the current block of the best chain.
// 	// We need the block number to serve as a version number for the key.
// 	curHeader, err := b.bestChain.getCurrentBlockHeader()
// 	if err != nil {
// 		return err
// 	}

// 	// At this point the account does not exist, so we need to create it.
// 	if account == nil {
// 		key := makeAccountKey(b.bestChain.id, account.Address, curHeader.Number)
// 		_ = key
// 	}

// 	return nil
// }

// GetAccount fetches an account on the main chain with the matching address.
func (b *Blockchain) GetAccount(address string) (*wire.Account, error) {
	b.lock.RLock()
	defer b.lock.RUnlock()

	// make an account key, then query the database for an account
	// matching the key
	var result []*database.KVObject
	if err := b.store.Get(makeAccountKey(b.bestChain.id, address), &result); err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, ErrAccountNotFound
	}

	object := result[0]
	var account wire.Account

	return &account, json.Unmarshal(object.Value, &account)
}
