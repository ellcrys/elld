package accountmgr

import (
	"github.com/ellcrys/mother/rpc/jsonrpc"
	"github.com/ellcrys/mother/types"
)

func (am *AccountManager) apiListAccounts(interface{}) *jsonrpc.Response {
	accounts, err := am.ListAccounts()
	if err != nil {
		return jsonrpc.Error(types.ErrCodeListAccountFailed, err.Error(), nil)
	}

	var addresses []string
	for _, acct := range accounts {
		addresses = append(addresses, acct.Address)
	}

	return jsonrpc.Success(addresses)
}

// APIs returns all API handlers
func (am *AccountManager) APIs() jsonrpc.APISet {
	return map[string]jsonrpc.APIInfo{
		// namespace: "personal"
		"listAccounts": {
			Namespace:   types.NamespacePersonal,
			Description: "List all accounts",
			Func:        am.apiListAccounts,
		},
	}
}
