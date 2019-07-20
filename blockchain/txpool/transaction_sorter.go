package txpool

import (
	"sort"

	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/types/core"
)

// TransactionSorter is used to sort a container of transaction items
type TransactionSorter struct {
	container     []*ContainerItem
	ignoreFeeRate bool
	ignoreValue   bool
}

// NewSorter creates an instance of TransactionSorter
func NewSorter(container []*ContainerItem) *TransactionSorter {
	return &TransactionSorter{container: container}
}

// NewSorterFromTxs creates a TransactionSorter from a slice of transactions
func NewSorterFromTxs(txs []types.Transaction) *TransactionSorter {
	sorter := &TransactionSorter{container: []*ContainerItem{}}
	for _, tx := range txs {
		sorter.container = append(sorter.container, newItem(tx))
	}
	return sorter
}

// IgnoreFeeRate causes the transactions to not be sorted by fee rate
func (ts *TransactionSorter) IgnoreFeeRate() *TransactionSorter {
	ts.ignoreFeeRate = true
	return ts
}

// IgnoreValue causes the transactions to not be sorted by their value
func (ts *TransactionSorter) IgnoreValue() *TransactionSorter {
	ts.ignoreValue = true
	return ts
}

// Sort sorts the container of transactions
func (ts *TransactionSorter) Sort() {

	if !ts.ignoreFeeRate {
		// Sort all transactions by highest fee rate
		sort.Slice(ts.container, func(i, j int) bool {
			txI := ts.container[i]
			txJ := ts.container[j]
			return txI.FeeRate.Decimal().GreaterThan(txJ.FeeRate.Decimal())
		})
	}

	if !ts.ignoreValue {
		// Sort only ticket bid transactions by value (bid) in descending order
		sort.Slice(ts.container, func(i, j int) bool {
			txI := ts.container[i]
			txJ := ts.container[j]
			if txI.Tx.GetType() == core.TxTypeTicketBid && txJ.Tx.GetType() == core.TxTypeTicketBid {
				if txI.Tx.GetValue().Decimal().GreaterThan(txJ.Tx.GetValue().Decimal()) {
					return true
				}
			}
			return false
		})
	}

	// When transaction i & j belongs to same sender, sort by nonce in ascending order.
	sort.Slice(ts.container, func(i, j int) bool {
		txI := ts.container[i]
		txJ := ts.container[j]

		if txI.Tx.GetFrom() == txJ.Tx.GetFrom() {
			if txI.Tx.GetNonce() < txJ.Tx.GetNonce() {
				return true
			}
		}

		return false
	})
}
