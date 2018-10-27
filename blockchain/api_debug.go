package blockchain

import (
	"bytes"

	"github.com/ellcrys/elld/blockchain/common"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/util"

	"github.com/ellcrys/elld/elldb"
	"github.com/ellcrys/elld/rpc"
	"github.com/ellcrys/elld/rpc/jsonrpc"
	"github.com/ellcrys/elld/types"
)

// apiGetDBObjects queries raw KV objects
func (b *Blockchain) apiGetDBObjects(arg interface{}) *jsonrpc.Response {

	if b.bestChain == nil {
		return jsonrpc.Error(types.ErrCodeQueryFailed,
			"best chain not set", nil)
	}

	mArgs, ok := arg.(map[string]interface{})
	if !ok {
		return jsonrpc.Error(types.ErrCodeUnexpectedArgType,
			rpc.ErrMethodArgType("JSON").Error(), nil)
	}

	queryType := mArgs["type"]
	if queryType == nil {
		return jsonrpc.Error(types.ErrCodeQueryParamError,
			"type is required", nil)
	}

	var result []map[string]interface{}
	var chainID []byte
	var blockNum uint64
	var address []byte

	if val, ok := mArgs["chainID"].(string); ok {
		chainID = util.String(val).Bytes()
	}

	if val, ok := mArgs["blockNumber"].(float64); ok {
		blockNum = uint64(val)
	}

	if val, ok := mArgs["address"].(string); ok {
		address = util.String(val).Bytes()
	}

	switch queryType {

	// query accounts
	case "accounts":

		// find accounts by chain id only
		if chainID != nil && blockNum == 0 {
			b.db.Iterate(common.MakeQueryKeyAccounts(chainID), true,
				func(kv *elldb.KVObject) bool {
					m := map[string]interface{}{
						"blockNumber": util.DecodeNumber(kv.Key),
						"prefix":      string(kv.Prefix),
					}
					var val map[string]interface{}
					util.BytesToObject(kv.Value, &val)
					m["value"] = val
					result = append(result, m)
					return false
				})
		}

		// find accounts by chain id, block number and address
		if chainID != nil && blockNum != 0 {
			b.db.Iterate(common.MakeKeyAccount(blockNum, chainID, address), true,
				func(kv *elldb.KVObject) bool {
					m := map[string]interface{}{
						"blockNumber": util.DecodeNumber(kv.Key),
						"prefix":      string(kv.Prefix),
					}
					var val map[string]interface{}
					util.BytesToObject(kv.Value, &val)
					m["value"] = val
					result = append(result, m)
					return false
				})
		}

	case "transactions":

		// find transactions in only chain id
		if chainID != nil && blockNum == 0 {
			b.db.Iterate(common.MakeQueryKeyTransactions(chainID), true,
				func(kv *elldb.KVObject) bool {
					m := map[string]interface{}{
						"prefix":      string(kv.Prefix),
						"blockNumber": util.DecodeNumber(kv.Key),
					}
					var val map[string]interface{}
					util.BytesToObject(kv.Value, &val)
					m["value"] = val
					result = append(result, m)
					return false
				})
		}

	case "chains":
		b.db.Iterate(common.MakeQueryKeyChains(), true,
			func(kv *elldb.KVObject) bool {
				var m map[string]interface{}
				util.BytesToObject(kv.Value, &m)
				result = append(result, m)
				return false
			})

	case "all":
		var errResp *jsonrpc.Response
		b.db.Iterate(nil, true,
			func(kv *elldb.KVObject) bool {
				m := map[string]interface{}{
					"prefix": string(kv.Prefix),
					"key":    string(kv.GetKey()),
				}

				// Since by convention, Key usually hold
				// block number, decode it and set to map
				if kv.Key != nil {
					m["blockNumber"] = util.DecodeNumber(kv.Key)
				}

				// If object represents a block, decode
				// the block and set to map
				if bytes.Index(kv.Prefix, common.TagBlock) != -1 {
					var block core.Block
					util.BytesToObject(kv.Value, &block)
					m["value"] = block
				} else {
					// Attempt to decode to map for other types
					var val map[string]interface{}
					err := util.BytesToObject(kv.Value, &val)
					if err != nil {
						errResp = jsonrpc.Error(types.ErrValueDecodeFailed, err.Error(), nil)
						return true
					}
					m["value"] = val
				}

				result = append(result, m)
				return false
			})
		if errResp != nil {
			return errResp
		}
	}

	return jsonrpc.Success(result)
}
