package blockchain

import (
	"github.com/ellcrys/elld/rpc"
	"github.com/ellcrys/elld/rpc/jsonrpc"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/util"
)

// apiGetChains fetches a list of chains
func (b *Blockchain) apiGetChains(interface{}) *jsonrpc.Response {
	b.chainLock.RLock()
	defer b.chainLock.RUnlock()

	var result []interface{}
	for id, chain := range b.chains {
		ch := map[string]interface{}{
			"id":        id,
			"timestamp": chain.info.Timestamp,
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
			ch["isSidechain"] = true
			ch["length"] = tip.GetNumber() - parent.GetNumber()
		}

		result = append(result, core.MapFieldsToHex(ch))
	}

	return jsonrpc.Success(result)
}

// apiGetBlock fetches a block by number
func (b *Blockchain) apiGetBlock(arg interface{}) *jsonrpc.Response {
	b.chainLock.RLock()
	defer b.chainLock.RUnlock()

	if b.bestChain == nil {
		return jsonrpc.Error(types.ErrCodeQueryFailed, "best chain not set", nil)
	}

	num, ok := arg.(float64)
	if !ok {
		return jsonrpc.Error(types.ErrCodeUnexpectedArgType, rpc.ErrMethodArgType("Integer").Error(), nil)
	}

	block, err := b.bestChain.GetBlock(uint64(num))
	if err != nil {
		if err != core.ErrBlockNotFound {
			return jsonrpc.Error(types.ErrCodeQueryFailed, err.Error(), nil)
		}
		return jsonrpc.Error(types.ErrCodeBlockNotFound, err.Error(), nil)
	}

	return jsonrpc.Success(core.MapFieldsToHex(block))
}

// apiGetBlockByHash fetches a block by hash
func (b *Blockchain) apiGetBlockByHash(arg interface{}) *jsonrpc.Response {
	b.chainLock.RLock()
	defer b.chainLock.RUnlock()

	if b.bestChain == nil {
		return jsonrpc.Error(types.ErrCodeQueryFailed, "best chain not set", nil)
	}

	hash, ok := arg.(string)
	if !ok {
		return jsonrpc.Error(types.ErrCodeUnexpectedArgType, rpc.ErrMethodArgType("String").Error(), nil)
	}
	blockHash, err := util.HexToHash(hash)
	if err != nil {
		return jsonrpc.Error(types.ErrCodeQueryFailed, "invalid block hash", nil)
	}

	block, err := b.bestChain.getBlockByHash(blockHash)
	if err != nil {
		if err != core.ErrBlockNotFound {
			return jsonrpc.Error(types.ErrCodeQueryFailed, err.Error(), nil)
		}
		return jsonrpc.Error(types.ErrCodeBlockNotFound, err.Error(), nil)
	}

	return jsonrpc.Success(core.MapFieldsToHex(block))
}

// apiGetOrphans fetches all orphan blocks
func (b *Blockchain) apiGetOrphans(arg interface{}) *jsonrpc.Response {
	b.chainLock.RLock()
	defer b.chainLock.RUnlock()
	var orphans = []interface{}{}
	for _, k := range b.orphanBlocks.Keys() {
		orphans = append(orphans, core.MapFieldsToHex(b.orphanBlocks.Peek(k)))
	}
	return jsonrpc.Success(orphans)
}

// apiGetBestchain fetches the best chain
func (b *Blockchain) apiGetBestchain(arg interface{}) *jsonrpc.Response {
	b.chainLock.RLock()
	defer b.chainLock.RUnlock()

	if b.bestChain == nil {
		return jsonrpc.Error(types.ErrCodeQueryFailed, "best chain not set", nil)
	}

	tip, err := b.bestChain.Current()
	if err != nil {
		return jsonrpc.Error(types.ErrCodeQueryFailed, err.Error(), nil)
	}

	return jsonrpc.Success(core.MapFieldsToHex(map[string]interface{}{
		"id":              b.bestChain.id,
		"timestamp":       b.bestChain.info.Timestamp,
		"height":          tip.GetNumber(),
		"totalDifficulty": tip.GetTotalDifficulty(),
	}))
}

// apiGetReOrgs fetches the re-organization records
func (b *Blockchain) apiGetReOrgs(arg interface{}) *jsonrpc.Response {
	b.chainLock.RLock()
	defer b.chainLock.RUnlock()
	return jsonrpc.Success(b.getReOrgs())
}

// apiGetAccount gets an account
func (b *Blockchain) apiGetAccount(arg interface{}) *jsonrpc.Response {
	b.chainLock.RLock()
	defer b.chainLock.RUnlock()

	address, ok := arg.(string)
	if !ok {
		return jsonrpc.Error(types.ErrCodeUnexpectedArgType, rpc.ErrMethodArgType("String").Error(), nil)
	}

	account, err := b.GetAccount(util.String(address))
	if err != nil {
		return jsonrpc.Error(types.ErrCodeAccountNotFound, err.Error(), nil)
	}

	return jsonrpc.Success(account)
}

// apiGetAccount gets the nonce of an account
func (b *Blockchain) apiGetNonce(arg interface{}) *jsonrpc.Response {
	b.chainLock.RLock()
	defer b.chainLock.RUnlock()

	address, ok := arg.(string)
	if !ok {
		return jsonrpc.Error(types.ErrCodeUnexpectedArgType, rpc.ErrMethodArgType("String").Error(), nil)
	}

	account, err := b.GetAccount(util.String(address))
	if err != nil {
		return jsonrpc.Error(types.ErrCodeAccountNotFound, err.Error(), nil)
	}

	return jsonrpc.Success(account.GetNonce())
}

// APIs returns all API handlers
func (b *Blockchain) APIs() jsonrpc.APISet {
	return map[string]jsonrpc.APIInfo{
		"getChains": jsonrpc.APIInfo{
			Namespace:   "node",
			Description: "Get all chains",
			Func:        b.apiGetChains,
		},
		"getBlock": jsonrpc.APIInfo{
			Namespace:   "node",
			Description: "Get a block by number",
			Func:        b.apiGetBlock,
		},
		"getBlockByHash": jsonrpc.APIInfo{
			Namespace:   "node",
			Description: "Get a block by hash",
			Func:        b.apiGetBlockByHash,
		},
		"getOrphans": jsonrpc.APIInfo{
			Namespace:   "node",
			Description: "Get a list of orphans",
			Func:        b.apiGetOrphans,
		},
		"getBestChain": jsonrpc.APIInfo{
			Namespace:   "node",
			Description: "Get the best chain",
			Func:        b.apiGetBestchain,
		},
		"getReOrgs": jsonrpc.APIInfo{
			Namespace:   "node",
			Description: "Get a list of re-organization events",
			Func:        b.apiGetReOrgs,
		},
		"getAccount": jsonrpc.APIInfo{
			Namespace:   "node",
			Description: "Get an account",
			Func:        b.apiGetAccount,
		},
		"getAccountNonce": jsonrpc.APIInfo{
			Namespace:   "node",
			Description: "Get the nonce of an account",
			Func:        b.apiGetNonce,
		},
	}
}
