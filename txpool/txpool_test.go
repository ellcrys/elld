package txpool

import (
	"time"

	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/types/core/objects"
	"github.com/ellcrys/elld/util"
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
			Expect(err).To(Equal(ErrQueueFull))
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
			Expect(tp.queue.Size()).To(Equal(int64(1)))
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
				Expect(tp.queue.Size()).To(Equal(int64(1)))
			}()
			event := <-tp.event.Once(core.EventNewTransaction)
			Expect(event.Args[0]).To(Equal(tx))
		})
	})

	Describe(".HasTxWithSameNonce", func() {

		It("should return true when a transaction with the given address and nonce exist in the pool", func() {
			tp := New(1)
			tx := objects.NewTransaction(objects.TxTypeBalance, 100, "something", util.String("abc"), "0", "0", time.Now().Unix())
			tp.Put(tx)
			result := tp.SenderHasTxWithSameNonce(tx.GetFrom(), 100)
			Expect(result).To(BeTrue())
		})

		It("should return false when a transaction with the given address and nonce does not exist in the pool", func() {
			tp := New(1)
			tx := objects.NewTransaction(objects.TxTypeBalance, 100, "something", util.String("abc"), "0", "0", time.Now().Unix())
			tp.Put(tx)
			result := tp.SenderHasTxWithSameNonce(tx.GetFrom(), 10)
			Expect(result).To(BeFalse())
		})
	})
})
