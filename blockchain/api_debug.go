package blockchain

import (
	"bytes"

	"github.com/ellcrys/elld/blockchain/common"
	"github.com/ellcrys/elld/util"

	"github.com/ellcrys/elld/elldb"
	"github.com/ellcrys/elld/rpc"
	"github.com/ellcrys/elld/rpc/jsonrpc"
	"github.com/ellcrys/elld/types"
)

// apiGetKVObjects queries raw KV objects
func (b *Blockchain) apiGetKVObjects(arg interface{}) *jsonrpc.Response {

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
			b.db.Iterate(common.MakeAccountsKey(chainID), true,
				func(kv *elldb.KVObject) bool {
					m := map[string]interface{}{
						"blockNumber": common.DecodeBlockNumber(kv.Key),
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
			b.db.Iterate(common.MakeAccountKey(blockNum, chainID, address), true,
				func(kv *elldb.KVObject) bool {
					m := map[string]interface{}{
						"blockNumber": common.DecodeBlockNumber(kv.Key),
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
			b.db.Iterate(common.MakeTxsQueryKey(chainID), true,
				func(kv *elldb.KVObject) bool {
					m := map[string]interface{}{
						"prefix":      string(kv.Prefix),
						"blockNumber": common.DecodeBlockNumber(kv.Key),
					}
					var val map[string]interface{}
					util.BytesToObject(kv.Value, &val)
					m["value"] = val
					result = append(result, m)
					return false
				})
		}

	case "chains":
		b.db.Iterate(common.MakeChainsQueryKey(), true,
			func(kv *elldb.KVObject) bool {
				var m map[string]interface{}
				util.BytesToObject(kv.Value, &m)
				result = append(result, m)
				return false
			})

	case "all":
		b.db.Iterate(nil, true,
			func(kv *elldb.KVObject) bool {
				m := map[string]interface{}{
					"prefix": string(kv.Prefix),
					"key":    string(kv.GetKey()),
				}

				if bytes.Index(kv.GetKey(), []byte(elldb.KeyPrefixSeparator)) != -1 {
					m["blockNumber"] = common.DecodeBlockNumber(kv.Key)
				}

				var val map[string]interface{}
				util.BytesToObject(kv.Value, &val)
				m["value"] = val
				result = append(result, m)
				return false
			})
	}

	return jsonrpc.Success(result)
}
