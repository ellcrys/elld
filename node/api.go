package node

import (
	"github.com/ellcrys/elld/rpc"
	"github.com/ellcrys/elld/rpc/jsonrpc"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/types/core/objects"
	"github.com/ellcrys/elld/util"
)

// apiSendTx sends adds a transaction to transaction pool
func (n *Node) apiSendTx(params interface{}) *jsonrpc.Response {

	p, ok := params.(map[string]interface{})
	if !ok {
		return jsonrpc.Error(types.ErrCodeUnexpectedArgType, rpc.ErrMethodArgType("JSON").Error(), nil)
	}

	var tx objects.Transaction
	util.MapDecode(p, &tx)

	return jsonrpc.Success(n.addTransaction(&tx))
}

// apiJoin attempts to establish connection with a node
// at the specified address.
func (n *Node) apiJoin(arg interface{}) *jsonrpc.Response {

	address, ok := arg.(string)
	if !ok {
		return jsonrpc.Error(types.ErrCodeUnexpectedArgType, rpc.ErrMethodArgType("String").Error(), nil)
	}

	n.AddBootstrapNodes([]string{address}, true)
	rn, err := n.NodeFromAddr(address, true)
	if err != nil {
		return jsonrpc.Error(types.ErrCodeAddress, err.Error(), nil)
	}
	rn.isHardcodedSeed = true

	if err := n.connectToNode(rn); err != nil {
		return jsonrpc.Error(types.ErrCodeNodeConnect, err.Error(), nil)
	}

	return jsonrpc.Success(nil)
}

// apiNumConnections returns the number of peers connected to
func (n *Node) apiNumConnections(arg interface{}) *jsonrpc.Response {
	return jsonrpc.Success(n.peerManager.connMgr.connectionCount())
}

// apiGetActivePeers fetches active peers
func (n *Node) apiGetActivePeers(arg interface{}) *jsonrpc.Response {
	var peers = []map[string]interface{}{}
	for _, p := range n.peerManager.GetActivePeers(0) {
		peers = append(peers, map[string]interface{}{
			"id":          p.StringID(),
			"lastSeen":    p.GetTimestamp(),
			"connected":   p.Connected(),
			"isHardcoded": p.IsHardcodedSeed(),
		})
	}
	return jsonrpc.Success(peers)
}

// APIs returns all API handlers
func (n *Node) APIs() jsonrpc.APISet {
	return map[string]jsonrpc.APIInfo{
		"sendTx": jsonrpc.APIInfo{
			Private: true,
			Func:    n.apiSendTx,
		},
		"join": jsonrpc.APIInfo{
			Private: true,
			Func:    n.apiJoin,
		},
		"numConnections": jsonrpc.APIInfo{
			Func: n.apiNumConnections,
		},
		"getActivePeers": jsonrpc.APIInfo{
			Func: n.apiGetActivePeers,
		},
	}
}
