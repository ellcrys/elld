package node

import (
	"time"

	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/testutil"
	"github.com/ellcrys/elld/wire"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ActionTransaction", func() {

	BeforeEach(func() {
		var err error
		cfg, err = testutil.SetTestCfg()
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		Expect(testutil.RemoveTestCfgDir()).To(BeNil())
	})

	Describe(".ActionAddTx", func() {

		var err error
		var n *Node

		BeforeEach(func() {
			n, err = NewNode(cfg, "127.0.0.1:40001", crypto.NewKeyFromIntSeed(1), log)
			Expect(err).To(BeNil())
		})

		AfterEach(func() {
			n.Host().Close()
		})

		It("should return err = 'field = senderPubKey, msg=invalid format: version and/or checksum bytes missing' when sender public key is invalid", func() {
			tx := &wire.Transaction{SenderPubKey: "48nCZsmoU7wvA3__invalid_fS8UhXQv4u43eny8qpT7ubdVxp3kus3eNZP"}
			err := n.ActionAddTx(tx)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("field = senderPubKey, msg=invalid format: version and/or checksum bytes missing"))
		})

		It("should return err = 'field = to, msg=recipient address is required' when recipient address is not provided", func() {
			tx := &wire.Transaction{SenderPubKey: "48nCZsmoU7wvA3ULfS8UhXQv4u43eny8qpT7ubdVxp3kus3eNZH"}
			err := n.ActionAddTx(tx)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("field = to, msg=recipient address is required"))
		})

		It("should include err = 'field = to, msg=address is not valid' when recipient address is not provided", func() {
			tx := &wire.Transaction{SenderPubKey: "48nCZsmoU7wvA3ULfS8UhXQv4u43eny8qpT7ubdVxp3kus3eNZH", To: "e5a3zJReMgLJrNn4GsYcnKf1Qa6GQFimC4"}
			err := n.ActionAddTx(tx)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("field = to, msg=address is not valid"))
		})

		It("should include err = 'field = timestamp, msg=timestamp cannot be a future time' when timestamp is a future time", func() {
			address, _ := crypto.NewKey(nil)
			tx := &wire.Transaction{
				To:           address.Addr(),
				SenderPubKey: "48nCZsmoU7wvA3ULfS8UhXQv4u43eny8qpT7ubdVxp3kus3eNZH",
				Timestamp:    time.Now().Add(10 * time.Second).Unix(),
			}
			err := n.ActionAddTx(tx)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("field = timestamp, msg=timestamp cannot be a future time"))
		})

		It("should return err = 'field = timestamp, msg=timestamp cannot over 7 days ago'", func() {
			address, _ := crypto.NewKey(nil)
			tx := &wire.Transaction{
				To:           address.Addr(),
				SenderPubKey: "48nCZsmoU7wvA3ULfS8UhXQv4u43eny8qpT7ubdVxp3kus3eNZH",
				Timestamp:    time.Now().Add(-8 * 24 * time.Hour).Unix(),
			}
			err := n.ActionAddTx(tx)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("field = timestamp, msg=timestamp cannot over 7 days ago"))
		})

		It("should return err = 'field = sig, msg=signature is required'", func() {
			address, _ := crypto.NewKey(nil)
			tx := &wire.Transaction{
				To:           address.Addr(),
				SenderPubKey: "48nCZsmoU7wvA3ULfS8UhXQv4u43eny8qpT7ubdVxp3kus3eNZH",
				Timestamp:    time.Now().Unix(),
			}
			err := n.ActionAddTx(tx)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("field = sig, msg=signature is required"))
		})

		It("should err = 'signature verification failed' when signature is not set or is invalid", func() {
			address, _ := crypto.NewKey(nil)
			sender, _ := crypto.NewKey(nil)
			tx := wire.NewTransaction(wire.TxTypeA2A, 1, address.Addr(), sender.PubKey().Base58(), "0", "0.1", time.Now().Unix())
			tx.Sig = []byte("invalid")
			err := n.ActionAddTx(tx)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("signature verification failed"))
		})

		It("should return err = 'value must be greater than zero' when tx type is wire.TxTypeA2A and has <= 0 value", func() {
			address, _ := crypto.NewKey(nil)
			sender, _ := crypto.NewKey(nil)
			tx := wire.NewTransaction(wire.TxTypeA2A, 1, address.Addr(), sender.PubKey().Base58(), "0", "0.1", time.Now().Unix())
			tx.Sig, err = wire.TxSign(tx, sender.PrivKey().Base58())
			Expect(err).To(BeNil())

			err := n.ActionAddTx(tx)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("value must be greater than zero"))
		})

		It("should return err = 'insufficient fee' when tx type is wire.TxTypeA2A and fee is less than min fee", func() {
			address, _ := crypto.NewKey(nil)
			sender, _ := crypto.NewKey(nil)
			tx := wire.NewTransaction(wire.TxTypeA2A, 1, address.Addr(), sender.PubKey().Base58(), "10", "0.0000000001", time.Now().Unix())
			tx.Sig, err = wire.TxSign(tx, sender.PrivKey().Base58())
			Expect(err).To(BeNil())

			err := n.ActionAddTx(tx)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("insufficient fee"))
		})

		It("should return err = 'unknown transaction type' when tx type is unknown", func() {
			address, _ := crypto.NewKey(nil)
			sender, _ := crypto.NewKey(nil)
			tx := wire.NewTransaction(0x300, 1, address.Addr(), sender.PubKey().Base58(), "10", "100", time.Now().Unix())
			tx.Sig, err = wire.TxSign(tx, sender.PrivKey().Base58())
			Expect(err).To(BeNil())

			err := n.ActionAddTx(tx)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("unknown transaction type"))
		})
	})
})
