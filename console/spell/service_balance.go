package spell

import (
	net_rpc "net/rpc"
	"time"

	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/rpc"
	"github.com/ellcrys/elld/wire"
	"github.com/jinzhu/copier"
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
		SenderPubKey: es.key.PubKey().Base58(),
		To:           params["to"].(string),
		Value:        params["value"].(string),
		Fee:          "1", // TODO: if params["fee"] is null, use a default
		Timestamp:    time.Now().Unix(),
	}

	sig, err := wire.TxSign(tx, es.key.PrivKey().Base58())
	if err != nil {
		return ConsoleErr(err.Error(), nil)
	}

	var sendTxArgs rpc.SendTxArgs
	copier.Copy(&sendTxArgs, tx)

	var result rpc.Result
	var payload = rpc.SendTxPayload{
		Args: sendTxArgs,
		Sig:  sig,
	}

	err = es.client.Call("Service.Send", payload, &result)
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
