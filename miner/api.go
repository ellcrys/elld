package miner

import (
	"github.com/ellcrys/elld/rpc/jsonrpc"
	"github.com/ellcrys/elld/types/core"
)

// APIs returns all API handlers
func (m *Miner) APIs() jsonrpc.APISet {
	return map[string]jsonrpc.APIInfo{

		// namespace: "miner"
		"start": {
			Namespace:   core.NamespaceMiner,
			Description: "Start the CPU miner",
			Private:     true,
			Func: func(params interface{}) *jsonrpc.Response {
				go m.Mine()
				return jsonrpc.Success(nil)
			},
		},
		"stop": {
			Namespace:   core.NamespaceMiner,
			Description: "Stop the CPU miner",
			Private:     true,
			Func: func(params interface{}) *jsonrpc.Response {
				m.Stop()
				return jsonrpc.Success(nil)
			},
		},
		"isMining": {
			Namespace:   core.NamespaceMiner,
			Description: "Check miner status",
			Func: func(params interface{}) *jsonrpc.Response {
				return jsonrpc.Success(m.IsMining())
			},
		},
		"getHashrate": {
			Namespace:   core.NamespaceMiner,
			Description: "Get current hashrate",
			Func: func(params interface{}) *jsonrpc.Response {
				return jsonrpc.Success(m.blakimoto.Hashrate())
			},
		},
	}
}
