package blockchain

import (
	"github.com/ellcrys/elld/blockchain/common"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/util"
)

// CreateAccount creates an account
// that is associated with the given block number
// and chain.
func (b *Blockchain) CreateAccount(blockNo uint64, chain core.Chainer, account core.Account) error {
	b.chainLock.Lock()
	defer b.chainLock.Unlock()
	return chain.CreateAccount(blockNo, account)
}

// GetAccountNonce gets the nonce of an account
func (b *Blockchain) GetAccountNonce(address util.String, opts ...core.CallOp) (uint64, error) {
	b.chainLock.RLock()
	defer b.chainLock.RUnlock()
	account, err := b.GetAccount(address, opts...)
	if err != nil {
		return 0, err
	}
	return account.GetNonce(), nil
}

// GetAccount gets an account by its address
func (b *Blockchain) GetAccount(address util.String, opts ...core.CallOp) (core.Account, error) {
	b.chainLock.RLock()
	defer b.chainLock.RUnlock()
	opt := common.GetChainerOp(opts...)
	account, err := b.NewWorldReader().GetAccount(opt.Chain, address, opts...)
	if err != nil {
		return nil, err
	}
	return account, nil
}
