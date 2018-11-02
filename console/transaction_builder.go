package console

import (
	"fmt"
	"time"

	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/util"
)

// TxBuilder provides methods for building
// and executing a transaction
type TxBuilder struct {
	e *Executor
}

// NewTxBuilder creates a TxBuilder
func NewTxBuilder(e *Executor) *TxBuilder {
	return &TxBuilder{
		e: e,
	}
}

// TxBalanceBuilder provides methods for building
// a balance transaction.
type TxBalanceBuilder struct {
	e    *Executor
	data map[string]interface{}
}

// Balance creates a balance transaction builder.
// It will attempt to fetch the address
func (o *TxBuilder) Balance() *TxBalanceBuilder {

	if o.e.coinbase == nil {
		panic(o.e.vm.MakeCustomError("BuilderError", "account not loaded"))
	}

	return &TxBalanceBuilder{
		e: o.e,
		data: map[string]interface{}{
			"from":         o.e.coinbase.Addr(),
			"type":         core.TxTypeBalance,
			"senderPubKey": o.e.coinbase.PubKey().Base58(),
		},
	}
}

// Payload returns the builder's payload
func (o *TxBalanceBuilder) Payload() map[string]interface{} {
	return o.data
}

// Send signs, compute hash and signature
// and sends the payload to the transaction
// handling RPC API.
func (o *TxBalanceBuilder) Send() map[string]interface{} {
	resp, err := o.send()
	if err != nil {
		panic(o.e.vm.MakeCustomError("SendError", err.Error()))
	}
	return resp
}

func (o *TxBalanceBuilder) send() (map[string]interface{}, error) {

	var result map[string]interface{}
	var err error

	// If the nonce has not been sent at this point
	// we must attempt to determine the current
	// nonce of the account, increment it and set it
	if o.data["nonce"] != nil {
		goto send
	}

	result, err = o.e.callRPCMethod("state_getAccountNonce", o.data["from"])
	if err != nil {
		return nil, err
	}

	if result["error"] != nil {
		errMsg := fmt.Errorf(result["error"].(map[string]interface{})["message"].(string))
		switch errMsg.Error() {
		case "account not found":
			errMsg = fmt.Errorf("sender account not found")
		}
		return nil, errMsg
	}

	o.data["nonce"] = int64(result["result"].(float64)) + 1

send:
	// Set the timestamp
	o.data["timestamp"] = time.Now().Unix()

	// marshal into core.Transaction
	var tx core.Transaction
	util.MapDecode(o.data, &tx)

	// Compute and set hash
	o.data["hash"] = tx.ComputeHash()

	// Compute and set signature
	sig, err := core.TxSign(&tx, o.e.coinbase.PrivKey().Base58())
	if err != nil {
		return nil, fmt.Errorf("failed to sign tx: %s", err)
	}
	o.data["sig"] = sig

	// Call the RPC method
	resp, err := o.e.callRPCMethod("ell_send", o.data)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// Nonce sets the nonce
func (o *TxBalanceBuilder) Nonce(nonce int64) *TxBalanceBuilder {
	o.data["nonce"] = nonce
	return o
}

// To sets the recipient's address
func (o *TxBalanceBuilder) To(address string) *TxBalanceBuilder {
	o.data["to"] = address
	return o
}

// From sets the sender's address
func (o *TxBalanceBuilder) From(address string) *TxBalanceBuilder {
	o.data["from"] = address
	return o
}

// Type sets the transaction type
func (o *TxBalanceBuilder) Type(txType int) *TxBalanceBuilder {
	o.data["type"] = txType
	return o
}

// SenderPubKey sets the senders public key
func (o *TxBalanceBuilder) SenderPubKey(pk int) *TxBalanceBuilder {
	o.data["senderPubKey"] = pk
	return o
}

// Value sets the amount to be sent
func (o *TxBalanceBuilder) Value(amount string) *TxBalanceBuilder {
	o.data["value"] = amount
	return o
}

// Fee sets the fee
func (o *TxBalanceBuilder) Fee(amount string) *TxBalanceBuilder {
	o.data["fee"] = amount
	return o
}

// Reset the builder
func (o *TxBalanceBuilder) Reset() {
	o.data = make(map[string]interface{})
}
