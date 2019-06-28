package rpc

import (
	"fmt"

	"github.com/ellcrys/mother/rpc/jsonrpc"
)

// auth creates a session token when username and
// password match the configured rpc username and password.
// The returned token can be used to access private
// RPC APIs
func (s *Server) auth(username, password string) (string, error) {

	// Compare username and password with RPC
	if username != s.cfg.RPC.Username || password != s.cfg.RPC.Password {
		return "", fmt.Errorf("username or password are invalid")
	}

	// Create JWT token 
	tokenStr := jsonrpc.MakeSessionToken(username, s.cfg.RPC.SessionSecretKey, s.cfg.RPC.SessionTTL)

	return tokenStr, nil
}
