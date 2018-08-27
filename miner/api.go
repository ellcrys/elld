package miner

import "github.com/ellcrys/elld/rpc/jsonrpc"

// APIs returns all API handlers
func (m *Miner) APIs() jsonrpc.APISet {
	return map[string]jsonrpc.APIInfo{

		"minerStop": jsonrpc.APIInfo{
			Private: true,
			Func: func(params interface{}) *jsonrpc.Response {
				m.Stop()
				return jsonrpc.Success(nil)
			},
		},

		"minerStart": jsonrpc.APIInfo{
			Private: true,
			Func: func(params interface{}) *jsonrpc.Response {
				go m.Mine()
				return jsonrpc.Success(nil)
			},
		},

		"mining": jsonrpc.APIInfo{
			Func: func(params interface{}) *jsonrpc.Response {
				return jsonrpc.Success(m.IsMining())
			},
		},

		"minerHashrate": jsonrpc.APIInfo{
			Func: func(params interface{}) *jsonrpc.Response {
				return jsonrpc.Success(m.blakimoto.Hashrate())
			},
		},
	}
}
