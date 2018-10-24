package common

import (
	"fmt"

	"github.com/syndtr/goleveldb/leveldb"

	"github.com/ellcrys/elld/elldb"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/util"
)

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
func GetTxOp(db elldb.TxCreator, opts ...core.CallOp) *TxOp {
	for _, op := range opts {
		switch _op := op.(type) {
		case *TxOp:
			if _op.Tx != nil {
				return _op
			}
		}
	}

	txOp := &TxOp{
		CanFinish: true,
	}

	// Create new transaction and wrap
	// it within a TxOp. Set the TxOp
	// to finished, if db is closed.
	tx, err := db.NewTx()
	if err != nil {
		if err != leveldb.ErrClosed {
			panic(fmt.Errorf("failed to create transaction: %s", err))
		}
		txOp.finished = true
	}
	txOp.Tx = tx

	return txOp
}

// GetBlockQueryRangeOp is a convenience method to get QueryBlockRange
// option from a slice of CallOps
func GetBlockQueryRangeOp(opts ...core.CallOp) *BlockQueryRange {
	for _, op := range opts {
		switch _op := op.(type) {
		case *BlockQueryRange:
			return _op
		}
	}
	return &BlockQueryRange{}
}

// GetTransitions finds a Transitions option from a given
// slice of call options and returns a slice of transition objects
func GetTransitions(opts ...core.CallOp) (transitions []Transition) {
	for _, op := range opts {
		switch _op := op.(type) {
		case *TransitionsOp:
			for _, t := range *_op {
				transitions = append(transitions, t)
			}
			return
		}
	}
	return []Transition{}
}

// GetChainerOp is a convenience method to get ChainerOp
// option from a slice of CallOps
func GetChainerOp(opts ...core.CallOp) *ChainerOp {
	for _, op := range opts {
		switch _op := op.(type) {
		case *ChainerOp:
			return _op
		}
	}
	return &ChainerOp{}
}

// ComputeTxsRoot computes the merkle root of a set of transactions.
func ComputeTxsRoot(txs []core.Transaction) util.Hash {

	tree := NewTree()
	for _, tx := range txs {
		tree.Add(TreeItem(tx.GetHash().Bytes()))
	}

	tree.Build()
	return tree.Root()
}
