package miner

import "github.com/ellcrys/elld/rpc/jsonrpc"

// APIs returns all API handlers
func (m *Miner) APIs() jsonrpc.APISet {
	return map[string]jsonrpc.APIInfo{

		"minerStop": jsonrpc.APIInfo{
			Namespace:   "node",
			Description: "Start mining",
			Private:     true,
			Func: func(params interface{}) *jsonrpc.Response {
				m.Stop()
				return jsonrpc.Success(nil)
			},
		},

		"minerStart": jsonrpc.APIInfo{
			Namespace:   "node",
			Description: "Stop mining",
			Private:     true,
			Func: func(params interface{}) *jsonrpc.Response {
				go m.Mine()
				return jsonrpc.Success(nil)
			},
		},

		"mining": jsonrpc.APIInfo{
			Namespace:   "node",
			Description: "Check mining status",
			Func: func(params interface{}) *jsonrpc.Response {
				return jsonrpc.Success(m.IsMining())
			},
		},

		"minerHashrate": jsonrpc.APIInfo{
			Namespace:   "node",
			Description: "Get current hashrate",
			Func: func(params interface{}) *jsonrpc.Response {
				return jsonrpc.Success(m.blakimoto.Hashrate())
			},
		},
	}
}
