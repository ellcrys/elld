package txpool

import (
	"sort"

	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/util"
)

var (

	// SortByTxFeeAsc sorts a slice of transactions by transaction fee in ascending order
	SortByTxFeeAsc = func(container []core.Transaction) {
		sort.Sort(ByTxFeeAsc(container))
	}

	// SortByTxFeeDesc sorts a slice of transactions by transaction fee in descending order
	SortByTxFeeDesc = func(container []core.Transaction) {
		sort.Sort(ByTxFeeDesc(container))
	}
)

// ByTxFeeAsc describes functions necessary to sort a slice
// of transactions by transaction fee in ascending order.
type ByTxFeeAsc []core.Transaction

// Len returns the length of the slice
func (s ByTxFeeAsc) Len() int {
	return len(s)
}

// Swap witches the indexes of two items
func (s ByTxFeeAsc) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less checks the smallest value between two txs
func (s ByTxFeeAsc) Less(i, j int) bool {
	iTxFee := util.StrToDec(s[i].GetFee().String())
	jTxFee := util.StrToDec(s[j].GetFee().String())
	return iTxFee.Cmp(jTxFee) == -1
}

// ByTxFeeDesc describes functions necessary to sort a slice
// of transactions by transaction fee in descending order.
type ByTxFeeDesc []core.Transaction

// Len returns the length of the slice
func (s ByTxFeeDesc) Len() int {
	return len(s)
}

// Swap witches the indexes of two items
func (s ByTxFeeDesc) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less checks the smallest value between two txs
func (s ByTxFeeDesc) Less(i, j int) bool {
	iTxFee := s[i].GetFee().Decimal()
	jTxFee := s[j].GetFee().Decimal()
	return iTxFee.Cmp(jTxFee) == 1
}
