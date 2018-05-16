package txpool

import (
	"time"

	"github.com/ellcrys/druid/crypto"
	"github.com/ellcrys/druid/wire"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TxPool", func() {

	Describe(".Put", func() {
		It("should return err = 'capacity reached' when txpool capacity is reached", func() {
			tp := NewTxPool(0)
			a, _ := crypto.NewKey(nil)
			tx := wire.NewTransaction(wire.TxTypeA2A, 1, "something", a.PubKey().Base58(), "0", "0", time.Now().Unix())
			err := tp.Put(tx)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("capacity reached"))
		})

		It("should return err = 'exact transaction already in pool' when transaction has already been added", func() {
			tp := NewTxPool(10)
			a, _ := crypto.NewKey(nil)
			tx := wire.NewTransaction(wire.TxTypeA2A, 1, "something", a.PubKey().Base58(), "0", "0", time.Now().Unix())
			tx.Sig, _ = wire.TxSign(tx, a.PrivKey().Base58())
			err := tp.Put(tx)
			Expect(err).To(BeNil())
			err = tp.Put(tx)
			Expect(err.Error()).To(Equal("exact transaction already in pool"))
		})

		It("should return err = 'unknown transaction type' when tx type is unknown", func() {
			tp := NewTxPool(1)
			a, _ := crypto.NewKey(nil)
			tx := wire.NewTransaction(10200, 1, "something", a.PubKey().Base58(), "0", "0", time.Now().Unix())
			tx.Sig, _ = wire.TxSign(tx, a.PrivKey().Base58())
			err := tp.Put(tx)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("unknown transaction type"))
		})

		It("should return nil and added to queue", func() {
			tp := NewTxPool(1)
			a, _ := crypto.NewKey(nil)
			tx := wire.NewTransaction(wire.TxTypeA2A, 1, "something", a.PubKey().Base58(), "0", "0", time.Now().Unix())
			tx.Sig, _ = wire.TxSign(tx, a.PrivKey().Base58())
			err := tp.Put(tx)
			Expect(err).To(BeNil())
			Expect(tp.queue.Size()).To(Equal(int64(1)))
		})

		It("should return nil and call onQueueCB function ", func() {
			onQueueFuncCalled := false
			tp := NewTxPool(1)
			tp.OnQueued(func(tx *wire.Transaction) error {
				onQueueFuncCalled = true
				return nil
			})
			a, _ := crypto.NewKey(nil)
			tx := wire.NewTransaction(wire.TxTypeA2A, 1, "something", a.PubKey().Base58(), "0", "0", time.Now().Unix())
			tx.Sig, _ = wire.TxSign(tx, a.PrivKey().Base58())
			err := tp.Put(tx)
			Expect(err).To(BeNil())
			Expect(tp.queue.Size()).To(Equal(int64(1)))
			Expect(onQueueFuncCalled).To(BeTrue())
		})
	})

	Describe(".Has", func() {
		It("should return true when transaction is not in the queue", func() {
			tp := NewTxPool(1)
			a, _ := crypto.NewKey(nil)
			tx := wire.NewTransaction(wire.TxTypeA2A, 1, "something", a.PubKey().Base58(), "0", "0", time.Now().Unix())
			tx.Sig, _ = wire.TxSign(tx, a.PrivKey().Base58())
			err := tp.Put(tx)
			Expect(err).To(BeNil())
			Expect(tp.queue.Size()).To(Equal(int64(1)))

			has := tp.Has(tx)
			Expect(has).To(BeTrue())
		})

		It("should return false when transaction is not in the queue", func() {
			tp := NewTxPool(1)
			a, _ := crypto.NewKey(nil)
			tx := wire.NewTransaction(wire.TxTypeA2A, 1, "something", a.PubKey().Base58(), "0", "0", time.Now().Unix())
			tx.Sig, _ = wire.TxSign(tx, a.PrivKey().Base58())
			err := tp.Put(tx)
			Expect(err).To(BeNil())
			Expect(tp.queue.Size()).To(Equal(int64(1)))

			has := tp.Has(tx)
			Expect(has).To(BeTrue())
		})
	})
})
