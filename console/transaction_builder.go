package console

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/btcsuite/btcutil/base58"

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

// TxTransferBuilder provides methods for building
// a balance transaction.
type TxTransferBuilder struct {
	e    *Executor
	data map[string]interface{}
}

// Balance creates a balance transaction builder.
func (o *TxBuilder) Balance() *TxTransferBuilder {

	if o.e.coinbase == nil {
		panic(o.e.vm.MakeCustomError("BuilderError", "account not loaded"))
	}

	return &TxTransferBuilder{
		e: o.e,
		data: map[string]interface{}{
			"from":         o.e.coinbase.Addr(),
			"type":         core.TxTypeBalance,
			"senderPubKey": o.e.coinbase.PubKey().Base58(),
		},
	}
}

// TicketBid creates a ticket bid transaction builder.
func (o *TxBuilder) TicketBid() *TxTransferBuilder {

	if o.e.coinbase == nil {
		panic(o.e.vm.MakeCustomError("BuilderError", "account not loaded"))
	}

	return &TxTransferBuilder{
		e: o.e,
		data: map[string]interface{}{
			"from":         o.e.coinbase.Addr(),
			"type":         core.TxTypeTicketBid,
			"senderPubKey": o.e.coinbase.PubKey().Base58(),
		},
	}
}

// Payload returns the transaction being built.
// If finalize is true, the builder attempts
// to compute the hash, sign and other fields
// before returning the transaction
func (o *TxTransferBuilder) Payload(finalize bool) map[string]interface{} {
	if finalize {
		o.Finalize()
	}
	return o.data
}

// Send signs, compute hash and signature
// and sends the payload to the transaction
// handling RPC API.
func (o *TxTransferBuilder) Send() map[string]interface{} {
	resp, err := o.send()
	if err != nil {
		panic(o.e.vm.MakeCustomError("SendError", err.Error()))
	}
	return resp
}

// Finalize returns the transaction payload
// with nonce, timestamp, hash and signature
// computed and ready for broadcast.
func (o *TxTransferBuilder) Finalize() map[string]interface{} {

	var result map[string]interface{}
	var err error

	// If the nonce has not been sent at this point
	// we must attempt to determine the current
	// nonce of the account, increment it and set it
	if o.data["nonce"] != nil {
		goto sign
	}

	result, err = o.e.callRPCMethod("state_suggestNonce", o.data["from"])
	if err != nil {
		panic(o.e.vm.MakeCustomError("BuilderError", err.Error()))
	}

	if result["error"] != nil {
		errMsg := fmt.Errorf(result["error"].(map[string]interface{})["message"].(string))
		switch errMsg.Error() {
		case "account not found":
			errMsg = fmt.Errorf("sender account not found")
		}
		panic(o.e.vm.MakeCustomError("BuilderError", errMsg.Error()))
	}

	o.data["nonce"] = int64(result["result"].(float64))

sign:
	// Set the timestamp
	o.data["timestamp"] = time.Now().Unix()

	// marshal into core.Transaction
	var tx core.Transaction
	_ = util.MapDecode(o.data, &tx)

	// Compute and set hash
	o.data["hash"] = tx.ComputeHash().HexStr()

	// Compute and set signature
	sig, err := core.TxSign(&tx, o.e.coinbase.PrivKey().Base58())
	if err != nil {
		err = fmt.Errorf("failed to sign tx: %s", err)
		panic(o.e.vm.MakeCustomError("BuilderError", err.Error()))
	}
	o.data["sig"] = util.ToHex(sig)

	return o.data
}

// Serialize returns a base58check encode equivalent of the signed payload.
func (o *TxTransferBuilder) Serialize() string {
	data := o.Finalize()
	bs, _ := json.Marshal(data)
	return base58.CheckEncode(bs, core.Base58CheckVersionTxPayload)
}

func (o *TxTransferBuilder) send() (map[string]interface{}, error) {

	data := o.Finalize()

	// Call the RPC method
	resp, err := o.e.callRPCMethod("ell_send", data)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// Nonce sets the nonce
func (o *TxTransferBuilder) Nonce(nonce int64) *TxTransferBuilder {
	o.data["nonce"] = nonce
	return o
}

// To sets the recipient's address
func (o *TxTransferBuilder) To(address string) *TxTransferBuilder {
	o.data["to"] = address
	return o
}

// From sets the sender's address
func (o *TxTransferBuilder) From(address string) *TxTransferBuilder {
	o.data["from"] = address
	return o
}

// SenderPubKey sets the senders public key
func (o *TxTransferBuilder) SenderPubKey(pk string) *TxTransferBuilder {
	o.data["senderPubKey"] = pk
	return o
}

// Value sets the amount to be sent
func (o *TxTransferBuilder) Value(amount string) *TxTransferBuilder {
	o.data["value"] = amount
	return o
}

// Fee sets the fee
func (o *TxTransferBuilder) Fee(amount string) *TxTransferBuilder {
	o.data["fee"] = amount
	return o
}

// Reset the builder
func (o *TxTransferBuilder) Reset() {
	o.data = make(map[string]interface{})
}
