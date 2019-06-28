package console

import (
	"github.com/ellcrys/mother/rpc/jsonrpc"
)

// APIs returns all API handlers
func (c *Console) APIs() jsonrpc.APISet {
	return map[string]jsonrpc.APIInfo{}
}
