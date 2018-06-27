package logic

import (
	"github.com/ellcrys/elld/constants"
	"github.com/ellcrys/elld/wire"
	"github.com/shopspring/decimal"
)

// TransactionAdd adds a transaction to the transaction pool.
// Any error is sent to errCh
func (l *Logic) TransactionAdd(tx *wire.Transaction, errCh chan error) error {

	// Validate the transaction's field
	if err := tx.Validate(); err != nil {
		return sendErr(errCh, err)
	}

	// Verify that the transaction's signature is valid
	if err := wire.TxVerify(tx); err != nil {
		return sendErr(errCh, wire.ErrTxVerificationFailed)
	}

	switch tx.Type {
	case wire.TxTypeBalance:
		// For balance transaction, ensure the value is not set to
		// zero or a non-numeric value.
		value, _ := decimal.NewFromString(tx.Value)
		if value.LessThanOrEqual(decimal.NewFromFloat(0)) {
			return sendErr(errCh, wire.ErrTxLowValue)
		}

		// Do not allow a transaction with fee below the minimum
		// network transaction fee.
		fee, _ := decimal.NewFromString(tx.Fee)
		if fee.Cmp(constants.A2AMinimumTxFee) == -1 {
			return sendErr(errCh, wire.ErrTxInsufficientFee)
		}

		// Add the transaction to the unsigned transaction pool where
		// it will be picked and broadcast to other peers
		if err := l.engine.GetUnSignedTxPool().Put(tx); err != nil {
			return sendErr(errCh, err)
		}

		// Create a transaction session so this node will wait
		// for the signed equivalent of the transaction from endorsers.
		l.engine.AddTxSession(tx.ID())

		return sendErr(errCh, nil)

	default:
		return sendErr(errCh, wire.ErrTxTypeUnknown)
	}
}
