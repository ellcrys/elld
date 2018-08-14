package rpc

import (
	"fmt"

	"github.com/ellcrys/elld/wire"
	"github.com/mitchellh/mapstructure"
)

// SendTxArgs represents arguments for method SendELL
type SendTxArgs struct {
	Type         int64  `json:"type"`
	Nonce        int64  `json:"nonce"`
	SenderPubKey string `json:"senderPubKey"`
	To           string `json:"to"`
	Value        string `json:"value"`
	Fee          string `json:"fee"`
	Timestamp    int64  `json:"timestamp"`
}

// SendTxPayload is used to define payload for a transaction
type SendTxPayload map[string]interface{}

// TransactionAdd adds a transaction to the transaction pool.
// The transaction is validated and verified before it is sent to the node.
func (s *Service) TransactionAdd(payload SendTxPayload, result *Result) error {

	// var err error
	var tx wire.Transaction

	if err := mapstructure.Decode(payload, &tx); err != nil {
		return fmt.Errorf("failed to decode payload: %s", err)
	}

	switch tx.Type {
	case wire.TxTypeBalance:

		apiFunc := s.engine.MustGet("TransactionAdd")
		_, err := apiFunc(&tx)
		if err != nil {
			return NewErrorResult(result, err.Error(), errCodeTransaction, 500)
		}

		return NewOKResult(result, 200, map[string]interface{}{
			"txId": string(tx.ID()),
		})

	default:
		return NewErrorResult(result, "unknown transaction type", errCodeUnknownTransactionType, 400)
	}
}
