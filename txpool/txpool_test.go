package txpool

import (
	"time"

	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/wire"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TxPool", func() {

	Describe(".Put", func() {
		It("should return err = 'capacity reached' when txpool capacity is reached", func() {
			tp := NewTxPool(0)
			a, _ := crypto.NewKey(nil)
			tx := wire.NewTransaction(wire.TxTypeBalance, 1, "something", util.String(a.PubKey().Base58()), "0", "0", time.Now().Unix())
			err := tp.Put(tx)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("capacity reached"))
		})

		It("should return err = 'exact transaction already in pool' when transaction has already been added", func() {
			tp := NewTxPool(10)
			a, _ := crypto.NewKey(nil)
			tx := wire.NewTransaction(wire.TxTypeBalance, 1, "something", util.String(a.PubKey().Base58()), "0", "0", time.Now().Unix())
			sig, _ := wire.TxSign(tx, a.PrivKey().Base58())
			tx.Sig = sig
			err := tp.Put(tx)
			Expect(err).To(BeNil())
			err = tp.Put(tx)
			Expect(err.Error()).To(Equal("exact transaction already in pool"))
		})

		It("should return err = 'unknown transaction type' when tx type is unknown", func() {
			tp := NewTxPool(1)
			a, _ := crypto.NewKey(nil)
			tx := wire.NewTransaction(10200, 1, "something", util.String(a.PubKey().Base58()), "0", "0", time.Now().Unix())
			sig, _ := wire.TxSign(tx, a.PrivKey().Base58())
			tx.Sig = sig
			err := tp.Put(tx)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("unknown transaction type"))
		})

		It("should return nil and added to queue", func() {
			tp := NewTxPool(1)
			a, _ := crypto.NewKey(nil)
			tx := wire.NewTransaction(wire.TxTypeBalance, 1, "something", util.String(a.PubKey().Base58()), "0", "0", time.Now().Unix())
			sig, _ := wire.TxSign(tx, a.PrivKey().Base58())
			tx.Sig = sig
			err := tp.Put(tx)
			Expect(err).To(BeNil())
			Expect(tp.queue.Size()).To(Equal(int64(1)))
		})

		It("should emit core.EventNewTransaction", func() {
			tp := NewTxPool(1)
			a, _ := crypto.NewKey(nil)
			tx := wire.NewTransaction(wire.TxTypeBalance, 1, "something", util.String(a.PubKey().Base58()), "0", "0", time.Now().Unix())
			go func() {
				GinkgoRecover()
				sig, _ := wire.TxSign(tx, a.PrivKey().Base58())
				tx.Sig = sig
				err := tp.Put(tx)
				Expect(err).To(BeNil())
				Expect(tp.queue.Size()).To(Equal(int64(1)))
			}()
			event := <-tp.event.Once(core.EventNewTransaction)
			Expect(event.Args[0]).To(Equal(tx))
		})
	})

	Describe(".Has", func() {
		It("should return true when transaction is not in the queue", func() {
			tp := NewTxPool(1)
			a, _ := crypto.NewKey(nil)
			tx := wire.NewTransaction(wire.TxTypeBalance, 1, "something", util.String(a.PubKey().Base58()), "0", "0", time.Now().Unix())
			sig, _ := wire.TxSign(tx, a.PrivKey().Base58())
			tx.Sig = sig
			err := tp.Put(tx)
			Expect(err).To(BeNil())
			Expect(tp.queue.Size()).To(Equal(int64(1)))

			has := tp.Has(tx)
			Expect(has).To(BeTrue())
		})

		It("should return false when transaction is not in the queue", func() {
			tp := NewTxPool(1)
			a, _ := crypto.NewKey(nil)
			tx := wire.NewTransaction(wire.TxTypeBalance, 1, "something", util.String(a.PubKey().Base58()), "0", "0", time.Now().Unix())
			sig, _ := wire.TxSign(tx, a.PrivKey().Base58())
			tx.Sig = sig
			err := tp.Put(tx)
			Expect(err).To(BeNil())
			Expect(tp.queue.Size()).To(Equal(int64(1)))

			has := tp.Has(tx)
			Expect(has).To(BeTrue())
		})
	})
})
