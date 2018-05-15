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
			tx := wire.NewTransaction(wire.TxTypeRepoCreate, 1, "something", a.PubKey().Base58(), "0", "0", time.Now().Unix())
			err := tp.Put(tx)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("capacity reached"))
		})

		It("should return err = 'future time not allowed' when transaction timestamp is in the future", func() {
			tp := NewTxPool(10)
			a, _ := crypto.NewKey(nil)
			tx := wire.NewTransaction(wire.TxTypeRepoCreate, 1, "something", a.PubKey().Base58(), "0", "0", time.Now().Add(10*time.Second).Unix())
			tx.Sig, _ = wire.TxSign(tx, a.PrivKey().Base58())
			err := tp.Put(tx)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("future time not allowed"))
		})

		It("should return err = 'exact transaction already in pool' when transaction timestamp is in the future", func() {
			tp := NewTxPool(10)
			a, _ := crypto.NewKey(nil)
			tx := wire.NewTransaction(wire.TxTypeRepoCreate, 1, "something", a.PubKey().Base58(), "0", "0", time.Now().Unix())
			tx.Sig, _ = wire.TxSign(tx, a.PrivKey().Base58())
			err := tp.Put(tx)
			Expect(err).To(BeNil())
			err = tp.Put(tx)
			Expect(err.Error()).To(Equal("exact transaction already in pool"))
		})

		It("should return nil and added to queue", func() {
			tp := NewTxPool(1)
			a, _ := crypto.NewKey(nil)
			tx := wire.NewTransaction(wire.TxTypeRepoCreate, 1, "something", a.PubKey().Base58(), "0", "0", time.Now().Unix())
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
			tx := wire.NewTransaction(wire.TxTypeRepoCreate, 1, "something", a.PubKey().Base58(), "0", "0", time.Now().Unix())
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
			tx := wire.NewTransaction(wire.TxTypeRepoCreate, 1, "something", a.PubKey().Base58(), "0", "0", time.Now().Unix())
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
			tx := wire.NewTransaction(wire.TxTypeRepoCreate, 1, "something", a.PubKey().Base58(), "0", "0", time.Now().Unix())
			tx.Sig, _ = wire.TxSign(tx, a.PrivKey().Base58())
			err := tp.Put(tx)
			Expect(err).To(BeNil())
			Expect(tp.queue.Size()).To(Equal(int64(1)))

			has := tp.Has(tx)
			Expect(has).To(BeTrue())
		})
	})
})
