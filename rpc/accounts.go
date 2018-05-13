package rpc

// AccountsGetArgs represents arguments for method AccountsGet
type AccountsGetArgs struct {
}

// AccountsGetPayload is used to define method arguments
type AccountsGetPayload struct {
	Args AccountsGetArgs `json:"args"`
	Sig  []byte          `json:"sig"`
}

// GetAccounts returns accounts known to the node
func (s *Service) GetAccounts(payload AccountsGetPayload, result *Result) error {

	storedAccounts, err := s.node.GetAccounts()
	if err != nil {
		return NewErrorResult(result, err.Error(), accountErrStoredAccount, 500)
	}

	resp := map[string]interface{}{
		"accounts": []string{},
	}

	for _, a := range storedAccounts {
		resp["accounts"] = append(resp["accounts"].([]string), a.Address)
	}

	return NewOKResult(result, 200, resp)
}
