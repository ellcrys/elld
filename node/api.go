package node

import (
	"encoding/base64"

	"github.com/ellcrys/elld/config"

	"github.com/ellcrys/elld/rpc"
	"github.com/ellcrys/elld/rpc/jsonrpc"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/types/core/objects"
	"github.com/ellcrys/elld/util"
)

// apiBasicNodeInfo returns basic information
// about the node.
func (n *Node) apiBasicNodeInfo(arg interface{}) *jsonrpc.Response {

	var mode = "development"
	if n.ProdMode() {
		mode = "production"
	} else if n.TestMode() {
		mode = "test"
	}

	return jsonrpc.Success(map[string]interface{}{
		"id":                n.ID().Pretty(),
		"address":           n.GetMultiAddr(),
		"mode":              mode,
		"netVersion":        config.ProtocolVersion,
		"syncing":           n.isSyncing(),
		"coinbasePublicKey": n.signatory.PubKey().Base58(),
		"coinbase":          n.signatory.Addr(),
	})
}

func (n *Node) apiGetConfig(arg interface{}) *jsonrpc.Response {
	return jsonrpc.Success(n.cfg)
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
		return jsonrpc.Error(types.ErrCodeNodeConnectFailure, err.Error(), nil)
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

// apiIsSyncing fetches the sync status
func (n *Node) apiIsSyncing(arg interface{}) *jsonrpc.Response {
	return jsonrpc.Success(n.isSyncing())
}

// apiGetSyncState fetches the sync status
func (n *Node) apiGetSyncState(arg interface{}) *jsonrpc.Response {
	return jsonrpc.Success(n.getSyncStateInfo())
}

// apiTxPoolSizeInfo fetches the size information
// of the transaction pool
func (n *Node) apiTxPoolSizeInfo(arg interface{}) *jsonrpc.Response {
	return jsonrpc.Success(map[string]int64{
		"byteSize": n.GetTxPool().ByteSize(),
		"numTxs":   n.GetTxPool().Size(),
	})
}

// apiSend creates a balance transaction and
// attempts to add it to the transaction pool.
func (n *Node) apiSend(arg interface{}) *jsonrpc.Response {

	txData, ok := arg.(map[string]interface{})
	if !ok {
		return jsonrpc.Error(types.ErrCodeUnexpectedArgType, rpc.ErrMethodArgType("JSON").Error(), nil)
	}
	// set the type to TxTypeBalance.
	// it will override the type given
	txData["type"] = objects.TxTypeBalance

	// Copy data in txData to a core.Transaction
	var tx objects.Transaction
	util.MapDecode(txData, &tx)

	// The signature being of type []uint8, will be
	// encoded to base64 by the json encoder.
	// We must convert the base64 back to []uint8
	if sig := txData["sig"]; sig != nil {
		tx.Sig, _ = base64.StdEncoding.DecodeString(sig.(string))
	}

	// Attempt to add the transaction to the pool
	if err := n.addTransaction(&tx); err != nil {
		return jsonrpc.Error(types.ErrCodeTxFailed, err.Error(), nil)
	}

	return jsonrpc.Success(map[string]interface{}{
		"id": tx.ID(),
	})
}

// APIs returns all API handlers
func (n *Node) APIs() jsonrpc.APISet {
	return map[string]jsonrpc.APIInfo{
		"config": {
			Namespace:   "node",
			Description: "Get node configurations",
			Private:     true,
			Func:        n.apiGetConfig,
		},
		"info": {
			Namespace:   "node",
			Description: "Get basic information of the node",
			Func:        n.apiBasicNodeInfo,
		},
		"send": {
			Namespace:   "ell",
			Description: "Create a balance transaction",
			Private:     true,
			Func:        n.apiSend,
		},
		"join": {
			Namespace:   "net",
			Description: "Connect to a peer",
			Private:     true,
			Func:        n.apiJoin,
		},
		"numConnections": {
			Namespace:   "net",
			Description: "Get number of active connections",
			Func:        n.apiNumConnections,
		},
		"getActivePeers": {
			Namespace:   "net",
			Description: "Get a list of active peers",
			Func:        n.apiGetActivePeers,
		},
		"isSyncing": {
			Namespace:   "node",
			Description: "Check whether blockchain synchronization is active",
			Func:        n.apiIsSyncing,
		},
		"getSyncState": {
			Namespace:   "node",
			Description: "Get blockchain synchronization status",
			Func:        n.apiGetSyncState,
		},
		"getPoolSize": {
			Namespace:   "node",
			Description: "Get size information of the transaction pool",
			Func:        n.apiTxPoolSizeInfo,
		},
	}
}
