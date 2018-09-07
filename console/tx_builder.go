package console

import (
	"time"

	"github.com/ellcrys/elld/types/core/objects"
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
	data map[string]interface{}
}

// Balance creates a balance transaction builder.
// It will attempt to fetch the address
func (o *TxBuilder) Balance(senderAddress string) *TxBalanceBuilder {
	return &TxBalanceBuilder{
		data: map[string]interface{}{
			"from":         senderAddress,
			"type":         objects.TxTypeBalance,
			"senderPubKey": "",
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
func (o *TxBalanceBuilder) Send() {
	o.data["timestamp"] = time.Now().Unix()
}

// Nonce sets the nonce
func (o *TxBalanceBuilder) Nonce(nonce int64) *TxBalanceBuilder {
	o.data["nonce"] = nonce
	return o
}

// To sets the recipient address
func (o *TxBalanceBuilder) To(address string) *TxBalanceBuilder {
	o.data["to"] = address
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
