package rpc

import (
	"github.com/ellcrys/elld/rpc/jsonrpc"
	"github.com/ellcrys/elld/types"
)

func (s *Server) apiRPCAuth(params interface{}) *jsonrpc.Response {

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

func (s *Server) apiRPCStop(params interface{}) *jsonrpc.Response {
	s.Stop()
	return jsonrpc.Success("stopped")
}

func (s *Server) apiRPCEcho(params interface{}) *jsonrpc.Response {
	return jsonrpc.Success(params)
}

// APIs returns all API handlers
func (s *Server) APIs() jsonrpc.APISet {
	return map[string]jsonrpc.APIInfo{

		// namespace: "admin"
		"auth": {
			Namespace:   types.NamespaceAdmin,
			Description: "Get a session token",
			Func:        s.apiRPCAuth,
		},

		// namespace: "rpc"
		"stop": {
			Private:     true,
			Namespace:   types.NamespaceRPC,
			Description: "Get a session token",
			Func:        s.apiRPCStop,
		},
		"echo": {
			Namespace:   types.NamespaceRPC,
			Description: "Sends back the parameter passed to it",
			Func:        s.apiRPCEcho,
		},
	}
}
