package rpc

import (
	"github.com/ellcrys/elld/rpc/jsonrpc"
	"github.com/ellcrys/elld/types"
)

func (s *Server) rpcAuth(params jsonrpc.Params) *jsonrpc.Response {
	var m map[string]string
	if err := params.Scan(&m); err != nil {
		return jsonrpc.Error(types.ErrInvalidAuthParams, err.Error(), nil)
	}

	// perform authentication and create a session token
	token, err := s.auth(m["username"], m["password"])
	if err != nil {
		return jsonrpc.Error(types.ErrInvalidAuthCredentials, err.Error(), nil)
	}

	return jsonrpc.Success(token)
}

// APIs returns all API handlers
func (s *Server) APIs() jsonrpc.APISet {
	return map[string]jsonrpc.APIInfo{
		"auth": jsonrpc.APIInfo{
			Func: s.rpcAuth,
		},
	}
}
