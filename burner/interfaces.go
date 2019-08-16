package burner

import (
	"github.com/ellcrys/ltcd/btcjson"
	"github.com/ellcrys/ltcd/chaincfg/chainhash"
	"github.com/ellcrys/ltcd/wire"
)

// RPCClient represents an RPC client used to
// access the burner chain RPC service
type RPCClient interface {
	GetBlockHash(blockHeight int64) (*chainhash.Hash, error)
	GetBlockHeader(blockHash *chainhash.Hash) (*wire.BlockHeader, error)
	GetBestBlock() (*chainhash.Hash, int32, error)
	GetBlockVerboseTx(blockHash *chainhash.Hash) (*btcjson.GetBlockVerboseResult, error)
}
