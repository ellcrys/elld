package rpc

// GetAccountsArgs represents arguments for method AccountsGet
type GetAccountsArgs struct {
}

// GetAccountsPayload is used to define payload for AccountsGet
type GetAccountsPayload struct {
	Args GetAccountsArgs `json:"args"`
	Sig  []byte          `json:"sig"`
}

// GetAccounts returns accounts known to the node
func (s *Service) GetAccounts(payload GetAccountsPayload, result *Result) error {

	storedAccounts, err := s.node.GetAccounts()
	if err != nil {
		return NewErrorResult(result, err.Error(), errCodeAccountStoredAccount, 500)
	}

	resp := map[string]interface{}{
		"accounts": []string{},
	}

	for _, a := range storedAccounts {
		resp["accounts"] = append(resp["accounts"].([]string), a.Address)
	}

	return NewOKResult(result, 200, resp)
}
