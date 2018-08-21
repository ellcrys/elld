package node

import (
	"github.com/ellcrys/elld/rpc/jsonrpc"
	"github.com/ellcrys/elld/wire"
)

// APIs returns all API handlers
func (n *Node) APIs() jsonrpc.APISet {
	return map[string]jsonrpc.APIInfo{
		"TransactionAdd": jsonrpc.APIInfo{
			Private: true,
			Func: func(params jsonrpc.Params) *jsonrpc.Response {
				var tx wire.Transaction
				params.Scan(&tx)
				return jsonrpc.Success(n.addTransaction(&tx))
			},
		},
	}
}
