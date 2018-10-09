package node

import (
	"encoding/base64"
	"time"

	"github.com/ellcrys/elld/config"

	"github.com/ellcrys/elld/rpc"
	"github.com/ellcrys/elld/rpc/jsonrpc"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/types/core/objects"
	"github.com/ellcrys/elld/util"
)

// apiBasicNodeInfo returns basic
// information about the node.
func (n *Node) apiBasicNodeInfo(arg interface{}) *jsonrpc.Response {

	var mode = "development"
	if n.ProdMode() {
		mode = "production"
	} else if n.TestMode() {
		mode = "test"
	}

	return jsonrpc.Success(map[string]interface{}{
		"id":                n.ID().Pretty(),
		"address":           n.GetAddress(),
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

// apiJoin attempts to establish connection
// with a node at the specified address.
func (n *Node) apiJoin(arg interface{}) *jsonrpc.Response {
	address, ok := arg.(string)
	if !ok {
		return jsonrpc.Error(types.ErrCodeUnexpectedArgType,
			rpc.ErrMethodArgType("String").Error(), nil)
	}

	if !util.IsValidConnectionString(address) {
		return jsonrpc.Error(types.ErrCodeAddress,
			"address format is invalid", nil)
	}

	rp, err := n.NodeFromAddr(util.AddressFromConnString(address), true)
	if err != nil {
		return jsonrpc.Error(types.ErrCodeAddress, err.Error(), nil)
	}

	if rp.IsSame(n) {
		return jsonrpc.Error(types.ErrCodeAddress,
			"can't add own address as a peer", nil)
	}

	go func(rp *Node) {
		n.connectToNode(rp)
	}(rp)

	return jsonrpc.Success(nil)
}

// apiAddPeer adds an address of a
// remote node to the list of known
// addresses. Unlike apiJoin it does
// not initiate a connection.
func (n *Node) apiAddPeer(arg interface{}) *jsonrpc.Response {
	address, ok := arg.(string)
	if !ok {
		return jsonrpc.Error(types.ErrCodeUnexpectedArgType,
			rpc.ErrMethodArgType("String").Error(), nil)
	}

	if !util.IsValidConnectionString(address) {
		return jsonrpc.Error(types.ErrCodeAddress,
			"address format is invalid", nil)
	}

	rp, err := n.NodeFromAddr(util.AddressFromConnString(address), true)
	if err != nil {
		return jsonrpc.Error(types.ErrCodeAddress, err.Error(), nil)
	}

	rp.lastSeen = time.Now().UTC()

	if rp.IsSame(n) {
		return jsonrpc.Error(types.ErrCodeAddress,
			"can't add self as a peer", nil)
	}

	n.PM().AddPeer(rp)

	return jsonrpc.Success(nil)
}

// apiNumConnections returns the
// number of peers connected to
func (n *Node) apiNumConnections(arg interface{}) *jsonrpc.Response {
	return jsonrpc.Success(n.peerManager.connMgr.connectionCount())
}

// apiGetSyncQueueSize returns the
// size of the block hash queue
func (n *Node) apiGetSyncQueueSize(arg interface{}) *jsonrpc.Response {
	return jsonrpc.Success(n.blockHashQueue.Size())
}

// apiGetActivePeers fetches active peers
func (n *Node) apiGetActivePeers(arg interface{}) *jsonrpc.Response {
	var peers = []map[string]interface{}{}
	for _, p := range n.peerManager.GetActivePeers(0) {
		peers = append(peers, map[string]interface{}{
			"id":           p.StringID(),
			"lastSeen":     p.GetLastSeen(),
			"connected":    p.Connected(),
			"isHardcoded":  p.IsHardcodedSeed(),
			"isAcquainted": p.IsAcquainted(),
		})
	}
	return jsonrpc.Success(peers)
}

// apiGetPeers fetches all peers
func (n *Node) apiGetPeers(arg interface{}) *jsonrpc.Response {
	var peers = []map[string]interface{}{}
	for _, p := range n.peerManager.GetPeers() {
		peers = append(peers, map[string]interface{}{
			"id":           p.StringID(),
			"lastSeen":     p.GetLastSeen(),
			"connected":    p.Connected(),
			"isHardcoded":  p.IsHardcodedSeed(),
			"isAcquainted": p.IsAcquainted(),
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
		return jsonrpc.Error(types.ErrCodeUnexpectedArgType,
			rpc.ErrMethodArgType("JSON").Error(), nil)
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

// apiFetchPool fetches transactions currently in the pool
func (n *Node) apiFetchPool(arg interface{}) *jsonrpc.Response {
	var txs []core.Transaction
	n.GetTxPool().Container().IFind(func(tx core.Transaction) bool {
		txs = append(txs, tx)
		return false
	})
	return jsonrpc.Success(txs)
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
		"addPeer": {
			Namespace:   "net",
			Description: "Add a peer address",
			Private:     true,
			Func:        n.apiAddPeer,
		},
		"numConnections": {
			Namespace:   "net",
			Description: "Get number of active connections",
			Func:        n.apiNumConnections,
		},
		"getPeers": {
			Namespace:   "net",
			Description: "Get a list of all peers",
			Func:        n.apiGetPeers,
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
		"getSyncQueueSize": {
			Namespace:   "node",
			Description: "Get number of block hashes in the sync queue",
			Func:        n.apiGetSyncQueueSize,
		},
		"fetchPool": {
			Namespace:   "node",
			Description: "Get the transactions in the pool",
			Func:        n.apiFetchPool,
		},
	}
}
