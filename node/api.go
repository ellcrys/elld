package node

import (
	"encoding/hex"
	"encoding/json"
	"strings"

	"github.com/btcsuite/btcutil/base58"

	"github.com/ellcrys/mother/rpc"
	"github.com/ellcrys/mother/rpc/jsonrpc"
	"github.com/ellcrys/mother/types"
	"github.com/ellcrys/mother/types/core"
	"github.com/ellcrys/mother/util"
)

// apiBasicPublicNodeInfo returns basic
// information about the node that can be
// shared publicly
func (n *Node) apiBasicPublicNodeInfo(arg interface{}) *jsonrpc.Response {
	return nil
}

// apiBasicNodeInfo returns basic
// information about the node.
func (n *Node) apiBasicNodeInfo(arg interface{}) *jsonrpc.Response {
	return nil
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

		// Validated and/or resolve the address
		resolvedAddr, err := util.ValidateAndResolveConnString(address)
		if err != nil {
			return jsonrpc.Error(types.ErrCodeAddress,
				"could not join address ("+address+"): "+err.Error(), nil)
		}

		rp := n.NewRemoteNode(util.AddressFromConnString(resolvedAddr))
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

		// Validated and/or resolve the address
		resolvedAddr, err := util.ValidateAndResolveConnString(address)
		if err != nil {
			return jsonrpc.Error(types.ErrCodeAddress,
				"could not add address ("+address+"): "+err.Error(), nil)
		}

		rp := n.NewRemoteNode(util.AddressFromConnString(resolvedAddr))
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
	}
	return jsonrpc.Success(result)
}

// Clear all caches, disk and memory store of peers
func (n *Node) apiForgetPeers(arg interface{}) *jsonrpc.Response {
	n.PM().ForgetPeers()
	return jsonrpc.Success(true)
}

// apiSyncEnable enables block synchronization
func (n *Node) apiSyncEnable(arg interface{}) *jsonrpc.Response {
	if n.GetSyncMode().IsDisabled() {
		n.log.Debug("Synchronization has been re-enabled")
		n.GetSyncMode().Enable()
	}
	return jsonrpc.Success(true)
}

// apiSyncDisabled disables block synchronization
func (n *Node) apiSyncDisabled(arg interface{}) *jsonrpc.Response {
	if !n.GetSyncMode().IsDisabled() {
		n.log.Debug("Synchronization has been disabled")
		n.GetSyncMode().Disable()
	}
	return jsonrpc.Success(true)
}

// apiIsSyncEnabled disables block synchronization
func (n *Node) apiIsSyncEnabled(arg interface{}) *jsonrpc.Response {
	return jsonrpc.Success(!n.GetSyncMode().IsDisabled())
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
			"name":         p.GetName(),
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
			"name":         p.GetName(),
		})
	}
	return jsonrpc.Success(peers)
}

// apiIsSyncing fetches the sync status
func (n *Node) apiIsSyncing(arg interface{}) *jsonrpc.Response {
	return nil
}

// apiGetSyncStat fetches the sync status
func (n *Node) apiGetSyncStat(arg interface{}) *jsonrpc.Response {
	return nil
}

// apiTxPoolSizeInfo fetches the size information
// of the transaction pool
func (n *Node) apiTxPoolSizeInfo(arg interface{}) *jsonrpc.Response {
	return jsonrpc.Success(map[string]int64{
		"byteSize": n.txsPool.ByteSize(),
		"numTxs":   n.txsPool.Size(),
	})
}

// processTx takes a map that represents a transaction
// and attempts to add it to the pool
func (n *Node) processTx(txData map[string]interface{}) *jsonrpc.Response {

	var tx core.Transaction
	util.MapDecode(txData, &tx)

	// We expect signature to be an hex string
	if sig, ok := txData["sig"].(string); ok {

		// If the signature begins with `0x`,
		// we must strip it away
		if strings.HasPrefix(sig, "0x") {
			sig = sig[2:]
		}

		var err error
		tx.Sig, err = hex.DecodeString(sig)
		if err != nil {
			return jsonrpc.Error(types.ErrCodeTxFailed,
				"signature is not a valid hex string", nil)
		}
	}

	// We expect the hash to be an hex string
	if hash, ok := txData["hash"].(string); ok {

		// If the signature begins with `0x`,
		// we must strip it away
		if strings.HasPrefix(hash, "0x") {
			hash = hash[2:]
		}

		hashBytes, err := hex.DecodeString(hash)
		if err != nil {
			return jsonrpc.Error(types.ErrCodeTxFailed,
				"hash is not a valid hex string", nil)
		}
		tx.Hash = util.BytesToHash(hashBytes)
	}

	// Attempt to add the transaction to the pool
	if err := n.txManager.AddTx(&tx); err != nil {
		return jsonrpc.Error(types.ErrCodeTxFailed, err.Error(), nil)
	}

	return jsonrpc.Success(map[string]interface{}{
		"id": tx.GetID(),
	})
}

// apiSend creates a balance transaction and
// attempts to add it to the transaction pool.
func (n *Node) apiSendRaw(arg interface{}) *jsonrpc.Response {

	encTx, ok := arg.(string)
	if !ok {
		return jsonrpc.Error(types.ErrCodeUnexpectedArgType,
			rpc.ErrMethodArgType("String").Error(), nil)
	}

	// Attempt to decode encoded transaction
	txBytes, v, err := base58.CheckDecode(encTx)
	if err != nil {
		return jsonrpc.Error(types.ErrCodeTxFailed,
			"base58: transaction not correctly encoded", nil)
	}

	// Ensure version is expected
	if v != core.Base58CheckVersionTxPayload {
		return jsonrpc.Error(types.ErrCodeTxFailed,
			"encoded transaction has an invalid version", nil)
	}

	var txData map[string]interface{}
	if err := json.Unmarshal(txBytes, &txData); err != nil {
		return jsonrpc.Error(types.ErrCodeTxFailed,
			"json: tx not correctly encoded", nil)
	}

	return n.processTx(txData)

}

// apiSend creates a balance transaction and
// attempts to add it to the transaction pool.
func (n *Node) apiSend(arg interface{}) *jsonrpc.Response {

	txData, ok := arg.(map[string]interface{})
	if !ok {
		return jsonrpc.Error(types.ErrCodeUnexpectedArgType,
			rpc.ErrMethodArgType("JSON").Error(), nil)
	}

	return n.processTx(txData)
}

// apiFetchPool fetches transactions currently in the pool
func (n *Node) apiFetchPool(arg interface{}) *jsonrpc.Response {
	var txs = []types.Transaction{}
	n.GetTxPool().Container().IFind(func(tx types.Transaction) bool {
		txs = append(txs, tx)
		return false
	})
	return jsonrpc.Success(txs)
}

func (n *Node) apiBroadcastPeers(arg interface{}) *jsonrpc.Response {
	// var result = map[string][]string{
	// 	"broadcasters":       {},
	// 	"randomBroadcasters": {},
	// }
	// for _, p := range n.Gossip().GetBroadcasters().Peers() {
	// 	result["broadcasters"] = append(result["broadcasters"],
	// 		p.GetAddress().ConnectionString())
	// }
	// for _, p := range n.Gossip().GetRandBroadcasters().Peers() {
	// 	result["randomBroadcasters"] = append(result["randomBroadcasters"],
	// 		p.GetAddress().ConnectionString())
	// }
	// return jsonrpc.Success(result)
	return nil
}

func (n *Node) apiNoNetwork(arg interface{}) *jsonrpc.Response {
	n.DisableNetwork()
	n.host.Close()
	return jsonrpc.Success(true)
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
		"basic": {
			Namespace:   types.NamespaceNode,
			Description: "Get basic public information of the node",
			Func:        n.apiBasicPublicNodeInfo,
		},
		"isSyncing": {
			Namespace:   types.NamespaceNode,
			Description: "Check whether blockchain synchronization is active",
			Func:        n.apiIsSyncing,
		},
		"getSyncStat": {
			Namespace:   types.NamespaceNode,
			Description: "Get blockchain synchronization statistic",
			Func:        n.apiGetSyncStat,
		},
		"enableSync": {
			Namespace:   types.NamespaceNode,
			Private:     true,
			Description: "Enable block synchronization",
			Func:        n.apiSyncEnable,
		},
		"disableSync": {
			Namespace:   types.NamespaceNode,
			Private:     true,
			Description: "Disable block synchronization",
			Func:        n.apiSyncDisabled,
		},
		"isSyncEnabled": {
			Namespace:   types.NamespaceNode,
			Description: "Returns whether synchronization is enabled",
			Func:        n.apiIsSyncEnabled,
		},

		// namespace: "ell"
		"send": {
			Namespace:   types.NamespaceEll,
			Description: "Send a balance transaction",
			Private:     true,
			Func:        n.apiSend,
		},
		"sendRaw": {
			Namespace:   types.NamespaceEll,
			Description: "Send a base58 encoded balance transaction",
			Func:        n.apiSendRaw,
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
			Private:     true,
			Description: "Delete all peers in memory and on disk",
			Func:        n.apiForgetPeers,
		},
		"broadcasters": {
			Namespace:   types.NamespaceNet,
			Description: "Get broadcast peers",
			Func:        n.apiBroadcastPeers,
		},
		"noNet": {
			Namespace:   types.NamespaceNet,
			Private:     true,
			Description: "Close the host connection and prevent in/out connections",
			Func:        n.apiNoNetwork,
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
