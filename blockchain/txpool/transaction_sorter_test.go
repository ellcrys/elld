package txpool

import (
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/types/core"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TransactionSorter", func() {
	Describe(".NewSorterFromTxs", func() {
		var txs = []types.Transaction{
			&core.Transaction{From: "a", Nonce: 2, Value: "10"},
			&core.Transaction{From: "a", Nonce: 1, Value: "10"},
		}

		It("should create a sorter", func() {
			sorter := NewSorterFromTxs(txs)
			Expect(sorter.container).To(HaveLen(2))
		})
	})
})
