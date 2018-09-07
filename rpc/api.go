package rpc

import (
	"github.com/ellcrys/elld/rpc/jsonrpc"
	"github.com/ellcrys/elld/types"
)

func (s *Server) rpcAuth(params interface{}) *jsonrpc.Response {

	p, ok := params.(map[string]interface{})
	if !ok {
		return jsonrpc.Error(types.ErrCodeUnexpectedArgType, ErrMethodArgType("JSON").Error(), nil)
	}

	username, _ := p["username"].(string)
	password, _ := p["password"].(string)

	// perform authentication and create a session token
	token, err := s.auth(username, password)
	if err != nil {
		return jsonrpc.Error(types.ErrCodeInvalidAuthCredentials, err.Error(), nil)
	}

	return jsonrpc.Success(token)
}

// APIs returns all API handlers
func (s *Server) APIs() jsonrpc.APISet {
	return map[string]jsonrpc.APIInfo{
		"auth": jsonrpc.APIInfo{
			Namespace:   "admin",
			Description: "Get a session token",
			Func:        s.rpcAuth,
		},
	}
}
