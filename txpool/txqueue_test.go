package txpool

import (
	"time"

	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/types/core/objects"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Txqueue", func() {

	Describe(".Append", func() {
		It("should return false when capacity is reached", func() {
			tx := objects.NewTransaction(objects.TxTypeBalance, 1, "something", "pub_key", "0", "0", time.Now().Unix())
			q := newQueue(0)
			Expect(q.Append(tx)).To(BeFalse())
		})

		It("should return nil and add item", func() {
			tx := objects.NewTransaction(objects.TxTypeBalance, 1, "something", "pub_key", "0", "0", time.Now().Unix())
			q := newQueue(1)
			Expect(q.Append(tx)).To(BeTrue())
			Expect(q.container).To(HaveLen(1))
		})

		When("sorting is disabled", func() {
			It("should return transactions in the following order tx2, tx1", func() {
				tx1 := objects.NewTransaction(objects.TxTypeBalance, 1, "something", "pub_key", "0", "0.10", time.Now().Unix())
				tx2 := objects.NewTransaction(objects.TxTypeBalance, 1, "something", "pub_key", "0", "1", time.Now().Unix())
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
			tx := objects.NewTransaction(objects.TxTypeBalance, 1, "something", "pub_key", "0", "0", time.Now().Unix())
			q := newQueue(0)
			Expect(q.Prepend(tx)).To(BeFalse())
		})

		It("should return nil and add item at the head", func() {
			tx := objects.NewTransaction(objects.TxTypeBalance, 1, "something", "pub_key", "0", "0", time.Now().Unix())
			q := newQueue(2)
			Expect(q.Prepend(tx)).To(BeTrue())
			Expect(q.container).To(HaveLen(1))

			tx2 := objects.NewTransaction(objects.TxTypeBalance, 2, "something", "pub_key", "0", "0", time.Now().Unix())
			Expect(q.Prepend(tx2)).To(BeTrue())
			Expect(q.container).To(HaveLen(2))
			Expect(tx2).To(Equal(q.container[0]))
		})

		When("sorting is disabled", func() {
			It("should return transactions in the following order tx2, tx1", func() {
				tx1 := objects.NewTransaction(objects.TxTypeBalance, 1, "something", "pub_key", "0", "1", time.Now().Unix())
				tx2 := objects.NewTransaction(objects.TxTypeBalance, 1, "something", "pub_key", "0", "0.10", time.Now().Unix())
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
			tx := objects.NewTransaction(objects.TxTypeBalance, 1, "something", "pub_key", "0", "0", time.Now().Unix())
			q := newQueue(2)
			Expect(q.Prepend(tx)).To(BeTrue())
			Expect(q.Size()).To(Equal(int64(1)))
		})
	})

	Describe(".First", func() {
		It("should return nil when queue is empty", func() {
			q := newQueue(2)
			Expect(q.First()).To(BeNil())
		})

		It("should return first transaction in the queue and reduce queue size to 1", func() {
			q := newQueue(2)
			tx := objects.NewTransaction(objects.TxTypeBalance, 1, "something", "pub_key", "0", "0", time.Now().Unix())
			tx2 := objects.NewTransaction(objects.TxTypeBalance, 2, "something", "pub_key", "0", "0", time.Now().Unix())
			q.Append(tx)
			q.Append(tx2)
			Expect(q.First()).To(Equal(tx))
			Expect(q.Size()).To(Equal(int64(1)))
			Expect(q.container[0]).To(Equal(tx2))
		})

		Context("with varying transaction fee", func() {
			It("should return first transaction in the queue and reduce queue size to 1", func() {
				q := newQueue(2)
				tx := objects.NewTransaction(objects.TxTypeBalance, 1, "something", "pub_key", "0", "0.2", time.Now().Unix())
				tx2 := objects.NewTransaction(objects.TxTypeBalance, 2, "something", "pub_key", "0", "1", time.Now().Unix())
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
			q := newQueue(2)
			Expect(q.Last()).To(BeNil())
		})

		It("should return last transaction in the queue and reduce queue size to 1", func() {
			q := newQueue(2)
			tx := objects.NewTransaction(objects.TxTypeBalance, 1, "something", "pub_key", "0", "0", time.Now().Unix())
			tx2 := objects.NewTransaction(objects.TxTypeBalance, 2, "something", "pub_key", "0", "0", time.Now().Unix())
			q.Append(tx)
			q.Append(tx2)
			Expect(q.Last()).To(Equal(tx2))
			Expect(q.Size()).To(Equal(int64(1)))
			Expect(q.container[0]).To(Equal(tx))
		})

		Context("with varying transaction fee", func() {
			It("should return last transaction in the queue and reduce queue size to 1", func() {
				q := newQueue(2)
				tx := objects.NewTransaction(objects.TxTypeBalance, 1, "something", "pub_key", "0", "0.2", time.Now().Unix())
				tx2 := objects.NewTransaction(objects.TxTypeBalance, 2, "something", "pub_key", "0", "1", time.Now().Unix())
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
			q := newQueue(3)
			q.Append(objects.NewTransaction(objects.TxTypeBalance, 1, "something", "pub_key", "0", "0.2", time.Now().Unix()))
			q.Append(objects.NewTransaction(objects.TxTypeBalance, 1, "something", "pub_key", "0", "0.5", time.Now().Unix()))
			q.Append(objects.NewTransaction(objects.TxTypeBalance, 1, "something", "pub_key", "0", "1", time.Now().Unix()))
			q.Sort(SortByTxFeeAsc)
			Expect(q.container[0].GetFee().String()).To(Equal("0.2"))
			Expect(q.container[1].GetFee().String()).To(Equal("0.5"))
			Expect(q.container[2].GetFee().String()).To(Equal("1"))
		})

		It("should sort in descending order", func() {
			q := newQueue(3)
			q.Append(objects.NewTransaction(objects.TxTypeBalance, 1, "something", "pub_key", "0", "0.2", time.Now().Unix()))
			q.Append(objects.NewTransaction(objects.TxTypeBalance, 1, "something", "pub_key", "0", "0.5", time.Now().Unix()))
			q.Append(objects.NewTransaction(objects.TxTypeBalance, 1, "something", "pub_key", "0", "1", time.Now().Unix()))
			q.Sort(SortByTxFeeDesc)
			Expect(q.container[0].GetFee().String()).To(Equal("1"))
			Expect(q.container[1].GetFee().String()).To(Equal("0.5"))
			Expect(q.container[2].GetFee().String()).To(Equal("0.2"))
		})
	})

	Describe(".Has", func() {
		It("should return true when tx exist in queue", func() {
			q := newQueue(1)
			tx := objects.NewTransaction(objects.TxTypeBalance, 1, "something", "pub_key", "0", "0.2", time.Now().Unix())
			added := q.Append(tx)
			Expect(added).To(BeTrue())
			has := q.Has(tx)
			Expect(has).To(BeTrue())
		})

		It("should return false when tx does not exist in queue", func() {
			q := newQueue(1)
			tx := objects.NewTransaction(objects.TxTypeBalance, 1, "something", "pub_key", "0", "0.2", time.Now().Unix())
			has := q.Has(tx)
			Expect(has).To(BeFalse())
		})
	})

	Describe(".IFind", func() {

		var q *TxQueue
		var tx1, tx2, tx3 core.Transaction

		BeforeEach(func() {
			q = newQueue(3)
			tx1 = objects.NewTransaction(objects.TxTypeBalance, 1, "something", "pub_key", "0", "0.2", time.Now().Unix())
			tx2 = objects.NewTransaction(objects.TxTypeBalance, 2, "something", "pub_key", "0", "0.2", time.Now().Unix())
			tx3 = objects.NewTransaction(objects.TxTypeBalance, 3, "something", "pub_key", "0", "0.2", time.Now().Unix())
			q.Append(tx1)
			q.Append(tx2)
			q.Append(tx3)
		})

		It("should stop iterating when predicate returns true", func() {
			var iterated []core.Transaction
			result := q.IFind(func(tx core.Transaction) bool {
				iterated = append(iterated, tx)
				return tx.GetNonce() == 2
			})

			Describe("it should return the last item sent to the predicate", func() {
				Expect(result).To(Equal(tx2))
			})

			Describe("it should contain the first and second transaction and not the 3rd transaction", func() {
				Expect(iterated).To(HaveLen(2))
				Expect(iterated).ToNot(ContainElement(tx3))
			})
		})

		It("should return nil when predicate did not return true", func() {
			var iterated []core.Transaction
			result := q.IFind(func(tx core.Transaction) bool {
				iterated = append(iterated, tx)
				return false
			})
			Expect(result).To(BeNil())

			Describe("it should contain all transactions", func() {
				Expect(iterated).To(HaveLen(3))
			})
		})
	})
})
