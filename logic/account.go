package logic

import (
	path "path/filepath"

	"github.com/ellcrys/elld/accountmgr"
	"github.com/ellcrys/elld/configdir"
)

// AccountsGet returns all accounts on this node.
// The results are sent to result channel and error is sent to errCh
func (l *Logic) AccountsGet(result chan []*accountmgr.StoredAccount, errCh chan error) error {

	// Using the account manager, fetch all accounts in the config directory
	am := accountmgr.New(path.Join(l.engine.Cfg().ConfigDir(), configdir.AccountDirName))
	accounts, err := am.GetAccountsOnDisk()
	if err != nil {
		return sendErr(errCh, err)
	}

	result <- accounts

	return sendErr(errCh, nil)
}
