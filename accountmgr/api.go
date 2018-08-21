package accountmgr

import (
	"github.com/ellcrys/elld/rpc/jsonrpc"
	"github.com/ellcrys/elld/types"
)

// APIs returns all API handlers
func (am *AccountManager) APIs() jsonrpc.APISet {
	return map[string]jsonrpc.APIInfo{

		"ListAccounts": jsonrpc.APIInfo{
			Func: func(params jsonrpc.Params) jsonrpc.Response {
				accounts, err := am.ListAccounts()
				if err != nil {
					return jsonrpc.Error(types.ErrCodeListAccountFailed, err.Error(), nil)
				}
				return jsonrpc.Success(accounts)
			},
		},
	}
}
