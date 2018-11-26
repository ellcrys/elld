package node

import (
	"encoding/base64"

	"github.com/ellcrys/elld/config"

	"github.com/ellcrys/elld/rpc"
	"github.com/ellcrys/elld/rpc/jsonrpc"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/types/core"
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
		"id":                 n.ID().Pretty(),
		"address":            n.GetAddress().ConnectionString(),
		"mode":               mode,
		"netVersion":         config.ProtocolVersion,
		"syncing":            n.isSyncing(),
		"coinbasePublicKey":  n.signatory.PubKey().Base58(),
		"coinbase":           n.signatory.Addr(),
		"buildVersion":       n.cfg.VersionInfo.BuildVersion,
		"buildCommit":        n.cfg.VersionInfo.BuildCommit,
		"buildDate":          n.cfg.VersionInfo.BuildDate,
		"goVersion":          n.cfg.VersionInfo.GoVersion,
		"listeningAddresses": n.GetListenAddresses(),
	})
}

func (n *Node) apiGetConfig(arg interface{}) *jsonrpc.Response {
	return jsonrpc.Success(n.cfg)
}

// apiJoin attempts to establish connection
// with a node at the specified address.
func (n *Node) apiJoin(arg interface{}) *jsonrpc.Response {

	var addrs = []string{}

	// Expect a string or slice
	address, isStr := arg.(string)
	addresses, isSlice := arg.([]interface{})
	if !isStr && !isSlice {
		return jsonrpc.Error(types.ErrCodeUnexpectedArgType,
			rpc.ErrMethodArgType("String or Array{String}").Error(), nil)
	}

	// When a slice is provided as argument,
	// Check that all values are string type
	if isSlice {
		for _, val := range addresses {
			if _, isStr := val.(string); !isStr {
				return jsonrpc.Error(types.ErrCodeUnexpectedArgType,
					rpc.ErrMethodArgType(`String or Array{String}`).Error(), nil)
			}
			addrs = append(addrs, val.(string))
		}
	} else {
		if len(address) > 0 {
			addrs = append(addrs, address)
		}
	}

	if len(addrs) == 0 {
		return jsonrpc.Error(types.ErrCodeAddress,
			"one or more addresses are required", nil)
	}

	for _, address := range addrs {

		if !util.IsValidConnectionString(address) {
			return jsonrpc.Error(types.ErrCodeAddress,
				"address ("+address+") format is invalid", nil)
		}

		rp := n.NewRemoteNode(util.AddressFromConnString(address))
		if rp.IsSame(n) {
			return jsonrpc.Error(types.ErrCodeAddress,
				"can't add self ("+address+") as a peer", nil)
		}

		go func(rp core.Engine) {
			n.peerManager.ConnectToNode(rp)
		}(rp)
	}

	return jsonrpc.Success(true)
}

// apiAddPeer adds an address of a
// remote node to the list of known
// addresses. Unlike apiJoin it does
// not initiate a connection.
func (n *Node) apiAddPeer(arg interface{}) *jsonrpc.Response {
	var addrs = []string{}

	// Expect a string or slice
	address, isStr := arg.(string)
	addresses, isSlice := arg.([]interface{})
	if !isStr && !isSlice {
		return jsonrpc.Error(types.ErrCodeUnexpectedArgType,
			rpc.ErrMethodArgType("String or Array{String}").Error(), nil)
	}

	// When a slice is provided as argument,
	// Check that all values are string type
	if isSlice {
		for _, val := range addresses {
			if _, isStr := val.(string); !isStr {
				return jsonrpc.Error(types.ErrCodeUnexpectedArgType,
					rpc.ErrMethodArgType(`String or Array{String}`).Error(), nil)
			}
			addrs = append(addrs, val.(string))
		}
	} else {
		if len(address) > 0 {
			addrs = append(addrs, address)
		}
	}

	if len(addrs) == 0 {
		return jsonrpc.Error(types.ErrCodeAddress,
			"one or more addresses are required", nil)
	}

	for _, address := range addrs {

		if !util.IsValidConnectionString(address) {
			return jsonrpc.Error(types.ErrCodeAddress,
				"address ("+address+") format is invalid", nil)
		}

		rp := n.NewRemoteNode(util.AddressFromConnString(address))
		if rp.IsSame(n) {
			return jsonrpc.Error(types.ErrCodeAddress,
				"can't add self ("+address+") as a peer", nil)
		}

		n.PM().AddPeer(rp)
	}

	return jsonrpc.Success(true)
}

// apiNetStats returns the
// number of peers connected to
func (n *Node) apiNetStats(arg interface{}) *jsonrpc.Response {
	var connsInfo = n.peerManager.ConnMgr().GetConnsCount()
	in, out := connsInfo.Info()
	var result = map[string]int{
		"total":    out + in,
		"inbound":  in,
		"outbound": out,
		"intros":   n.CountIntros(),
	}
	return jsonrpc.Success(result)
}

// apiGetSyncQueueSize returns the
// size of the block hash queue
func (n *Node) apiGetSyncQueueSize(arg interface{}) *jsonrpc.Response {
	return jsonrpc.Success(n.blockHashQueue.Size())
}

func (n *Node) apiForgetPeers(arg interface{}) *jsonrpc.Response {
	n.PM().ForgetPeers()
	return jsonrpc.Success(true)
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
			"isAcquainted": n.PM().IsAcquainted(p),
			"isInbound":    p.IsInbound(),
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
			"isAcquainted": n.PM().IsAcquainted(p),
			"isInbound":    p.IsInbound(),
			"isBanned":     n.peerManager.IsBanned(p),
			"banEndTime":   n.peerManager.GetBanTime(p),
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
	return jsonrpc.Success(n.GetSyncStateInfo())
}

// apiTxPoolSizeInfo fetches the size information
// of the transaction pool
func (n *Node) apiTxPoolSizeInfo(arg interface{}) *jsonrpc.Response {
	return jsonrpc.Success(map[string]int64{
		"byteSize": n.GetBlockchain().GetTxPool().ByteSize(),
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
	txData["type"] = core.TxTypeBalance

	// Copy data in txData to a core.Transaction
	var tx core.Transaction
	util.MapDecode(txData, &tx)

	// The signature being of type []uint8, will be
	// encoded to base64 by the json encoder.
	// We must convert the base64 back to []uint8
	if sig := txData["sig"]; sig != nil {
		tx.Sig, _ = base64.StdEncoding.DecodeString(sig.(string))
	}

	// Attempt to add the transaction to the pool
	if err := n.AddTransaction(&tx); err != nil {
		return jsonrpc.Error(types.ErrCodeTxFailed, err.Error(), nil)
	}

	return jsonrpc.Success(map[string]interface{}{
		"id": tx.GetID(),
	})
}

// apiFetchPool fetches transactions currently in the pool
func (n *Node) apiFetchPool(arg interface{}) *jsonrpc.Response {
	var txs []types.Transaction
	n.GetTxPool().Container().IFind(func(tx types.Transaction) bool {
		txs = append(txs, tx)
		return false
	})
	return jsonrpc.Success(txs)
}

// APIs returns all API handlers
func (n *Node) APIs() jsonrpc.APISet {
	return map[string]jsonrpc.APIInfo{

		// namespace: "logger"
		"debug": {
			Namespace:   types.NamespaceLogger,
			Description: "Set log level to DEBUG",
			Func: func(arg interface{}) *jsonrpc.Response {
				n.log.SetToDebug()
				return jsonrpc.Success(true)
			},
		},

		"default": {
			Namespace:   types.NamespaceLogger,
			Description: "Set log level to the default (INFO)",
			Func: func(arg interface{}) *jsonrpc.Response {
				n.log.SetToInfo()
				return jsonrpc.Success(true)
			},
		},

		// namespace: "node"
		"config": {
			Namespace:   types.NamespaceNode,
			Description: "Get node configurations",
			Private:     true,
			Func:        n.apiGetConfig,
		},
		"info": {
			Namespace:   types.NamespaceNode,
			Description: "Get basic information of the node",
			Private:     true,
			Func:        n.apiBasicNodeInfo,
		},
		"isSyncing": {
			Namespace:   types.NamespaceNode,
			Description: "Check whether blockchain synchronization is active",
			Func:        n.apiIsSyncing,
		},
		"getSyncState": {
			Namespace:   types.NamespaceNode,
			Description: "Get blockchain synchronization status",
			Func:        n.apiGetSyncState,
		},
		"getSyncQueueSize": {
			Namespace:   types.NamespaceNode,
			Description: "Get number of block hashes in the sync queue",
			Func:        n.apiGetSyncQueueSize,
		},

		// namespace: "ell"
		"send": {
			Namespace:   types.NamespaceEll,
			Description: "Create a balance transaction",
			Private:     true,
			Func:        n.apiSend,
		},

		// namespace: "net"
		"join": {
			Namespace:   types.NamespaceNet,
			Description: "Connect to a peer",
			Private:     true,
			Func:        n.apiJoin,
		},
		"addPeer": {
			Namespace:   types.NamespaceNet,
			Description: "Add a peer address",
			Private:     true,
			Func:        n.apiAddPeer,
		},
		"stats": {
			Namespace:   types.NamespaceNet,
			Description: "Get number connections and network nodes",
			Func:        n.apiNetStats,
		},
		"getPeers": {
			Namespace:   types.NamespaceNet,
			Description: "Get a list of all peers",
			Func:        n.apiGetPeers,
		},
		"getActivePeers": {
			Namespace:   types.NamespaceNet,
			Description: "Get a list of active peers",
			Func:        n.apiGetActivePeers,
		},
		"dumpPeers": {
			Namespace:   types.NamespaceNet,
			Description: "Delete all peers in memory and on disk",
			Func:        n.apiForgetPeers,
		},

		// namespace: "pool"
		"getSize": {
			Namespace:   types.NamespacePool,
			Description: "Get size information of the transaction pool",
			Func:        n.apiTxPoolSizeInfo,
		},
		"getAll": {
			Namespace:   types.NamespacePool,
			Description: "Get transactions in the pool",
			Func:        n.apiFetchPool,
		},
	}
}
