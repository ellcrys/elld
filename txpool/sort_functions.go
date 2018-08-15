package txpool

import (
	"sort"

	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/wire"
)

var (

	// SortByTxFeeAsc sorts a slice of transactions by transaction fee in ascending order
	SortByTxFeeAsc = func(container []*wire.Transaction) {
		sort.Sort(ByTxFeeAsc(container))
	}

	// SortByTxFeeDesc sorts a slice of transactions by transaction fee in descending order
	SortByTxFeeDesc = func(container []*wire.Transaction) {
		sort.Sort(ByTxFeeDesc(container))
	}
)

// ByTxFeeAsc describes functions necessary to sort a slice
// of transactions by transaction fee in ascending order.
type ByTxFeeAsc []*wire.Transaction

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
	iTxFee := util.StrToDec(s[i].Fee.String())
	jTxFee := util.StrToDec(s[j].Fee.String())
	return iTxFee.Cmp(jTxFee) == -1
}

// ByTxFeeDesc describes functions necessary to sort a slice
// of transactions by transaction fee in descending order.
type ByTxFeeDesc []*wire.Transaction

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
	iTxFee := s[i].Fee.Decimal()
	jTxFee := s[j].Fee.Decimal()
	return iTxFee.Cmp(jTxFee) == 1
}
