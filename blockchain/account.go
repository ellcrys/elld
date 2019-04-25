package blockchain

import (
	"sort"

	"github.com/ellcrys/elld/blockchain/common"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/util"
)

// CreateAccount creates an account
// that is associated with the given block number
// and chain.
func (b *Blockchain) CreateAccount(blockNo uint64, chain types.Chainer,
	account types.Account) error {
	return chain.CreateAccount(blockNo, account)
}

// GetAccountNonce gets the nonce of an account
func (b *Blockchain) GetAccountNonce(address util.String,
	opts ...types.CallOp) (uint64, error) {
	account, err := b.GetAccount(address, opts...)
	if err != nil {
		return 0, err
	}
	return account.GetNonce(), nil
}

// GetAccount gets an account by its address
func (b *Blockchain) GetAccount(address util.String,
	opts ...types.CallOp) (types.Account, error) {
	opt := common.GetChainerOp(opts...)
	account, err := b.NewWorldReader().GetAccount(opt.Chain, address, opts...)
	if err != nil {
		return nil, err
	}
	return account, nil
}

// ListAccounts list all accounts
func (b *Blockchain) ListAccounts(opts ...types.CallOp) ([]types.Account, error) {
	b.chl.RLock()
	defer b.chl.RUnlock()

	if b.bestChain == nil {
		return nil, core.ErrBestChainUnknown
	}

	return b.bestChain.GetAccounts(opts...)
}

// ListTopAccounts lists top n accounts
func (b *Blockchain) ListTopAccounts(n int, opts ...types.CallOp) ([]types.Account, error) {
	accounts, err := b.ListAccounts(opts...)
	if err != nil {
		return nil, err
	}

	sort.Slice(accounts, func(i, j int) bool {
		aI := accounts[i].GetBalance().Decimal()
		aJ := accounts[j].GetBalance().Decimal()
		return aI.GreaterThan(aJ)
	})

	if nAccts := len(accounts); nAccts < n {
		n = nAccts
	}

	return accounts[:n], nil
}
