package txpool

import (
	"time"

	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/types/core"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TxContainer", func() {

	Describe(".Add", func() {
		It("should return false when capacity is reached", func() {
			tx := core.NewTxObj(core.TxTypeBalance, 1, "something", "pub_key", "0", "0", time.Now().Unix())
			q := newTxContainer(0)
			Expect(q.Add(tx)).To(BeFalse())
		})

		It("should successfully add transaction", func() {
			tx := core.NewTxObj(core.TxTypeBalance, 1, "something", "pub_key", "0", "0", time.Now().Unix())
			q := newTxContainer(1)
			Expect(q.Add(tx)).To(BeTrue())
			Expect(q.container).To(HaveLen(1))
		})

		When("sorting is disabled", func() {
			It("should return transactions in the following order tx2, tx1", func() {
				tx1 := core.NewTxObj(core.TxTypeBalance, 1, "something", "pub_key", "0", "0.10", time.Now().Unix())
				tx2 := core.NewTxObj(core.TxTypeBalance, 1, "something", "pub_key", "0", "1", time.Now().Unix())
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
			tx := core.NewTxObj(core.TxTypeBalance, 1, "something", "pub_key", "0", "0", time.Now().Unix())
			q := newTxContainer(2)
			Expect(q.Add(tx)).To(BeTrue())
			Expect(q.Size()).To(Equal(int64(1)))
		})
	})

	Describe(".First", func() {

		It("should return nil when queue is empty", func() {
			q := newTxContainer(2)
			Expect(q.First()).To(BeNil())
		})

		Context("with sorting disabled", func() {
			It("should return first transaction in the queue and reduce queue size to 1", func() {
				q := NewQueueNoSort(2)
				tx := core.NewTxObj(core.TxTypeBalance, 1, "something", "pub_key", "0", "0", time.Now().Unix())
				tx2 := core.NewTxObj(core.TxTypeBalance, 2, "something", "pub_key", "0", "0", time.Now().Unix())
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
					q := newTxContainer(2)
					tx := core.NewTxObj(core.TxTypeBalance, 1, "something", "pub_key", "0", "0.2", time.Now().Unix())
					tx.From = "sender_a"
					tx2 := core.NewTxObj(core.TxTypeBalance, 1, "something", "pub_key", "0", "1", time.Now().Unix())
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
					q := newTxContainer(2)
					tx := core.NewTxObj(core.TxTypeBalance, 1, "something", "pub_key", "0", "0.2", time.Now().Unix())
					tx.From = "sender_a"
					tx2 := core.NewTxObj(core.TxTypeBalance, 2, "something", "pub_key", "0", "1", time.Now().Unix())
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
					q := newTxContainer(3)
					tx := core.NewTxObj(core.TxTypeBalance, 1, "something", "pub_key", "0", "0.2", time.Now().Unix())
					tx.From = "sender_a"
					tx2 := core.NewTxObj(core.TxTypeBalance, 2, "something", "pub_key", "0", "1", time.Now().Unix())
					tx2.From = "sender_a"
					tx3 := core.NewTxObj(core.TxTypeBalance, 2, "something", "pub_key", "0", "2", time.Now().Unix())
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

			When("container has 2 tx bids with same fee and same value", func() {
				var tx, tx2 *core.Transaction
				var q *TxContainer

				BeforeEach(func() {
					q = newTxContainer(2)
					tx = core.NewTxObj(core.TxTypeTicketBid, 1, "something", "pub_key", "10", "0.1", 1562165124)
					tx.From = "sender_a"
					q.Add(tx)
					tx2 = core.NewTxObj(core.TxTypeTicketBid, 1, "something", "pub_key", "10", "0.1", 1562165124)
					tx2.From = "sender_b"
					q.Add(tx2)
				})

				Specify("that there is not change in the insertion order", func() {
					Expect(q.container).To(HaveLen(2))
					Expect(q.First()).To(Equal(tx))
					Expect(q.First()).To(Equal(tx2))
				})
			})

			When("container has 2 tx bids; both with same value but 1 with a higher fee", func() {
				var tx, tx2 *core.Transaction
				var q *TxContainer

				BeforeEach(func() {
					q = newTxContainer(2)
					tx = core.NewTxObj(core.TxTypeTicketBid, 1, "something", "pub_key", "10", "0.1", 1562165124)
					tx.From = "sender_a"
					q.Add(tx)
					tx2 = core.NewTxObj(core.TxTypeTicketBid, 1, "something", "pub_key", "10", "0.2", 1562165124)
					tx2.From = "sender_b"
					q.Add(tx2)
				})

				Specify("that the tx with the higher fee is at index 0 (head)", func() {
					Expect(q.container).To(HaveLen(2))
					Expect(q.First()).To(Equal(tx2))
				})
			})

			When("container has 2 tx bids; both with same fee but 1 with a higher value", func() {
				var tx, tx2 *core.Transaction
				var q *TxContainer

				BeforeEach(func() {
					q = newTxContainer(2)
					tx = core.NewTxObj(core.TxTypeTicketBid, 1, "something", "pub_key", "10", "0.1", 1562165124)
					tx.From = "sender_a"
					q.Add(tx)
					tx2 = core.NewTxObj(core.TxTypeTicketBid, 1, "something", "pub_key", "20", "0.1", 1562165124)
					tx2.From = "sender_b"
					q.Add(tx2)
				})

				Specify("that the tx with the higher value is at index 0 (head)", func() {
					Expect(q.container).To(HaveLen(2))
					Expect(q.First()).To(Equal(tx2))
					Expect(q.First()).To(Equal(tx))
				})
			})

			When("container has 2 tx bids; one with a lower value but higher fee", func() {
				var tx, tx2 *core.Transaction
				var q *TxContainer

				BeforeEach(func() {
					q = newTxContainer(2)
					tx = core.NewTxObj(core.TxTypeTicketBid, 1, "something", "pub_key", "10", "1000", 1562165124)
					tx.From = "sender_a"
					q.Add(tx)
					tx2 = core.NewTxObj(core.TxTypeTicketBid, 1, "something", "pub_key", "20", "0.1", 1562165124)
					tx2.From = "sender_b"
					q.Add(tx2)
				})

				Specify("that the tx with the lower fee is at index 0 (head)", func() {
					Expect(q.container).To(HaveLen(2))
					Expect(q.First()).To(Equal(tx2))
					Expect(q.First()).To(Equal(tx))
				})
			})

			When("container has 2 tx bids; both with same fee but 1 with a higher value", func() {
				var tx, tx2 *core.Transaction
				var q *TxContainer

				BeforeEach(func() {
					q = newTxContainer(2)
					tx = core.NewTxObj(core.TxTypeTicketBid, 2, "something", "pub_key", "10", "0.1", 1562165124)
					tx.From = "sender_a"
					q.Add(tx)
					tx2 = core.NewTxObj(core.TxTypeTicketBid, 1, "something", "pub_key", "20", "0.1", 1562165124)
					tx2.From = "sender_a"
					q.Add(tx2)
				})

				Specify("that the tx with the higher value is at index 0 (head)", func() {
					Expect(q.container).To(HaveLen(2))
					Expect(q.First()).To(Equal(tx2))
					Expect(q.First()).To(Equal(tx))
				})
			})
		})
	})

	Describe(".Last", func() {
		It("should return nil when queue is empty", func() {
			q := newTxContainer(2)
			Expect(q.Last()).To(BeNil())
		})

		Context("with sorting disabled", func() {
			It("should return last transaction in the queue and reduce queue size to 1", func() {
				q := newTxContainer(2)
				tx := core.NewTxObj(core.TxTypeBalance, 1, "something", "pub_key", "0", "0", time.Now().Unix())
				tx2 := core.NewTxObj(core.TxTypeBalance, 2, "something", "pub_key", "0", "0", time.Now().Unix())
				q.Add(tx)
				q.Add(tx2)
				Expect(q.Last()).To(Equal(tx2))
				Expect(q.Size()).To(Equal(int64(1)))
			})
		})

		Context("with sorting enabled", func() {
			When("sender has two transactions with same nonce", func() {
				It("after sorting, the last transaction must be the one with the lowest fee rate", func() {
					q := newTxContainer(2)
					tx := core.NewTxObj(core.TxTypeBalance, 1, "something", "pub_key", "0", "0.2", time.Now().Unix())
					tx.From = "sender_a"
					tx2 := core.NewTxObj(core.TxTypeBalance, 1, "something", "pub_key", "0", "1", time.Now().Unix())
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
				q := newTxContainer(2)
				tx := core.NewTxObj(core.TxTypeBalance, 1, "something", "pub_key", "0", "0.2", time.Now().Unix())
				tx.From = "sender_a"
				tx2 := core.NewTxObj(core.TxTypeBalance, 2, "something", "pub_key", "0", "1", time.Now().Unix())
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
				q := newTxContainer(3)
				tx := core.NewTxObj(core.TxTypeBalance, 1, "something", "pub_key", "0", "0.2", time.Now().Unix())
				tx.From = "sender_a"
				tx2 := core.NewTxObj(core.TxTypeBalance, 2, "something", "pub_key", "0", "1", time.Now().Unix())
				tx2.From = "sender_a"
				tx3 := core.NewTxObj(core.TxTypeBalance, 2, "something", "pub_key", "0", "2", time.Now().Unix())
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
		Context("2 transactions by same sender; sort by nonce in ascending order", func() {
			Specify("that item at index 1 with lowest nonce is the first", func() {
				q := newTxContainer(2)
				items := []*ContainerItem{
					{Tx: &core.Transaction{From: "a", Nonce: 2, Value: "10"}, FeeRate: "0"},
					{Tx: &core.Transaction{From: "a", Nonce: 1, Value: "10"}, FeeRate: "0"},
				}
				q.container = append(q.container, items...)
				q.Sort()
				Expect(q.container[0]).To(Equal(items[1]))
			})
		})

		Context("2 transactions by same sender; same nonce; sort by fee rate in descending order", func() {
			Specify("that the item at index 1 with the highest fee rate is the first", func() {
				q := newTxContainer(2)
				items := []*ContainerItem{
					{Tx: &core.Transaction{From: "a", Nonce: 1, Value: "10"}, FeeRate: "0.1"},
					{Tx: &core.Transaction{From: "a", Nonce: 1, Value: "10"}, FeeRate: "0.2"},
				}
				q.container = append(q.container, items...)
				q.Sort()
				Expect(q.container[0]).To(Equal(items[1]))
			})
		})

		Context("2 transactions by same sender; same nonce, with fee rate sorting disabled", func() {
			It("should remain unsorted/the same", func() {
				q := newTxContainer(2)
				q.sorter.IgnoreFeeRate()
				items := []*ContainerItem{
					{Tx: &core.Transaction{From: "a", Nonce: 1, Value: "10"}, FeeRate: "0.1"},
					{Tx: &core.Transaction{From: "a", Nonce: 1, Value: "10"}, FeeRate: "0.2"},
				}
				q.container = append(q.container, items...)
				q.Sort()
				Expect(q.container[0]).To(Equal(items[0]))
				Expect(q.container[1]).To(Equal(items[1]))
			})
		})

		Context("2 transactions; same fee rate; with value sorting disabled", func() {
			It("should remain unsorted/the same", func() {
				q := newTxContainer(2)
				items := []*ContainerItem{
					{Tx: &core.Transaction{Type: core.TxTypeTicketBid, From: "a", Nonce: 1, Value: "10"}, FeeRate: "0.1"},
					{Tx: &core.Transaction{Type: core.TxTypeTicketBid, From: "b", Nonce: 2, Value: "20"}, FeeRate: "0.1"},
				}
				q.container = append(q.container, items...)
				q.sorter.IgnoreValue()
				q.Sort()
				Expect(q.container[0]).To(Equal(items[0]))
				Expect(q.container[1]).To(Equal(items[1]))
			})
		})

		Specify(`3 transactions; 
				2 by same sender and different nonce; 
				1 with highest fee rate; 
				sort by nonce (ascending) for the same sender txs;
				sort by fee rate (descending) for others`, func() {
			q := newTxContainer(2)
			items := []*ContainerItem{
				{Tx: &core.Transaction{From: "a", Nonce: 1, Value: "10"}, FeeRate: "0.1"},
				{Tx: &core.Transaction{From: "a", Nonce: 2, Value: "10"}, FeeRate: "0.2"},
				{Tx: &core.Transaction{From: "b", Nonce: 4, Value: "10"}, FeeRate: "1.2"},
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
			q := newTxContainer(1)
			tx := core.NewTxObj(core.TxTypeBalance, 1, "something", "pub_key", "0", "0.2", time.Now().Unix())
			tx.Hash = tx.ComputeHash()
			added := q.Add(tx)
			Expect(added).To(BeTrue())
			has := q.Has(tx)
			Expect(has).To(BeTrue())
		})

		It("should return false when tx does not exist in queue", func() {
			q := newTxContainer(1)
			tx := core.NewTxObj(core.TxTypeBalance, 1, "something", "pub_key", "0", "0.2", time.Now().Unix())
			tx.Hash = tx.ComputeHash()
			has := q.Has(tx)
			Expect(has).To(BeFalse())
		})
	})

	Describe(".HasByHash", func() {
		It("should return true when tx exist in queue", func() {
			q := newTxContainer(1)
			tx := core.NewTxObj(core.TxTypeBalance, 1, "something", "pub_key", "0", "0.2", time.Now().Unix())
			tx.Hash = tx.ComputeHash()
			added := q.Add(tx)
			Expect(added).To(BeTrue())
			has := q.HasByHash(tx.GetHash().HexStr())
			Expect(has).To(BeTrue())
		})

		It("should return false when tx does not exist in queue", func() {
			q := newTxContainer(1)
			tx := core.NewTxObj(core.TxTypeBalance, 1, "something", "pub_key", "0", "0.2", time.Now().Unix())
			tx.Hash = tx.ComputeHash()
			has := q.HasByHash(tx.GetHash().HexStr())
			Expect(has).To(BeFalse())
		})
	})

	Describe(".remove", func() {

		var q *TxContainer
		var tx, tx2, tx3, tx4 *core.Transaction

		BeforeEach(func() {
			q = newTxContainer(4)
			tx = core.NewTxObj(core.TxTypeBalance, 1, "something", "pub_key", "0", "0.2", time.Now().Unix())
			tx.Hash = tx.ComputeHash()
			q.Add(tx)
			tx2 = core.NewTxObj(core.TxTypeBalance, 1, "something2", "pub_key", "0", "0.2", time.Now().Unix())
			tx2.Hash = tx2.ComputeHash()
			q.Add(tx2)
			tx3 = core.NewTxObj(core.TxTypeBalance, 1, "something2", "pub_key", "0", "0.2", time.Now().Unix())
			tx3.Hash = tx3.ComputeHash()
			q.Add(tx3)
			tx4 = core.NewTxObj(core.TxTypeBalance, 1, "something2", "pub_key", "0", "0.4", time.Now().Unix())
			tx4.Hash = tx4.ComputeHash()
			q.Add(tx4)
		})

		It("should do nothing when transaction does not exist in the container", func() {
			unknownTx := core.NewTxObj(core.TxTypeBalance, 1, "unknown", "pub_key", "0", "0.2", time.Now().Unix())
			unknownTx.Hash = unknownTx.ComputeHash()
			q.Remove(unknownTx)
			Expect(q.Size()).To(Equal(int64(4)))
		})

		It("should remove transactions", func() {
			q.Remove(tx2, tx3)
			Expect(q.Size()).To(Equal(int64(2)))
			Expect(q.container[0].Tx).To(Equal(tx4))
			Expect(q.container[1].Tx).To(Equal(tx))
			Expect(q.len).To(Equal(int64(2)))
			Expect(q.byteSize).To(Equal(int64(tx.GetSizeNoFee() + tx4.GetSizeNoFee())))
		})
	})

	Describe(".IFind", func() {

		var q *TxContainer
		var tx1, tx2, tx3 types.Transaction

		BeforeEach(func() {
			q = newTxContainer(3)
			tx1 = core.NewTxObj(core.TxTypeBalance, 1, "something", "pub_key", "0", "0.2", time.Now().Unix())
			tx2 = core.NewTxObj(core.TxTypeBalance, 2, "something", "pub_key", "0", "0.2", time.Now().Unix())
			tx3 = core.NewTxObj(core.TxTypeBalance, 3, "something", "pub_key", "0", "0.2", time.Now().Unix())
			q.Add(tx1)
			q.Add(tx2)
			q.Add(tx3)
		})

		It("should stop iterating when predicate returns true", func() {
			var iterated []types.Transaction
			result := q.IFind(func(tx types.Transaction) bool {
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
			var iterated []types.Transaction
			result := q.IFind(func(tx types.Transaction) bool {
				iterated = append(iterated, tx)
				return false
			})
			Expect(result).To(BeNil())

			Describe("it should contain all transactions", func() {
				Expect(iterated).To(HaveLen(3))
			})
		})
	})

	Describe(".Get", func() {
		It("should return Not nil when tx exist in queue", func() {
			q := newTxContainer(1)
			tx := core.NewTxObj(core.TxTypeBalance, 1, "something", "pub_key", "0", "0.2", time.Now().Unix())
			tx.Hash = tx.ComputeHash()
			added := q.Add(tx)
			Expect(added).To(BeTrue())
			txData := q.GetByHash(tx.Hash.HexStr())
			Expect(txData).ToNot(BeNil())
		})

		It("should return nil when tx does not exist in queue", func() {
			q := newTxContainer(1)
			tx := core.NewTxObj(core.TxTypeBalance, 1, "something", "pub_key", "0", "0.2", time.Now().Unix())
			tx.Hash = tx.ComputeHash()
			txData := q.GetByHash(tx.Hash.HexStr())
			Expect(txData).To(BeNil())
		})

	})

})
