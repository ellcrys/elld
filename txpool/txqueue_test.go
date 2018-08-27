package txpool

import (
	"time"

	"github.com/ellcrys/elld/wire"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Txqueue", func() {

	Describe(".Append", func() {
		It("should return false when capacity is reached", func() {
			tx := wire.NewTransaction(wire.TxTypeBalance, 1, "something", "pub_key", "0", "0", time.Now().Unix())
			q := NewQueue(0)
			Expect(q.Append(tx)).To(BeFalse())
		})

		It("should return nil and add item", func() {
			tx := wire.NewTransaction(wire.TxTypeBalance, 1, "something", "pub_key", "0", "0", time.Now().Unix())
			q := NewQueue(1)
			Expect(q.Append(tx)).To(BeTrue())
			Expect(q.container).To(HaveLen(1))
		})

		When("sorting is disabled", func() {
			It("should return transactions in the following order tx2, tx1", func() {
				tx1 := wire.NewTransaction(wire.TxTypeBalance, 1, "something", "pub_key", "0", "0.10", time.Now().Unix())
				tx2 := wire.NewTransaction(wire.TxTypeBalance, 1, "something", "pub_key", "0", "1", time.Now().Unix())
				q := NewQueueNoSort(2)
				q.Append(tx1)
				q.Append(tx2)
				Expect(q.Size()).To(Equal(int64(2)))
				Expect(q.container[0]).To(Equal(tx1))
				Expect(q.container[1]).To(Equal(tx2))
			})
		})
	})

	Describe(".Prepend", func() {
		It("should return false when capacity is reached", func() {
			tx := wire.NewTransaction(wire.TxTypeBalance, 1, "something", "pub_key", "0", "0", time.Now().Unix())
			q := NewQueue(0)
			Expect(q.Prepend(tx)).To(BeFalse())
		})

		It("should return nil and add item at the head", func() {
			tx := wire.NewTransaction(wire.TxTypeBalance, 1, "something", "pub_key", "0", "0", time.Now().Unix())
			q := NewQueue(2)
			Expect(q.Prepend(tx)).To(BeTrue())
			Expect(q.container).To(HaveLen(1))

			tx2 := wire.NewTransaction(wire.TxTypeBalance, 2, "something", "pub_key", "0", "0", time.Now().Unix())
			Expect(q.Prepend(tx2)).To(BeTrue())
			Expect(q.container).To(HaveLen(2))
			Expect(tx2).To(Equal(q.container[0]))
		})

		When("sorting is disabled", func() {
			It("should return transactions in the following order tx2, tx1", func() {
				tx1 := wire.NewTransaction(wire.TxTypeBalance, 1, "something", "pub_key", "0", "1", time.Now().Unix())
				tx2 := wire.NewTransaction(wire.TxTypeBalance, 1, "something", "pub_key", "0", "0.10", time.Now().Unix())
				q := NewQueueNoSort(2)
				q.Prepend(tx1)
				q.Prepend(tx2)
				Expect(q.Size()).To(Equal(int64(2)))
				Expect(q.container[0]).To(Equal(tx2))
				Expect(q.container[1]).To(Equal(tx1))
			})
		})
	})

	Describe(".Size", func() {
		It("should return size = 1", func() {
			tx := wire.NewTransaction(wire.TxTypeBalance, 1, "something", "pub_key", "0", "0", time.Now().Unix())
			q := NewQueue(2)
			Expect(q.Prepend(tx)).To(BeTrue())
			Expect(q.Size()).To(Equal(int64(1)))
		})
	})

	Describe(".First", func() {
		It("should return nil when queue is empty", func() {
			q := NewQueue(2)
			Expect(q.First()).To(BeNil())
		})

		It("should return first transaction in the queue and reduce queue size to 1", func() {
			q := NewQueue(2)
			tx := wire.NewTransaction(wire.TxTypeBalance, 1, "something", "pub_key", "0", "0", time.Now().Unix())
			tx2 := wire.NewTransaction(wire.TxTypeBalance, 2, "something", "pub_key", "0", "0", time.Now().Unix())
			q.Append(tx)
			q.Append(tx2)
			Expect(q.First()).To(Equal(tx))
			Expect(q.Size()).To(Equal(int64(1)))
			Expect(q.container[0]).To(Equal(tx2))
		})

		Context("with varying transaction fee", func() {
			It("should return first transaction in the queue and reduce queue size to 1", func() {
				q := NewQueue(2)
				tx := wire.NewTransaction(wire.TxTypeBalance, 1, "something", "pub_key", "0", "0.2", time.Now().Unix())
				tx2 := wire.NewTransaction(wire.TxTypeBalance, 2, "something", "pub_key", "0", "1", time.Now().Unix())
				q.Append(tx)
				q.Append(tx2)
				Expect(q.First()).To(Equal(tx2))
				Expect(q.Size()).To(Equal(int64(1)))
				Expect(q.container[0]).To(Equal(tx))
			})
		})
	})

	Describe(".Last", func() {
		It("should return nil when queue is empty", func() {
			q := NewQueue(2)
			Expect(q.Last()).To(BeNil())
		})

		It("should return last transaction in the queue and reduce queue size to 1", func() {
			q := NewQueue(2)
			tx := wire.NewTransaction(wire.TxTypeBalance, 1, "something", "pub_key", "0", "0", time.Now().Unix())
			tx2 := wire.NewTransaction(wire.TxTypeBalance, 2, "something", "pub_key", "0", "0", time.Now().Unix())
			q.Append(tx)
			q.Append(tx2)
			Expect(q.Last()).To(Equal(tx2))
			Expect(q.Size()).To(Equal(int64(1)))
			Expect(q.container[0]).To(Equal(tx))
		})

		Context("with varying transaction fee", func() {
			It("should return last transaction in the queue and reduce queue size to 1", func() {
				q := NewQueue(2)
				tx := wire.NewTransaction(wire.TxTypeBalance, 1, "something", "pub_key", "0", "0.2", time.Now().Unix())
				tx2 := wire.NewTransaction(wire.TxTypeBalance, 2, "something", "pub_key", "0", "1", time.Now().Unix())
				q.Append(tx)
				q.Append(tx2)
				Expect(q.Last()).To(Equal(tx))
				Expect(q.Size()).To(Equal(int64(1)))
				Expect(q.container[0]).To(Equal(tx2))
			})
		})
	})

	Describe(".Sort", func() {
		It("should sort in ascending order", func() {
			q := NewQueue(3)
			q.Append(wire.NewTransaction(wire.TxTypeBalance, 1, "something", "pub_key", "0", "0.2", time.Now().Unix()))
			q.Append(wire.NewTransaction(wire.TxTypeBalance, 1, "something", "pub_key", "0", "0.5", time.Now().Unix()))
			q.Append(wire.NewTransaction(wire.TxTypeBalance, 1, "something", "pub_key", "0", "1", time.Now().Unix()))
			q.Sort(SortByTxFeeAsc)
			Expect(q.container[0].GetFee().String()).To(Equal("0.2"))
			Expect(q.container[1].GetFee().String()).To(Equal("0.5"))
			Expect(q.container[2].GetFee().String()).To(Equal("1"))
		})

		It("should sort in descending order", func() {
			q := NewQueue(3)
			q.Append(wire.NewTransaction(wire.TxTypeBalance, 1, "something", "pub_key", "0", "0.2", time.Now().Unix()))
			q.Append(wire.NewTransaction(wire.TxTypeBalance, 1, "something", "pub_key", "0", "0.5", time.Now().Unix()))
			q.Append(wire.NewTransaction(wire.TxTypeBalance, 1, "something", "pub_key", "0", "1", time.Now().Unix()))
			q.Sort(SortByTxFeeDesc)
			Expect(q.container[0].GetFee().String()).To(Equal("1"))
			Expect(q.container[1].GetFee().String()).To(Equal("0.5"))
			Expect(q.container[2].GetFee().String()).To(Equal("0.2"))
		})
	})
})
