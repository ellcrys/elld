package rpc

import "github.com/ellcrys/elld/accountmgr"

// AccountGetAllPayload is used to define payload for AccountsGet
type AccountGetAllPayload map[string]interface{}

// AccountGetAll returns all accounts known to account manager
func (s *Service) AccountGetAll(payload AccountGetAllPayload, result *Result) error {

	apiFunc := s.accountMgr.MustGet("ListAccounts")
	apiResult, err := apiFunc()
	if err != nil {
		return NewErrorResult(result, err.Error(), errCodeAccountStoredAccount, 500)
	}

	resp := map[string]interface{}{
		"accounts": []string{},
	}

	for _, a := range apiResult.([]*accountmgr.StoredAccount) {
		resp["accounts"] = append(resp["accounts"].([]string), a.Address)
	}

	return NewOKResult(result, 200, resp)
}
