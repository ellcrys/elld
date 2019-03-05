package miner

import (
	"github.com/ellcrys/elld/params"
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

	m.SetNumThreads(int(num))

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
			Func: func(arg interface{}) *jsonrpc.Response {

				// Do not start miner when miner key is ephemeral.
				// Block rewards will be lost if allowed.
				if m.minerKey.Meta["ephemeral"] != nil {
					return jsonrpc.Success(params.ErrMiningWithEphemeralKey.Error())
				}

				go m.Begin()
				return jsonrpc.Success(true)
			},
		},
		"stop": {
			Namespace:   types.NamespaceMiner,
			Description: "Stop the CPU miner",
			Private:     true,
			Func: func(params interface{}) *jsonrpc.Response {
				m.Stop()
				return jsonrpc.Success(true)
			},
		},
		"isMining": {
			Namespace:   types.NamespaceMiner,
			Description: "Check miner status",
			Func: func(params interface{}) *jsonrpc.Response {
				return jsonrpc.Success(m.isMining())
			},
		},
		"getHashrate": {
			Namespace:   types.NamespaceMiner,
			Description: "Get current hashrate",
			Func: func(params interface{}) *jsonrpc.Response {
				return jsonrpc.Success(m.getHashrate())
			},
		},
		"numThreads": {
			Namespace:   types.NamespaceMiner,
			Description: "Get the number of miner threads",
			Func: func(params interface{}) *jsonrpc.Response {
				return jsonrpc.Success(m.numThreads)
			},
		},
		"setThreads": {
			Private:     true,
			Namespace:   types.NamespaceMiner,
			Description: "Set the number of miner threads",
			Func:        m.apiSetThreads,
		},
	}
}
