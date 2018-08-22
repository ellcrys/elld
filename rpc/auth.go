package rpc

import (
	"fmt"

	"github.com/ellcrys/elld/rpc/jsonrpc"
)

// auth creates a session token when username and
// password match the configured rpc username and password.
// The returned token can be used to access private
// RPC APIs
func (s *Server) auth(username, password string) (string, error) {

	// compare username and password with RPC
	if username != s.cfg.RPC.Username || password != s.cfg.RPC.Password {
		return "", fmt.Errorf("username or password are invalid")
	}

	// create JWT token that will expire in 1 hour.
	tokenStr := jsonrpc.MakeSessionToken(username, s.cfg.RPC.SessionSecretKey)

	return tokenStr, nil
}
