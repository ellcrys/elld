package rpc

import (
	"github.com/ellcrys/elld/accountmgr"
)

// AccountGetAllPayload is used to define payload for AccountsGet
type AccountGetAllPayload map[string]interface{}

// AccountGetAll returns all accounts known to the node
func (s *Service) AccountGetAll(payload AccountGetAllPayload, result *Result) error {

	var errCh = make(chan error, 1)
	var accts = make(chan []*accountmgr.StoredAccount, 1)
	err := s.logic.AccountsGet(accts, errCh)
	if err != nil {
		return NewErrorResult(result, err.Error(), errCodeAccountStoredAccount, 500)
	}

	resp := map[string]interface{}{
		"accounts": []string{},
	}

	for _, a := range <-accts {
		resp["accounts"] = append(resp["accounts"].([]string), a.Address)
	}

	return NewOKResult(result, 200, resp)
}
