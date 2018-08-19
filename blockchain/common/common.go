package common

import (
	"bytes"
	"sort"

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
	if len(opts) > 0 {
		for _, op := range opts {
			switch _op := op.(type) {
			case *TxOp:
				if _op.Tx != nil {
					return _op
				}
			}
		}
	}
	tx, err := db.NewTx()
	if err != nil {
		panic("failed to create transaction")
	}
	return &TxOp{
		Tx:        tx,
		CanFinish: true,
	}
}

// ComputeTxsRoot computes the merkle root of a set of transactions.
// Transactions are first lexicographically sorted and added to a
// brand new tree. Returns the tree root.
func ComputeTxsRoot(txs []core.Transaction) util.Hash {

	sort.Slice(txs, func(i, j int) bool {
		return bytes.Compare(txs[i].GetHash().Bytes(), txs[j].GetHash().Bytes()) == -1
	})

	tree := NewTree()
	for _, tx := range txs {
		tree.Add(TreeItem(tx.GetHash().Bytes()))
	}

	tree.Build()
	return tree.Root()
}
