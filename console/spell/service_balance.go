package spell

import (
	net_rpc "net/rpc"
	"time"

	"github.com/fatih/structs"

	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/rpc"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/wire"
)

// BalanceService provides functionalities for
// sending ELL from a balance account to another balance account.,
// checking amount of ELL in a balance account and more.
type BalanceService struct {
	client *net_rpc.Client
	key    *crypto.Key
}

// NewBalanceService creates a new ELL service instance
func NewBalanceService(client *net_rpc.Client, key *crypto.Key) *BalanceService {
	es := new(BalanceService)
	es.client = client
	es.key = key
	return es
}

// Send sends ELL from one account to another
func (es *BalanceService) Send(params map[string]interface{}) interface{} {

	if es.client == nil {
		return ConsoleErr("rpc client not initialized", nil)
	}
	if es.key == nil {
		return ConsoleErr("key not set", nil)
	}

	tx := &wire.Transaction{
		Type:         wire.TxTypeBalance,
		Nonce:        1, // TODO: fetch current nonce
		SenderPubKey: util.String(es.key.PubKey().Base58()),
		To:           util.String(params["to"].(string)),
		Value:        util.String(params["value"].(string)),
		Fee:          "1", // TODO: if params["fee"] is null, use as default
		Timestamp:    time.Now().Unix(),
	}

	sig, err := wire.TxSign(tx, es.key.PrivKey().Base58())
	if err != nil {
		return ConsoleErr(err.Error(), nil)
	}

	tx.Sig = sig
	tx.Hash = tx.ComputeHash()

	var result rpc.Result
	err = es.client.Call("Service.TransactionAdd", structs.Map(tx), &result)
	if err != nil {
		return ConsoleErr(err.Error(), nil)
	}

	if result.Status != 200 {
		return ConsoleErr(result.Error, map[string]interface{}{
			"code": result.ErrCode,
		})
	}

	return result.Data
}
