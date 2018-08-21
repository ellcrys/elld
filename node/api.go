package node

import (
	"github.com/ellcrys/elld/rpc/jsonrpc"
)

// APIs returns all API handlers
func (n *Node) APIs() jsonrpc.APISet {
	return map[string]jsonrpc.APIInfo{
		"TransactionAdd": jsonrpc.APIInfo{
			Private: true,
			Func: func(params jsonrpc.Params) jsonrpc.Response {
				// return jsonrpc.Success(n.addTransaction(args[0].(*wire.Transaction)))
				return jsonrpc.Success(nil)
			},
		},
	}
}
