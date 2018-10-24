package txpool

import (
	"testing"
	"time"

	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/params"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/types/core/objects"
	"github.com/ellcrys/elld/util"
	. "github.com/ncodes/goblin"
	"github.com/olebedev/emitter"
	. "github.com/onsi/gomega"
)

func TestTxPool(t *testing.T) {
	g := Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("TxPool", func() {

		g.Describe(".Put", func() {
			g.It("should return err = 'capacity reached' when txpool capacity is reached", func() {
				tp := New(0)
				a, _ := crypto.NewKey(nil)
				tx := objects.NewTransaction(objects.TxTypeBalance, 1, "something", util.String(a.PubKey().Base58()), "0", "0", time.Now().Unix())
				err := tp.Put(tx)
				Expect(err).ToNot(BeNil())
				Expect(err).To(Equal(ErrContainerFull))
			})

			g.It("should return err = 'exact transaction already in the pool' when transaction has already been added", func() {
				tp := New(10)
				a, _ := crypto.NewKey(nil)
				tx := objects.NewTransaction(objects.TxTypeBalance, 1, "something", util.String(a.PubKey().Base58()), "0", "0", time.Now().Unix())
				sig, _ := objects.TxSign(tx, a.PrivKey().Base58())
				tx.Sig = sig
				err := tp.Put(tx)
				Expect(err).To(BeNil())
				err = tp.Put(tx)
				Expect(err).To(Equal(ErrTxAlreadyAdded))
			})

			g.It("should return err = 'unknown transaction type' when tx type is unknown", func() {
				tp := New(1)
				a, _ := crypto.NewKey(nil)
				tx := objects.NewTransaction(10200, 1, "something", util.String(a.PubKey().Base58()), "0", "0", time.Now().Unix())
				sig, _ := objects.TxSign(tx, a.PrivKey().Base58())
				tx.Sig = sig
				err := tp.Put(tx)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("unknown transaction type"))
			})

			g.It("should return nil and added to queue", func() {
				tp := New(1)
				a, _ := crypto.NewKey(nil)
				tx := objects.NewTransaction(objects.TxTypeBalance, 1, "something", util.String(a.PubKey().Base58()), "0", "0", time.Now().Unix())
				sig, _ := objects.TxSign(tx, a.PrivKey().Base58())
				tx.Sig = sig
				err := tp.Put(tx)
				Expect(err).To(BeNil())
				Expect(tp.container.Size()).To(Equal(int64(1)))
			})

			g.It("should emit core.EventNewTransaction", func() {
				tp := New(1)
				a, _ := crypto.NewKey(nil)
				tx := objects.NewTransaction(objects.TxTypeBalance, 1, "something", util.String(a.PubKey().Base58()), "0", "0", time.Now().Unix())
				go func() {
					sig, _ := objects.TxSign(tx, a.PrivKey().Base58())
					tx.Sig = sig
					err := tp.Put(tx)
					Expect(err).To(BeNil())
					Expect(tp.container.Size()).To(Equal(int64(1)))
				}()
				event := <-tp.event.Once(core.EventNewTransaction)
				Expect(event.Args[0]).To(Equal(tx))
			})
		})

		g.Describe(".HasTxWithSameNonce", func() {

			var tp *TxPool

			g.BeforeEach(func() {
				tp = New(1)
			})

			g.It("should return true when a transaction with the given address and nonce exist in the pool", func() {
				tx := objects.NewTransaction(objects.TxTypeBalance, 100, "something", util.String("abc"), "0", "0", time.Now().Unix())
				tp.Put(tx)
				result := tp.SenderHasTxWithSameNonce(tx.GetFrom(), 100)
				Expect(result).To(BeTrue())
			})

			g.It("should return false when a transaction with the given address and nonce does not exist in the pool", func() {
				tx := objects.NewTransaction(objects.TxTypeBalance, 100, "something", util.String("abc"), "0", "0", time.Now().Unix())
				tp.Put(tx)
				result := tp.SenderHasTxWithSameNonce(tx.GetFrom(), 10)
				Expect(result).To(BeFalse())
			})
		})

		g.Describe(".Has", func() {

			var tp *TxPool

			g.BeforeEach(func() {
				tp = New(1)
			})

			g.It("should return true when tx exist", func() {
				tx := objects.NewTransaction(objects.TxTypeBalance, 100, "something", util.String("abc"), "0", "0", time.Now().Unix())
				tp.Put(tx)
				Expect(tp.Has(tx)).To(BeTrue())
			})

			g.It("should return false when tx does not exist", func() {
				tx := objects.NewTransaction(objects.TxTypeBalance, 100, "something", util.String("abc"), "0", "0", time.Now().Unix())
				Expect(tp.Has(tx)).To(BeFalse())
			})
		})

		g.Describe(".Size", func() {

			var tp *TxPool

			g.BeforeEach(func() {
				tp = New(1)
				Expect(tp.Size()).To(Equal(int64(0)))
			})

			g.It("should return 1", func() {
				tx := objects.NewTransaction(objects.TxTypeBalance, 100, "something", util.String("abc"), "0", "0", time.Now().Unix())
				tp.Put(tx)
				Expect(tp.Size()).To(Equal(int64(1)))
			})
		})

		g.Describe(".ByteSize", func() {

			var tx, tx2 core.Transaction
			var tp *TxPool

			g.BeforeEach(func() {
				tp = New(2)
			})

			g.BeforeEach(func() {
				tx = objects.NewTransaction(objects.TxTypeBalance, 100, "something", util.String("abc"), "0", "0", time.Now().Unix())
				tx.SetHash(util.StrToHash("hash1"))
				tx2 = objects.NewTransaction(objects.TxTypeBalance, 100, "something_2", util.String("xyz"), "0", "0", time.Now().Unix())
				tx2.SetHash(util.StrToHash("hash2"))
				tp.Put(tx)
				tp.Put(tx2)
			})

			g.It("should return expected byte size", func() {
				s := tp.ByteSize()
				Expect(s).To(Equal(tx.GetSizeNoFee() + tx2.GetSizeNoFee()))
			})

			g.When("a transaction is removed", func() {

				var curByteSize int64

				g.BeforeEach(func() {
					curByteSize = tp.ByteSize()
					Expect(curByteSize).To(Equal(tx.GetSizeNoFee() + tx2.GetSizeNoFee()))
				})

				g.It("should reduce the byte size when First is called", func() {
					rmTx := tp.container.First()
					s := tp.ByteSize()
					Expect(s).To(Equal(curByteSize - rmTx.GetSizeNoFee()))
				})

				g.It("should reduce the byte size when Last is called", func() {
					rmTx := tp.container.Last()
					s := tp.ByteSize()
					Expect(s).To(Equal(curByteSize - rmTx.GetSizeNoFee()))
				})
			})
		})

		g.Describe(".clean", func() {

			var tx, tx2 core.Transaction
			var tp *TxPool

			g.Context("when TxTTL is 1 day", func() {

				g.BeforeEach(func() {
					params.TxTTL = 1
					tp = New(2)

					tx = objects.NewTransaction(objects.TxTypeBalance, 100, "something", util.String("abc"), "0", "0", time.Now().Unix())
					tx.SetHash(util.StrToHash("hash1"))
					tx.SetTimestamp(time.Now().UTC().AddDate(0, 0, -2).Unix())

					tx2 = objects.NewTransaction(objects.TxTypeBalance, 101, "something2", util.String("abc"), "0", "0", time.Now().Unix())
					tx2.SetHash(util.StrToHash("hash2"))
					tx2.SetTimestamp(time.Now().Unix())

					tp.container.Add(tx)
					tp.container.Add(tx2)
					Expect(tp.Size()).To(Equal(int64(2)))
				})

				g.It("should remove expired transaction", func() {
					tp.clean()
					Expect(tp.Size()).To(Equal(int64(1)))
					Expect(tp.Has(tx2)).To(BeTrue())
					Expect(tp.Has(tx)).To(BeFalse())
				})
			})
		})

		g.Describe(".removeTransactionsInBlock", func() {

			var tp *TxPool
			var ee *emitter.Emitter
			var tx, tx2, tx3 *objects.Transaction

			g.BeforeEach(func() {
				tp = New(100)
				ee = &emitter.Emitter{}
				tp.SetEventEmitter(ee)

				tx = objects.NewTransaction(objects.TxTypeBalance, 100, "something", util.String("abc"), "0", "0", time.Now().Unix())
				tx.Hash = tx.ComputeHash()
				tp.Put(tx)

				tx2 = objects.NewTransaction(objects.TxTypeBalance, 100, "something2", util.String("abc2"), "0", "0", time.Now().Unix())
				tx2.Hash = tx2.ComputeHash()
				tp.Put(tx2)

				tx3 = objects.NewTransaction(objects.TxTypeBalance, 100, "something3", util.String("abc3"), "0", "0", time.Now().Unix())
				tx3.Hash = tx3.ComputeHash()
				tp.Put(tx3)
			})

			g.It("should remove the transactions included in the block", func() {
				txs := []core.Transaction{tx2, tx3}
				tp.remove(txs...)
				Expect(tp.Size()).To(Equal(int64(1)))
				Expect(tp.container.container[0].Tx).To(Equal(tx))
			})
		})
	})
}
