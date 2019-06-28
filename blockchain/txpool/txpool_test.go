package txpool

import (
	"time"

	"github.com/ellcrys/mother/crypto"
	"github.com/ellcrys/mother/params"
	"github.com/ellcrys/mother/types"
	"github.com/ellcrys/mother/types/core"

	"github.com/ellcrys/mother/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TxPool", func() {

	Describe(".Put", func() {
		It("should return err = 'capacity reached' when txpool capacity is reached", func() {
			tp := New(0)
			a, _ := crypto.NewKey(nil)
			tx := core.NewTransaction(core.TxTypeBalance, 1, "something", util.String(a.PubKey().Base58()), "0", "0", time.Now().Unix())
			err := tp.Put(tx)
			Expect(err).ToNot(BeNil())
			Expect(err).To(Equal(ErrContainerFull))
		})

		It("should return err = 'exact transaction already in the pool' when transaction has already been added", func() {
			tp := New(10)
			a, _ := crypto.NewKey(nil)
			tx := core.NewTransaction(core.TxTypeBalance, 1, "something", util.String(a.PubKey().Base58()), "0", "0", time.Now().Unix())
			sig, _ := core.TxSign(tx, a.PrivKey().Base58())
			tx.Sig = sig
			err := tp.Put(tx)
			Expect(err).To(BeNil())
			err = tp.Put(tx)
			Expect(err).To(Equal(ErrTxAlreadyAdded))
		})

		It("should return err = 'unknown transaction type' when tx type is unknown", func() {
			tp := New(1)
			a, _ := crypto.NewKey(nil)
			tx := core.NewTransaction(10200, 1, "something", util.String(a.PubKey().Base58()), "0", "0", time.Now().Unix())
			sig, _ := core.TxSign(tx, a.PrivKey().Base58())
			tx.Sig = sig
			err := tp.Put(tx)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("unknown transaction type"))
		})

		It("should return nil and added to queue", func() {
			tp := New(1)
			a, _ := crypto.NewKey(nil)
			tx := core.NewTransaction(core.TxTypeBalance, 1, "something", util.String(a.PubKey().Base58()), "0", "0", time.Now().Unix())
			sig, _ := core.TxSign(tx, a.PrivKey().Base58())
			tx.Sig = sig
			err := tp.Put(tx)
			Expect(err).To(BeNil())
			Expect(tp.container.Size()).To(Equal(int64(1)))
		})
	})

	Describe(".Has", func() {

		var tp *TxPool

		BeforeEach(func() {
			tp = New(1)
		})

		It("should return true when tx exist", func() {
			tx := core.NewTransaction(core.TxTypeBalance, 100, "something", util.String("abc"), "0", "0", time.Now().Unix())
			tp.Put(tx)
			Expect(tp.Has(tx)).To(BeTrue())
		})

		It("should return false when tx does not exist", func() {
			tx := core.NewTransaction(core.TxTypeBalance, 100, "something", util.String("abc"), "0", "0", time.Now().Unix())
			Expect(tp.Has(tx)).To(BeFalse())
		})
	})

	Describe(".GetByFrom", func() {

		var tp *TxPool
		var key1 = crypto.NewKeyFromIntSeed(1)
		var key2 = crypto.NewKeyFromIntSeed(2)
		var tx, tx2, tx3 types.Transaction

		BeforeEach(func() {
			tp = New(3)
			tx = core.NewTx(core.TxTypeBalance, 1, "a", key1, "12.2", "0.2", time.Now().Unix())
			tx2 = core.NewTx(core.TxTypeBalance, 2, "a", key1, "12.3", "0.2", time.Now().Unix())
			tx3 = core.NewTx(core.TxTypeBalance, 2, "a", key2, "12.3", "0.2", time.Now().Unix())
			_ = tp.addTx(tx)
			_ = tp.addTx(tx2)
			_ = tp.addTx(tx3)
			Expect(tp.Size()).To(Equal(int64(3)))
		})

		It("should return two transactions matching key1", func() {
			txs := tp.GetByFrom(key1.Addr())
			Expect(txs).To(HaveLen(2))
			Expect(txs[0]).To(Equal(tx))
			Expect(txs[1]).To(Equal(tx2))
		})
	})

	Describe(".Size", func() {

		var tp *TxPool

		BeforeEach(func() {
			tp = New(1)
			Expect(tp.Size()).To(Equal(int64(0)))
		})

		It("should return 1", func() {
			tx := core.NewTransaction(core.TxTypeBalance, 100, "something", util.String("abc"), "0", "0", time.Now().Unix())
			tp.Put(tx)
			Expect(tp.Size()).To(Equal(int64(1)))
		})
	})

	Describe(".ByteSize", func() {

		var tx, tx2 types.Transaction
		var tp *TxPool

		BeforeEach(func() {
			tp = New(2)
		})

		BeforeEach(func() {
			tx = core.NewTransaction(core.TxTypeBalance, 100, "something", util.String("abc"), "0", "0", time.Now().Unix())
			tx.SetHash(util.StrToHash("hash1"))
			tx2 = core.NewTransaction(core.TxTypeBalance, 100, "something_2", util.String("xyz"), "0", "0", time.Now().Unix())
			tx2.SetHash(util.StrToHash("hash2"))
			tp.Put(tx)
			tp.Put(tx2)
		})

		It("should return expected byte size", func() {
			s := tp.ByteSize()
			Expect(s).To(Equal(tx.GetSizeNoFee() + tx2.GetSizeNoFee()))
		})

		When("a transaction is removed", func() {

			var curByteSize int64

			BeforeEach(func() {
				curByteSize = tp.ByteSize()
				Expect(curByteSize).To(Equal(tx.GetSizeNoFee() + tx2.GetSizeNoFee()))
			})

			It("should reduce the byte size when First is called", func() {
				rmTx := tp.container.First()
				s := tp.ByteSize()
				Expect(s).To(Equal(curByteSize - rmTx.GetSizeNoFee()))
			})

			It("should reduce the byte size when Last is called", func() {
				rmTx := tp.container.Last()
				s := tp.ByteSize()
				Expect(s).To(Equal(curByteSize - rmTx.GetSizeNoFee()))
			})
		})
	})

	Describe(".clean", func() {

		var tx, tx2 types.Transaction
		var tp *TxPool

		Context("when TxTTL is 1 day", func() {

			BeforeEach(func() {
				params.TxTTL = 1
				tp = New(2)

				tx = core.NewTransaction(core.TxTypeBalance, 100, "something", util.String("abc"), "0", "0", time.Now().Unix())
				tx.SetHash(util.StrToHash("hash1"))
				tx.SetTimestamp(time.Now().UTC().AddDate(0, 0, -2).Unix())

				tx2 = core.NewTransaction(core.TxTypeBalance, 101, "something2", util.String("abc"), "0", "0", time.Now().Unix())
				tx2.SetHash(util.StrToHash("hash2"))
				tx2.SetTimestamp(time.Now().Unix())

				tp.container.Add(tx)
				tp.container.Add(tx2)
				Expect(tp.Size()).To(Equal(int64(2)))
			})

			It("should remove expired transaction", func() {
				tp.clean()
				Expect(tp.Size()).To(Equal(int64(1)))
				Expect(tp.Has(tx2)).To(BeTrue())
				Expect(tp.Has(tx)).To(BeFalse())
			})
		})
	})

	Describe(".Remove", func() {

		var tp *TxPool
		var tx, tx2, tx3 types.Transaction

		BeforeEach(func() {
			tp = New(100)

			tx = core.NewTransaction(core.TxTypeBalance, 100, "something", util.String("abc"), "0", "0", time.Now().Unix())
			tx.SetHash(tx.ComputeHash())
			tp.Put(tx)

			tx2 = core.NewTransaction(core.TxTypeBalance, 100, "something2", util.String("abc2"), "0", "0", time.Now().Unix())
			tx2.SetHash(tx2.ComputeHash())
			tp.Put(tx2)

			tx3 = core.NewTransaction(core.TxTypeBalance, 100, "something3", util.String("abc3"), "0", "0", time.Now().Unix())
			tx3.SetHash(tx3.ComputeHash())
			tp.Put(tx3)
		})

		It("should remove the transactions included in the block", func() {
			txs := []types.Transaction{tx2, tx3}
			tp.Remove(txs...)
			Expect(tp.Size()).To(Equal(int64(1)))
			Expect(tp.container.container[0].Tx).To(Equal(tx))
		})
	})

	Describe(".GetByHash", func() {

		var tp *TxPool
		var tx, tx2 types.Transaction

		BeforeEach(func() {
			tp = New(100)

			tx = core.NewTransaction(core.TxTypeBalance, 100, "something", util.String("abc"), "0", "0", time.Now().Unix())
			tx.SetHash(tx.ComputeHash())
			tp.Put(tx)

			tx2 = core.NewTransaction(core.TxTypeBalance, 100, "something2", util.String("abc2"), "0", "0", time.Now().Unix())
			tx2.SetHash(tx2.ComputeHash())
		})

		It("It should not be equal", func() {
			Expect(tx).ToNot(Equal(tx2))
		})

		It("should get transaction from pool", func() {
			txData := tp.GetByHash(tx.GetHash().HexStr())
			Expect(txData).ToNot(BeNil())
			Expect(txData).To(Equal(tx))
		})

		It("should return nil from  GetTransaction in pool", func() {
			txData := tp.GetByHash(tx2.GetHash().HexStr())
			Expect(txData).To(BeNil())
		})

	})

})
