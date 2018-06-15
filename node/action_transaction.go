package node

import (
	"github.com/ellcrys/elld/constants"
	"github.com/ellcrys/elld/wire"
	"github.com/shopspring/decimal"
)

// ActionAddTx adds a transaction to the transaction pool.
// It will:
// - Validate the transaction.
// - Verify the transaction's signature
// - Check and reject zero value
// - Reject transaction below minimum transaction fee
func (n *Node) ActionAddTx(tx *wire.Transaction) error {

	if errs := wire.TxValidate(tx); len(errs) > 0 {
		return errs[0]
	}

	if err := wire.TxVerify(tx); err != nil {
		return wire.ErrTxVerificationFailed
	}

	switch tx.Type {
	case wire.TxTypeBalance:

		value, _ := decimal.NewFromString(tx.Value)
		if value.LessThanOrEqual(decimal.NewFromFloat(0)) {
			return wire.ErrTxLowValue
		}

		fee, _ := decimal.NewFromString(tx.Fee)
		if fee.Cmp(constants.A2AMinimumTxFee) == -1 {
			return wire.ErrTxInsufficientFee
		}

		if err := n.protoc.GetUnSignedTxPool().Put(tx); err != nil {
			return err
		}

		n.protoc.AddTxSession(tx.ID())

		return nil

	default:
		return wire.ErrTxTypeUnknown
	}
}
