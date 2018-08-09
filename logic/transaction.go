package logic

import (
	"github.com/ellcrys/elld/blockchain"
	"github.com/ellcrys/elld/wire"
)

// TransactionAdd adds a transaction to the transaction pool.
// Any error is sent to errCh
func (l *Logic) TransactionAdd(tx *wire.Transaction, errCh chan error) error {

	txValidator := blockchain.NewTxValidator(tx, l.engine.GetTxPool(), l.engine.GetBlockchain(), true)
	if errs := txValidator.Validate(); len(errs) > 0 {
		return sendErr(errCh, errs[0])
	}

	switch tx.Type {
	case wire.TxTypeBalance:

		// Add the transaction to the transaction pool where
		// it will be picked and broadcast to other peers
		if err := l.engine.GetTxPool().Put(tx); err != nil {
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
