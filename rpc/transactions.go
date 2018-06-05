package rpc

import (
	"github.com/ellcrys/elld/wire"
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

// SendTxPayload is used to define payload for SendELL
type SendTxPayload struct {
	Args SendTxArgs `json:"args"`
	Sig  []byte     `json:"sig"`
}

// Send adds a transaction to the transaction pool.
// The transaction is validated and verified before it is sent to the node.
func (s *Service) Send(payload SendTxPayload, result *Result) error {

	tx := wire.NewTransaction(
		payload.Args.Type,
		payload.Args.Nonce,
		payload.Args.To,
		payload.Args.SenderPubKey,
		payload.Args.Value,
		payload.Args.Fee,
		payload.Args.Timestamp,
	)

	tx.Sig = payload.Sig

	switch payload.Args.Type {
	case wire.TxTypeA2A:

		if err := s.node.ActionAddTx(tx); err != nil {
			return NewErrorResult(result, err.Error(), errCodeTransaction, 400)
		}

		return NewOKResult(result, 200, map[string]interface{}{
			"txId": string(tx.ID()),
		})

	default:
		return NewErrorResult(result, "unknown transaction type", errCodeUnknownTransactionType, 400)
	}
}
