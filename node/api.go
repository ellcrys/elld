package node

import (
	"github.com/ellcrys/elld/rpc"
	"github.com/ellcrys/elld/rpc/jsonrpc"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/wire"
	"github.com/ncodes/mapstructure"
)

// APIs returns all API handlers
func (n *Node) APIs() jsonrpc.APISet {
	return map[string]jsonrpc.APIInfo{
		"TransactionAdd": jsonrpc.APIInfo{
			Private: true,
			Func: func(params interface{}) *jsonrpc.Response {

				p, ok := params.(map[string]interface{})
				if !ok {
					return jsonrpc.Error(types.ErrCodeUnexpectedArgType, rpc.ErrMethodArgType("JSON").Error(), nil)
				}

				var tx wire.Transaction
				mapstructure.Decode(p, &tx)

				return jsonrpc.Success(n.addTransaction(&tx))
			},
		},
	}
}
