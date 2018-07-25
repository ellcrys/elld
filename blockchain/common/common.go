package common

import "github.com/ellcrys/elld/database"

// DocType represents a document type
type DocType int

const (
	// TypeBlock represents a block document type
	TypeBlock DocType = 0x1

	// TypeTx represents a transaction document type
	TypeTx DocType = 0x2
)

// GetTxOp checks and return a transaction added in the supplied call
// option slice. If none is found, a new transaction is created and
// returned as a TxOp.
func GetTxOp(db database.TxCreator, opts ...CallOp) TxOp {
	if len(opts) > 0 {
		for _, op := range opts {
			switch _op := op.(type) {
			case TxOp:
				return _op
			}
		}
	}
	tx, err := db.NewTx()
	if err != nil {
		panic("failed to create transaction")
	}
	return TxOp{
		Tx:        tx,
		CanFinish: true,
	}
}
