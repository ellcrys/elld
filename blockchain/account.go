package blockchain

import (
	"github.com/ellcrys/elld/types/core"
)

// putAccount adds an account to the store
func (b *Blockchain) putAccount(blockNo uint64, chain *Chain, account core.Account) error {
	b.chainLock.Lock()
	defer b.chainLock.Unlock()
	return chain.CreateAccount(blockNo, account)
}
