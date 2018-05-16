package spell

import (
	net_rpc "net/rpc"
	"time"

	"github.com/ellcrys/druid/crypto"
	"github.com/ellcrys/druid/rpc"
	"github.com/ellcrys/druid/wire"
	"github.com/jinzhu/copier"
)

// ELLService provides implementation of actions
// that attempts to alter the state of the blockchain.
type ELLService struct {
	client *net_rpc.Client
	key    *crypto.Key
}

// NewELL creates a new ELL service instance
func NewELL(client *net_rpc.Client, key *crypto.Key) *ELLService {
	es := new(ELLService)
	es.client = client
	es.key = key
	return es
}

// Send sends ELL from one account to another
func (es *ELLService) Send(params map[string]interface{}) interface{} {

	if es.client == nil {
		return ConsoleErr("rpc client not initialized", nil)
	}

	tx := &wire.Transaction{
		Type:         wire.TxTypeA2A,
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
