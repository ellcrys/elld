package miner

import (
	"github.com/ellcrys/elld/rpc"
	"github.com/ellcrys/elld/rpc/jsonrpc"
	"github.com/ellcrys/elld/types"
)

func (m *Miner) apiSetThreads(args interface{}) *jsonrpc.Response {

	num, ok := args.(float64)
	if !ok {
		return jsonrpc.Error(types.ErrCodeUnexpectedArgType,
			rpc.ErrMethodArgType("Integer").Error(), nil)
	}

	m.blakimoto.SetThreads(int(num))

	return jsonrpc.Success(true)
}

// APIs returns all API handlers
func (m *Miner) APIs() jsonrpc.APISet {
	return map[string]jsonrpc.APIInfo{

		// namespace: "miner"
		"start": {
			Namespace:   types.NamespaceMiner,
			Description: "Start the CPU miner",
			Private:     true,
			Func: func(params interface{}) *jsonrpc.Response {
				go m.Mine()
				return jsonrpc.Success(nil)
			},
		},
		"stop": {
			Namespace:   types.NamespaceMiner,
			Description: "Stop the CPU miner",
			Private:     true,
			Func: func(params interface{}) *jsonrpc.Response {
				m.Stop()
				return jsonrpc.Success(nil)
			},
		},
		"isMining": {
			Namespace:   types.NamespaceMiner,
			Description: "Check miner status",
			Func: func(params interface{}) *jsonrpc.Response {
				return jsonrpc.Success(m.IsMining())
			},
		},
		"getHashrate": {
			Namespace:   types.NamespaceMiner,
			Description: "Get current hashrate",
			Func: func(params interface{}) *jsonrpc.Response {
				return jsonrpc.Success(m.blakimoto.Hashrate())
			},
		},
		"numThreads": {
			Namespace:   types.NamespaceMiner,
			Description: "Get the number of miner threads",
			Func: func(params interface{}) *jsonrpc.Response {
				return jsonrpc.Success(m.blakimoto.Threads())
			},
		},
		"setThreads": {
			Namespace:   types.NamespaceMiner,
			Description: "Set the number of miner threads",
			Func:        m.apiSetThreads,
		},
	}
}
