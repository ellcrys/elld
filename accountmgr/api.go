package accountmgr

import "github.com/ellcrys/elld/types"

// APIs returns all API handlers
func (am *AccountManager) APIs() types.APISet {
	return map[string]types.APIFunc{
		"ListAccounts": func(args ...interface{}) (interface{}, error) {
			return am.ListAccounts()
		},
	}
}
