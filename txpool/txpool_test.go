package txpool

import (
	"time"

	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/types/core/objects"
	"github.com/ellcrys/elld/util"
	"github.com/olebedev/emitter"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TxPool", func() {

	Describe(".Put", func() {
		It("should return err = 'capacity reached' when txpool capacity is reached", func() {
			tp := New(0)
			a, _ := crypto.NewKey(nil)
			tx := objects.NewTransaction(objects.TxTypeBalance, 1, "something", util.String(a.PubKey().Base58()), "0", "0", time.Now().Unix())
			err := tp.Put(tx)
			Expect(err).ToNot(BeNil())
			Expect(err).To(Equal(ErrContainerFull))
		})

		It("should return err = 'exact transaction already in the pool' when transaction has already been added", func() {
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

		It("should return err = 'unknown transaction type' when tx type is unknown", func() {
			tp := New(1)
			a, _ := crypto.NewKey(nil)
			tx := objects.NewTransaction(10200, 1, "something", util.String(a.PubKey().Base58()), "0", "0", time.Now().Unix())
			sig, _ := objects.TxSign(tx, a.PrivKey().Base58())
			tx.Sig = sig
			err := tp.Put(tx)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("unknown transaction type"))
		})

		It("should return nil and added to queue", func() {
			tp := New(1)
			a, _ := crypto.NewKey(nil)
			tx := objects.NewTransaction(objects.TxTypeBalance, 1, "something", util.String(a.PubKey().Base58()), "0", "0", time.Now().Unix())
			sig, _ := objects.TxSign(tx, a.PrivKey().Base58())
			tx.Sig = sig
			err := tp.Put(tx)
			Expect(err).To(BeNil())
			Expect(tp.container.Size()).To(Equal(int64(1)))
		})

		It("should emit core.EventNewTransaction", func() {
			tp := New(1)
			a, _ := crypto.NewKey(nil)
			tx := objects.NewTransaction(objects.TxTypeBalance, 1, "something", util.String(a.PubKey().Base58()), "0", "0", time.Now().Unix())
			go func() {
				GinkgoRecover()
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

	Describe(".HasTxWithSameNonce", func() {

		var tp *TxPool

		BeforeEach(func() {
			tp = New(1)
		})

		It("should return true when a transaction with the given address and nonce exist in the pool", func() {
			tx := objects.NewTransaction(objects.TxTypeBalance, 100, "something", util.String("abc"), "0", "0", time.Now().Unix())
			tp.Put(tx)
			result := tp.SenderHasTxWithSameNonce(tx.GetFrom(), 100)
			Expect(result).To(BeTrue())
		})

		It("should return false when a transaction with the given address and nonce does not exist in the pool", func() {
			tx := objects.NewTransaction(objects.TxTypeBalance, 100, "something", util.String("abc"), "0", "0", time.Now().Unix())
			tp.Put(tx)
			result := tp.SenderHasTxWithSameNonce(tx.GetFrom(), 10)
			Expect(result).To(BeFalse())
		})
	})

	Describe(".Has", func() {

		var tp *TxPool

		BeforeEach(func() {
			tp = New(1)
		})

		It("should return true when tx exist", func() {
			tx := objects.NewTransaction(objects.TxTypeBalance, 100, "something", util.String("abc"), "0", "0", time.Now().Unix())
			tp.Put(tx)
			Expect(tp.Has(tx)).To(BeTrue())
		})

		It("should return false when tx does not exist", func() {
			tx := objects.NewTransaction(objects.TxTypeBalance, 100, "something", util.String("abc"), "0", "0", time.Now().Unix())
			Expect(tp.Has(tx)).To(BeFalse())
		})
	})

	Describe(".Size", func() {

		var tp *TxPool

		BeforeEach(func() {
			tp = New(1)
			Expect(tp.Size()).To(Equal(int64(0)))
		})

		It("should return 1", func() {
			tx := objects.NewTransaction(objects.TxTypeBalance, 100, "something", util.String("abc"), "0", "0", time.Now().Unix())
			tp.Put(tx)
			Expect(tp.Size()).To(Equal(int64(1)))
		})
	})

	Describe(".ByteSize", func() {

		var tx, tx2 core.Transaction
		var tp *TxPool

		BeforeEach(func() {
			tp = New(2)
		})

		BeforeEach(func() {
			tx = objects.NewTransaction(objects.TxTypeBalance, 100, "something", util.String("abc"), "0", "0", time.Now().Unix())
			tx.SetHash(util.StrToHash("hash1"))
			tx2 = objects.NewTransaction(objects.TxTypeBalance, 100, "something_2", util.String("xyz"), "0", "0", time.Now().Unix())
			tx2.SetHash(util.StrToHash("hash2"))
			tp.Put(tx)
			tp.Put(tx2)
		})

		It("should return expected byte size", func() {
			s := tp.ByteSize()
			Expect(s).To(Equal(tx.SizeNoFee() + tx2.SizeNoFee()))
		})

		When("a transaction is removed", func() {

			var curByteSize int64

			BeforeEach(func() {
				curByteSize = tp.ByteSize()
				Expect(curByteSize).To(Equal(tx.SizeNoFee() + tx2.SizeNoFee()))
			})

			It("should reduce the byte size when First is called", func() {
				rmTx := tp.container.First()
				s := tp.ByteSize()
				Expect(s).To(Equal(curByteSize - rmTx.SizeNoFee()))
			})

			It("should reduce the byte size when Last is called", func() {
				rmTx := tp.container.Last()
				s := tp.ByteSize()
				Expect(s).To(Equal(curByteSize - rmTx.SizeNoFee()))
			})
		})
	})

	Describe(".Select", func() {

		var tp *TxPool
		var tx, tx2, tx3 *objects.Transaction

		BeforeEach(func() {
			tp = New(100)
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

		It("should only include transactions up to the given max size", func() {
			maxSize := tx.SizeNoFee() + tx2.SizeNoFee()
			txs := tp.Select(maxSize)
			Expect(txs).To(HaveLen(2))
		})

		It("should only include all transactions when max size exceeds pool size", func() {
			maxSize := tx.SizeNoFee() + tx2.SizeNoFee() + tx3.SizeNoFee() + 100
			txs := tp.Select(maxSize)
			Expect(txs).To(HaveLen(3))
		})

		When("max size is too small", func() {
			It("should select nothing and put back all transactions back in the pool", func() {
				maxSize := int64(1)
				txs := tp.Select(maxSize)
				Expect(txs).To(HaveLen(0))
				Expect(tp.container.container).To(HaveLen(3))
			})
		})
	})

	Describe(".handleNewBlockEvent", func() {

		var tp *TxPool
		var ee *emitter.Emitter
		var tx, tx2, tx3 *objects.Transaction

		BeforeEach(func() {
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

		It("should remove the transactions included in the block", func() {
			go tp.handleNewBlockEvent()
			block := &objects.Block{
				Transactions: []*objects.Transaction{tx2, tx3},
			}
			<-ee.Emit(core.EventNewBlock, block)
			Expect(tp.Size()).To(Equal(int64(1)))
			Expect(tp.container.container[0].Tx).To(Equal(tx))
		})
	})
})
