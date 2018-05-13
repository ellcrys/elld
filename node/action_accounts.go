package node

import (
	"path"

	"github.com/ellcrys/druid/accountmgr"
	"github.com/ellcrys/druid/configdir"
)

// GetAccounts returns all accounts on this node
func (n *Node) GetAccounts() ([]*accountmgr.StoredAccount, error) {

	am := accountmgr.New(path.Join(n.cfg.ConfigDir(), configdir.AccountDirName))
	accounts, err := am.GetAccountsOnDisk()
	if err != nil {
		return nil, err
	}

	return accounts, nil
}