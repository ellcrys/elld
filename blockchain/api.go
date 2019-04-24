package blockchain

import (
	"fmt"

	"github.com/thoas/go-funk"

	"github.com/mitchellh/mapstructure"

	"github.com/ellcrys/elld/rpc"
	"github.com/ellcrys/elld/rpc/jsonrpc"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/util"
)

// apiGetChains fetches a list of chains
func (b *Blockchain) apiGetChains(interface{}) *jsonrpc.Response {
	b.lock.RLock()
	defer b.lock.RUnlock()

	var result []interface{}
	for id, chain := range b.chains {
		ch := map[string]interface{}{
			"id":        id,
			"timestamp": chain.info.GetTimestamp(),
		}

		tip, err := chain.Current()
		if err != nil {
			return jsonrpc.Error(types.ErrCodeQueryFailed, err.Error(), nil)
		}

		ch["height"] = tip.GetNumber()
		ch["totalDifficulty"] = tip.GetTotalDifficulty()

		if chain.HasParent() {
			parent := chain.parentBlock
			ch["parentBlockHash"] = parent.GetHash().HexStr()
			ch["parentBlockNumber"] = parent.GetNumber()
			ch["isBranch"] = true
			ch["length"] = tip.GetNumber() - parent.GetNumber()
		}

		result = append(result, util.EncodeForJS(ch))
	}

	return jsonrpc.Success(result)
}

// apiGetBlock fetches a block by number
func (b *Blockchain) apiGetBlock(arg interface{}) *jsonrpc.Response {
	b.lock.RLock()
	defer b.lock.RUnlock()

	if b.bestChain == nil {
		return jsonrpc.Error(types.ErrCodeQueryFailed, "best chain not set", nil)
	}

	num, ok := arg.(float64)
	if !ok {
		return jsonrpc.Error(types.ErrCodeUnexpectedArgType,
			rpc.ErrMethodArgType("Integer").Error(), nil)
	}

	block, err := b.bestChain.GetBlock(uint64(num))
	if err != nil {
		if err != core.ErrBlockNotFound {
			return jsonrpc.Error(types.ErrCodeQueryFailed,
				err.Error(), nil)
		}
		return jsonrpc.Error(types.ErrCodeBlockNotFound,
			err.Error(), nil)
	}

	return jsonrpc.Success(util.EncodeForJS(block))
}

// apiGetMinedBlocks fetches blocks mined by this node
func (b *Blockchain) apiGetMinedBlocks(arg interface{}) *jsonrpc.Response {
	b.lock.RLock()
	defer b.lock.RUnlock()

	var opt core.ArgGetMinedBlock
	if arg != nil {
		decoded, ok := arg.(map[string]interface{})
		if !ok {
			return jsonrpc.Error(types.ErrCodeUnexpectedArgType,
				rpc.ErrMethodArgType("Map").Error(), nil)
		} else {
			mapstructure.Decode(decoded, &opt)
		}
	}

	result, hasMore, err := b.bestChain.GetMinedBlocks(&opt)
	if err != nil {
		return jsonrpc.Error(types.ErrCodeQueryFailed,
			err.Error(), nil)
	}

	var friendlyResult = []interface{}{}
	for _, r := range result {
		friendlyResult = append(friendlyResult, util.EncodeForJS(r, "timestamp"))
	}

	return jsonrpc.Success(map[string]interface{}{
		"blocks":  friendlyResult,
		"hasMore": hasMore,
	})
}

// apiGetTipBlock fetches the highest block on the main chain
func (b *Blockchain) apiGetTipBlock(arg interface{}) *jsonrpc.Response {
	b.lock.RLock()
	defer b.lock.RUnlock()

	if b.bestChain == nil {
		return jsonrpc.Error(types.ErrCodeQueryFailed, "best chain not set", nil)
	}

	block, err := b.bestChain.GetBlock(0)
	if err != nil {
		if err != core.ErrBlockNotFound {
			return jsonrpc.Error(types.ErrCodeQueryFailed,
				err.Error(), nil)
		}
		return jsonrpc.Error(types.ErrCodeBlockNotFound,
			err.Error(), nil)
	}

	return jsonrpc.Success(util.EncodeForJS(block))
}

// apiGetBlockByHash fetches a block by hash
func (b *Blockchain) apiGetBlockByHash(arg interface{}) *jsonrpc.Response {
	b.lock.RLock()
	defer b.lock.RUnlock()

	if b.bestChain == nil {
		return jsonrpc.Error(types.ErrCodeQueryFailed, "best chain not set", nil)
	}

	hash, ok := arg.(string)
	if !ok {
		return jsonrpc.Error(types.ErrCodeUnexpectedArgType,
			rpc.ErrMethodArgType("String").Error(), nil)
	}
	blockHash, err := util.HexToHash(hash)
	if err != nil {
		return jsonrpc.Error(types.ErrCodeQueryFailed,
			"invalid block hash", nil)
	}

	block, err := b.bestChain.getBlockByHash(blockHash)
	if err != nil {
		if err != core.ErrBlockNotFound {
			return jsonrpc.Error(types.ErrCodeQueryFailed,
				err.Error(), nil)
		}
		return jsonrpc.Error(types.ErrCodeBlockNotFound,
			err.Error(), nil)
	}

	return jsonrpc.Success(util.EncodeForJS(block))
}

// apiGetOrphans fetches all orphan blocks
func (b *Blockchain) apiGetOrphans(arg interface{}) *jsonrpc.Response {
	b.lock.RLock()
	defer b.lock.RUnlock()
	var orphans = []interface{}{}
	for _, k := range b.orphanBlocks.Keys() {
		orphans = append(orphans,
			util.EncodeForJS(b.orphanBlocks.Peek(k)))
	}
	return jsonrpc.Success(orphans)
}

// apiGetBestchain fetches the best chain
func (b *Blockchain) apiGetBestchain(arg interface{}) *jsonrpc.Response {
	if b.bestChain == nil {
		return jsonrpc.Error(types.ErrCodeQueryFailed, "best chain not set", nil)
	}

	tip, err := b.bestChain.Current()
	if err != nil {
		return jsonrpc.Error(types.ErrCodeQueryFailed,
			err.Error(), nil)
	}

	return jsonrpc.Success(util.EncodeForJS(map[string]interface{}{
		"id":              b.bestChain.id,
		"timestamp":       b.bestChain.info.GetTimestamp(),
		"height":          tip.GetNumber(),
		"totalDifficulty": tip.GetTotalDifficulty(),
	}))
}

// apiGetReOrgs fetches the re-organization records
func (b *Blockchain) apiGetReOrgs(arg interface{}) *jsonrpc.Response {
	return jsonrpc.Success(b.getReOrgs())
}

// apiGetAccount gets an account
func (b *Blockchain) apiGetAccount(arg interface{}) *jsonrpc.Response {
	address, ok := arg.(string)
	if !ok {
		return jsonrpc.Error(types.ErrCodeUnexpectedArgType,
			rpc.ErrMethodArgType("String").Error(), nil)
	}

	account, err := b.GetAccount(util.String(address))
	if err != nil {
		return jsonrpc.Error(types.ErrCodeAccountNotFound,
			err.Error(), nil)
	}

	return jsonrpc.Success(account)
}

// apiGetAccount gets the nonce of an account
func (b *Blockchain) apiGetNonce(arg interface{}) *jsonrpc.Response {

	address, ok := arg.(string)
	if !ok {
		return jsonrpc.Error(types.ErrCodeUnexpectedArgType,
			rpc.ErrMethodArgType("String").Error(), nil)
	}

	account, err := b.GetAccount(util.String(address))
	if err != nil {
		return jsonrpc.Error(types.ErrCodeAccountNotFound,
			err.Error(), nil)
	}

	return jsonrpc.Success(account.GetNonce())
}

// apiGetBalance gets the balance of an account
func (b *Blockchain) apiGetBalance(arg interface{}) *jsonrpc.Response {

	address, ok := arg.(string)
	if !ok {
		return jsonrpc.Error(types.ErrCodeUnexpectedArgType,
			rpc.ErrMethodArgType("String").Error(), nil)
	}

	account, err := b.GetAccount(util.String(address))
	if err != nil {
		return jsonrpc.Error(types.ErrCodeAccountNotFound,
			err.Error(), nil)
	}

	return jsonrpc.Success(account.GetBalance())
}

// apiGetTransaction gets a transaction by hash
func (b *Blockchain) apiGetTransaction(arg interface{}) *jsonrpc.Response {

	txHash, ok := arg.(string)
	if !ok {
		return jsonrpc.Error(types.ErrCodeUnexpectedArgType,
			rpc.ErrMethodArgType("String").Error(), nil)
	}

	hash, err := util.HexToHash(txHash)
	if err != nil {
		return jsonrpc.Error(
			types.ErrCodeQueryParamError,
			fmt.Sprintf("invalid transaction id: %s", err.Error()),
			nil,
		)
	}

	tx, err := b.GetTransaction(hash)
	if err != nil {
		return jsonrpc.Error(types.ErrCodeTransactionNotFound,
			err.Error(), nil)
	}

	return jsonrpc.Success(util.EncodeForJS(tx))
}

// apiGetTransactionStatus gets the status of
// a transaction matching a given hash.
// Status: 'unknown' - not found, 'pooled' - in
// the transaction pool & 'mined' - In a mined block.
func (b *Blockchain) apiGetTransactionStatus(arg interface{}) *jsonrpc.Response {

	txHash, ok := arg.(string)
	if !ok {
		return jsonrpc.Error(types.ErrCodeUnexpectedArgType,
			rpc.ErrMethodArgType("String").Error(), nil)
	}

	var status = "unknown"

	if b.txPool.HasByHash(txHash) {
		status = "pooled"
	}

	hash, err := util.HexToHash(txHash)
	if err != nil {
		return jsonrpc.Error(
			types.ErrCodeQueryParamError,
			fmt.Sprintf("invalid transaction id: %s", err.Error()),
			nil,
		)
	}

	tx, err := b.GetTransaction(hash)
	if err != nil {
		if err != core.ErrTxNotFound {
			return jsonrpc.Error(types.ErrCodeQueryFailed, err.Error(), nil)
		}
	}

	if tx != nil {
		status = "mined"
	}

	return jsonrpc.Success(map[string]interface{}{
		"status": status,
	})

}

// apiGetTransactionFromPool gets the transaction
// matching a given hash from the transaction pool.
func (b *Blockchain) apiGetTransactionFromPool(arg interface{}) *jsonrpc.Response {

	txHash, ok := arg.(string)
	if !ok {
		return jsonrpc.Error(types.ErrCodeUnexpectedArgType,
			rpc.ErrMethodArgType("String").Error(), nil)
	}

	_, err := util.HexToHash(txHash)
	if err != nil {
		return jsonrpc.Error(
			types.ErrCodeQueryParamError,
			fmt.Sprintf("invalid transaction id: %s", err.Error()),
			nil,
		)
	}

	if !b.txPool.HasByHash(txHash) {
		return jsonrpc.Error(types.ErrCodeTransactionNotFound,
			"transaction not found", nil)
	}

	tx := b.txPool.GetByHash(txHash)

	return jsonrpc.Success(util.EncodeForJS(tx))

}

// apiGetDifficultyInfo gets the difficulty and total
// difficulty of the main chain
func (b *Blockchain) apiGetDifficultyInfo(arg interface{}) *jsonrpc.Response {

	if b.bestChain == nil {
		return jsonrpc.Error(types.ErrCodeQueryFailed, "best chain not set", nil)
	}

	tip, err := b.bestChain.Current()
	if err != nil {
		if err != core.ErrBlockNotFound {
			return jsonrpc.Error(types.ErrCodeQueryFailed, err.Error(), nil)
		}
		return jsonrpc.Error(types.ErrCodeBlockNotFound, err.Error(), nil)
	}

	return jsonrpc.Success(util.EncodeForJS(map[string]interface{}{
		"difficulty":      tip.GetDifficulty(),
		"totalDifficulty": tip.GetTotalDifficulty(),
	}))
}

// apiListAccounts gets all accounts
func (b *Blockchain) apiListAccounts(arg interface{}) *jsonrpc.Response {

	if b.bestChain == nil {
		return jsonrpc.Error(types.ErrCodeQueryFailed, "best chain not set", nil)
	}

	accounts, err := b.bestChain.GetAccounts()
	if err != nil {
		return jsonrpc.Error(types.ErrCodeQueryFailed, err.Error(), nil)
	}

	return jsonrpc.Success(accounts)
}

// apiListTopAccounts gets the top N accounts
func (b *Blockchain) apiListTopNAccounts(arg interface{}) *jsonrpc.Response {

	numAccts, ok := arg.(float64)
	if !ok {
		return jsonrpc.Error(types.ErrCodeUnexpectedArgType,
			rpc.ErrMethodArgType("Integer").Error(), nil)
	}

	if b.bestChain == nil {
		return jsonrpc.Error(types.ErrCodeQueryFailed, "best chain not set", nil)
	}

	accounts, err := b.ListTopAccounts(int(numAccts))
	if err != nil {
		return jsonrpc.Error(types.ErrCodeQueryFailed, err.Error(), nil)
	}

	return jsonrpc.Success(accounts)
}

// apiSuggestNonce attempts to determine the next nonce of an account.
func (b *Blockchain) apiSuggestNonce(arg interface{}) *jsonrpc.Response {

	address, ok := arg.(string)
	if !ok {
		return jsonrpc.Error(types.ErrCodeUnexpectedArgType,
			rpc.ErrMethodArgType("String").Error(), nil)
	}

	// Get the current account nonce
	accountNonce, err := b.GetAccountNonce(util.String(address))
	if err != nil {
		if err == core.ErrAccountNotFound {
			return jsonrpc.Error(types.ErrCodeAccountNotFound,
				fmt.Sprintf("invalid transaction id: %s", err.Error()), nil)
		}
		return jsonrpc.Error(types.ErrCodeUnexpected, "unexpected error", err.Error())
	}

	suggested := accountNonce + 1

	// Get the nonce of any transactions in the pool
	// that originated from the account
	var txPoolNonces = []uint64{}
	for _, tx := range b.txPool.GetByFrom(util.String(address)) {
		nonce := tx.GetNonce()
		if !funk.Contains(txPoolNonces, nonce) {
			txPoolNonces = append(txPoolNonces, nonce)
		}
	}

	// We will now attempt to update the suggested
	// nonce based on other nonces of transactions
	// from the same account. When there is a nonce
	// greater than the suggested nonce, we keep
	// the suggested nonce. When there is a nonce
	// equal to the suggested nonce, we increment
	// the suggested nonce by 1 and start the loop
	// again
	for i := 0; i < len(txPoolNonces); i++ {
		var nonce = txPoolNonces[i]
		if suggested == nonce {
			suggested++
			i = 0
			continue
		}
	}

	// pp.Println(txPoolNonces, suggested)
	return jsonrpc.Success(suggested)
}

// APIs returns all API handlers
func (b *Blockchain) APIs() jsonrpc.APISet {
	return map[string]jsonrpc.APIInfo{

		// namespace: "state"
		"getBranches": {
			Namespace:   types.NamespaceState,
			Description: "Get all branches",
			Func:        b.apiGetChains,
		},
		"getBlock": {
			Namespace:   types.NamespaceState,
			Description: "Get a block by number",
			Func:        b.apiGetBlock,
		},
		"getTipBlock": {
			Namespace:   types.NamespaceState,
			Description: "Get a highest block on the main chain",
			Func:        b.apiGetTipBlock,
		},
		"getBlockByHash": {
			Namespace:   types.NamespaceState,
			Description: "Get a block by hash",
			Func:        b.apiGetBlockByHash,
		},
		"getMinedBlocks": {
			Namespace:   types.NamespaceState,
			Description: "Get blocks mined on this node",
			Func:        b.apiGetMinedBlocks,
		},
		"getOrphans": {
			Namespace:   types.NamespaceState,
			Description: "Get a list of orphans",
			Func:        b.apiGetOrphans,
		},
		"getBestChain": {
			Namespace:   types.NamespaceState,
			Description: "Get the best chain",
			Func:        b.apiGetBestchain,
		},
		"getReOrgs": {
			Namespace:   types.NamespaceState,
			Description: "Get a list of re-organization events",
			Func:        b.apiGetReOrgs,
		},
		"getAccount": {
			Namespace:   types.NamespaceState,
			Description: "Get an account",
			Func:        b.apiGetAccount,
		},
		"listAccounts": {
			Namespace:   types.NamespaceState,
			Description: "List all accounts",
			Func:        b.apiListAccounts,
		},
		"listTopAccounts": {
			Namespace:   types.NamespaceState,
			Description: "List top accounts",
			Func:        b.apiListTopNAccounts,
		},
		"getAccountNonce": {
			Namespace:   types.NamespaceState,
			Description: "Get the nonce of an account",
			Func:        b.apiGetNonce,
		},
		"getTransaction": {
			Namespace:   types.NamespaceState,
			Description: "Get a transaction by hash",
			Func:        b.apiGetTransaction,
		},
		"getDifficulty": {
			Namespace:   types.NamespaceState,
			Description: "Get difficulty information",
			Func:        b.apiGetDifficultyInfo,
		},
		"getObjects": {
			Namespace:   types.NamespaceState,
			Description: "Get raw database objects (for debugging)",
			Func:        b.apiGetDBObjects,
		},
		"suggestNonce": {
			Namespace:   types.NamespaceState,
			Description: "Suggest an account nonce to use in a new transaction",
			Func:        b.apiSuggestNonce,
		},

		// namespace: "node"
		"getTransactionStatus": {
			Namespace:   types.NamespaceNode,
			Description: "Get a transaction's status",
			Func:        b.apiGetTransactionStatus,
		},
		"getTransactionFromPool": {
			Namespace:   types.NamespaceNode,
			Description: "Get a transaction by hash from pool",
			Func:        b.apiGetTransactionFromPool,
		},
		// namespace: "ell"
		"getBalance": {
			Namespace:   types.NamespaceEll,
			Description: "Get account balance",
			Func:        b.apiGetBalance,
		},
	}
}
