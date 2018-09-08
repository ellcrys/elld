package txpool

import (
	"time"

	// "github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/types/core/objects"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TxContainer", func() {

	Describe(".Add", func() {
		It("should return false when capacity is reached", func() {
			tx := objects.NewTransaction(objects.TxTypeBalance, 1, "something", "pub_key", "0", "0", time.Now().Unix())
			q := newQueue(0)
			Expect(q.Add(tx)).To(BeFalse())
		})

		It("should successfully add transaction", func() {
			tx := objects.NewTransaction(objects.TxTypeBalance, 1, "something", "pub_key", "0", "0", time.Now().Unix())
			q := newQueue(1)
			Expect(q.Add(tx)).To(BeTrue())
			Expect(q.container).To(HaveLen(1))
		})

		When("sorting is disabled", func() {
			It("should return transactions in the following order tx2, tx1", func() {
				tx1 := objects.NewTransaction(objects.TxTypeBalance, 1, "something", "pub_key", "0", "0.10", time.Now().Unix())
				tx2 := objects.NewTransaction(objects.TxTypeBalance, 1, "something", "pub_key", "0", "1", time.Now().Unix())
				q := NewQueueNoSort(2)
				q.Add(tx1)
				q.Add(tx2)
				Expect(q.Size()).To(Equal(int64(2)))
				Expect(q.container[0].Tx).To(Equal(tx1))
				Expect(q.container[1].Tx).To(Equal(tx2))
			})
		})
	})

	Describe(".Size", func() {
		It("should return size = 1", func() {
			tx := objects.NewTransaction(objects.TxTypeBalance, 1, "something", "pub_key", "0", "0", time.Now().Unix())
			q := newQueue(2)
			Expect(q.Add(tx)).To(BeTrue())
			Expect(q.Size()).To(Equal(int64(1)))
		})
	})

	Describe(".First", func() {

		It("should return nil when queue is empty", func() {
			q := newQueue(2)
			Expect(q.First()).To(BeNil())
		})

		Context("with sorting disabled", func() {
			It("should return first transaction in the queue and reduce queue size to 1", func() {
				q := NewQueueNoSort(2)
				tx := objects.NewTransaction(objects.TxTypeBalance, 1, "something", "pub_key", "0", "0", time.Now().Unix())
				tx2 := objects.NewTransaction(objects.TxTypeBalance, 2, "something", "pub_key", "0", "0", time.Now().Unix())
				q.Add(tx)
				q.Add(tx2)
				Expect(q.First()).To(Equal(tx))
				Expect(q.Size()).To(Equal(int64(1)))
				Expect(q.container[0].Tx).To(Equal(tx2))
				Expect(q.Size()).To(Equal(int64(1)))
			})
		})

		Context("with sorting enabled", func() {
			When("sender has two transactions with same nonce", func() {
				It("after sorting, the first transaction must be the one with the highest fee rate", func() {
					q := newQueue(2)
					tx := objects.NewTransaction(objects.TxTypeBalance, 1, "something", "pub_key", "0", "0.2", time.Now().Unix())
					tx.From = "sender_a"
					tx2 := objects.NewTransaction(objects.TxTypeBalance, 1, "something", "pub_key", "0", "1", time.Now().Unix())
					tx2.From = "sender_a"
					q.Add(tx)
					q.Add(tx2)
					Expect(q.container).To(HaveLen(2))
					Expect(q.First()).To(Equal(tx2))
					Expect(q.Size()).To(Equal(int64(1)))
				})
			})

			When("sender has two transaction with different nonce", func() {
				It("after sorting, the first transaction must be the one with the lowest nonce", func() {
					q := newQueue(2)
					tx := objects.NewTransaction(objects.TxTypeBalance, 1, "something", "pub_key", "0", "0.2", time.Now().Unix())
					tx.From = "sender_a"
					tx2 := objects.NewTransaction(objects.TxTypeBalance, 2, "something", "pub_key", "0", "1", time.Now().Unix())
					tx2.From = "sender_a"
					q.Add(tx)
					q.Add(tx2)
					Expect(q.container).To(HaveLen(2))
					Expect(q.First()).To(Equal(tx))
					Expect(q.Size()).To(Equal(int64(1)))
				})
			})

			When("container has 2 transactions from a sender and one from a different sender", func() {
				It("after sorting, the first transaction must be the one with the highest fee rate", func() {
					q := newQueue(3)
					tx := objects.NewTransaction(objects.TxTypeBalance, 1, "something", "pub_key", "0", "0.2", time.Now().Unix())
					tx.From = "sender_a"
					tx2 := objects.NewTransaction(objects.TxTypeBalance, 2, "something", "pub_key", "0", "1", time.Now().Unix())
					tx2.From = "sender_a"
					tx3 := objects.NewTransaction(objects.TxTypeBalance, 2, "something", "pub_key", "0", "2", time.Now().Unix())
					tx3.From = "sender_b"
					q.Add(tx)
					q.Add(tx2)
					q.Add(tx3)
					Expect(q.container).To(HaveLen(3))
					Expect(q.First()).To(Equal(tx3))
					Expect(q.Size()).To(Equal(int64(2)))
					Expect(q.container[0].Tx).To(Equal(tx))
					Expect(q.container[1].Tx).To(Equal(tx2))
				})
			})
		})
	})

	Describe(".Last", func() {
		It("should return nil when queue is empty", func() {
			q := newQueue(2)
			Expect(q.Last()).To(BeNil())
		})

		Context("with sorting disabled", func() {
			It("should return last transaction in the queue and reduce queue size to 1", func() {
				q := newQueue(2)
				tx := objects.NewTransaction(objects.TxTypeBalance, 1, "something", "pub_key", "0", "0", time.Now().Unix())
				tx2 := objects.NewTransaction(objects.TxTypeBalance, 2, "something", "pub_key", "0", "0", time.Now().Unix())
				q.Add(tx)
				q.Add(tx2)
				Expect(q.Last()).To(Equal(tx2))
				Expect(q.Size()).To(Equal(int64(1)))
			})
		})

		Context("with sorting enabled", func() {
			When("sender has two transactions with same nonce", func() {
				It("after sorting, the last transaction must be the one with the lowest fee rate", func() {
					q := newQueue(2)
					tx := objects.NewTransaction(objects.TxTypeBalance, 1, "something", "pub_key", "0", "0.2", time.Now().Unix())
					tx.From = "sender_a"
					tx2 := objects.NewTransaction(objects.TxTypeBalance, 1, "something", "pub_key", "0", "1", time.Now().Unix())
					tx2.From = "sender_a"
					q.Add(tx)
					q.Add(tx2)
					Expect(q.container).To(HaveLen(2))
					Expect(q.Last()).To(Equal(tx))
					Expect(q.Size()).To(Equal(int64(1)))
				})
			})
		})

		When("sender has two transaction with different nonce", func() {
			It("after sorting, the last transaction must be the one with the highest nonce", func() {
				q := newQueue(2)
				tx := objects.NewTransaction(objects.TxTypeBalance, 1, "something", "pub_key", "0", "0.2", time.Now().Unix())
				tx.From = "sender_a"
				tx2 := objects.NewTransaction(objects.TxTypeBalance, 2, "something", "pub_key", "0", "1", time.Now().Unix())
				tx2.From = "sender_a"
				q.Add(tx)
				q.Add(tx2)
				Expect(q.container).To(HaveLen(2))
				Expect(q.Last()).To(Equal(tx2))
				Expect(q.Size()).To(Equal(int64(1)))
			})
		})

		When("container has 2 transactions from a sender (A) and one from a different sender (B)", func() {
			It("after sorting, the last transaction must be sender (A) transaction with the highest nonce", func() {
				q := newQueue(3)
				tx := objects.NewTransaction(objects.TxTypeBalance, 1, "something", "pub_key", "0", "0.2", time.Now().Unix())
				tx.From = "sender_a"
				tx2 := objects.NewTransaction(objects.TxTypeBalance, 2, "something", "pub_key", "0", "1", time.Now().Unix())
				tx2.From = "sender_a"
				tx3 := objects.NewTransaction(objects.TxTypeBalance, 2, "something", "pub_key", "0", "2", time.Now().Unix())
				tx3.From = "sender_b"
				q.Add(tx)
				q.Add(tx2)
				q.Add(tx3)
				Expect(q.container).To(HaveLen(3))
				Expect(q.Last()).To(Equal(tx2))
				Expect(q.Size()).To(Equal(int64(2)))
				Expect(q.container[0].Tx).To(Equal(tx3))
				Expect(q.container[1].Tx).To(Equal(tx))
			})
		})
	})

	Describe(".Sort", func() {
		It("with 2 transactions by same sender; sort by nonce in ascending order", func() {
			q := newQueue(2)
			items := []*ContainerItem{
				&ContainerItem{Tx: &objects.Transaction{From: "a", Nonce: 2, Value: "10"}},
				&ContainerItem{Tx: &objects.Transaction{From: "a", Nonce: 1, Value: "10"}},
			}
			q.container = append(q.container, items...)
			q.Sort()
			Expect(q.container[0]).To(Equal(items[1]))
		})

		It("with 2 transactions by same sender; same nonce; sort by fee rate in descending order", func() {
			q := newQueue(2)
			items := []*ContainerItem{
				&ContainerItem{Tx: &objects.Transaction{From: "a", Nonce: 1, Value: "10"}, FeeRate: "0.1"},
				&ContainerItem{Tx: &objects.Transaction{From: "a", Nonce: 1, Value: "10"}, FeeRate: "0.2"},
			}
			q.container = append(q.container, items...)
			q.Sort()
			Expect(q.container[0]).To(Equal(items[1]))
		})

		Specify(`3 transactions; 
			2 by same sender and different nonce; 
			1 with highest fee rate; 
			sort by nonce (ascending) for the same sender txs;
			sort by fee rate (descending) for others`, func() {
			q := newQueue(2)
			items := []*ContainerItem{
				&ContainerItem{Tx: &objects.Transaction{From: "a", Nonce: 1, Value: "10"}, FeeRate: "0.1"},
				&ContainerItem{Tx: &objects.Transaction{From: "a", Nonce: 2, Value: "10"}, FeeRate: "0.2"},
				&ContainerItem{Tx: &objects.Transaction{From: "b", Nonce: 4, Value: "10"}, FeeRate: "1.2"},
			}
			q.container = append(q.container, items...)
			q.Sort()
			Expect(q.container[0]).To(Equal(items[2]))
			Expect(q.container[1]).To(Equal(items[0]))
			Expect(q.container[2]).To(Equal(items[1]))
		})
	})

	Describe(".Has", func() {
		It("should return true when tx exist in queue", func() {
			q := newQueue(1)
			tx := objects.NewTransaction(objects.TxTypeBalance, 1, "something", "pub_key", "0", "0.2", time.Now().Unix())
			tx.Hash = tx.ComputeHash()
			added := q.Add(tx)
			Expect(added).To(BeTrue())
			has := q.Has(tx)
			Expect(has).To(BeTrue())
		})

		It("should return false when tx does not exist in queue", func() {
			q := newQueue(1)
			tx := objects.NewTransaction(objects.TxTypeBalance, 1, "something", "pub_key", "0", "0.2", time.Now().Unix())
			tx.Hash = tx.ComputeHash()
			has := q.Has(tx)
			Expect(has).To(BeFalse())
		})
	})

	Describe(".HasByHash", func() {
		It("should return true when tx exist in queue", func() {
			q := newQueue(1)
			tx := objects.NewTransaction(objects.TxTypeBalance, 1, "something", "pub_key", "0", "0.2", time.Now().Unix())
			tx.Hash = tx.ComputeHash()
			added := q.Add(tx)
			Expect(added).To(BeTrue())
			has := q.HasByHash(tx.GetHash().HexStr())
			Expect(has).To(BeTrue())
		})

		It("should return false when tx does not exist in queue", func() {
			q := newQueue(1)
			tx := objects.NewTransaction(objects.TxTypeBalance, 1, "something", "pub_key", "0", "0.2", time.Now().Unix())
			tx.Hash = tx.ComputeHash()
			has := q.HasByHash(tx.GetHash().HexStr())
			Expect(has).To(BeFalse())
		})
	})

	Describe(".IFind", func() {

		var q *TxContainer
		var tx1, tx2, tx3 core.Transaction

		BeforeEach(func() {
			q = newQueue(3)
			tx1 = objects.NewTransaction(objects.TxTypeBalance, 1, "something", "pub_key", "0", "0.2", time.Now().Unix())
			tx2 = objects.NewTransaction(objects.TxTypeBalance, 2, "something", "pub_key", "0", "0.2", time.Now().Unix())
			tx3 = objects.NewTransaction(objects.TxTypeBalance, 3, "something", "pub_key", "0", "0.2", time.Now().Unix())
			q.Add(tx1)
			q.Add(tx2)
			q.Add(tx3)
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
