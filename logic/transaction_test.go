package logic

import (
	"time"

	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/node"
	"github.com/ellcrys/elld/testutil"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/wire"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Transaction", func() {

	BeforeEach(func() {
		var err error
		cfg, err = testutil.SetTestCfg()
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		Expect(testutil.RemoveTestCfgDir()).To(BeNil())
	})

	Describe(".TransactionAdd", func() {

		var err error
		var n *node.Node
		var logic *Logic
		var errCh chan error

		BeforeEach(func() {
			errCh = make(chan error)
			n, err = node.NewNode(cfg, "127.0.0.1:40001", crypto.NewKeyFromIntSeed(1), log)
			Expect(err).To(BeNil())
			gossip := node.NewGossip(n, log)
			n.SetGossipProtocol(gossip)
			logic, _ = New(n, log)
		})

		AfterEach(func() {
			n.Host().Close()
		})

		It("should return err if transaction is invalid", func() {
			tx := &wire.Transaction{SenderPubKey: "48nCZsmoU7wvA3__invalid_fS8UhXQv4u43eny8qpT7ubdVxp3kus3eNZP"}
			logic.TransactionAdd(tx, errCh)
			err := <-errCh
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("field:"))
		})

		It("should return err = 'value must be greater than zero' when tx type is wire.TxTypeBalance and has <= 0 value", func() {
			address, _ := crypto.NewKey(nil)
			sender, _ := crypto.NewKey(nil)
			tx := wire.NewTransaction(wire.TxTypeBalance, 1, util.String(address.Addr()), util.String(sender.PubKey().Base58()), "0", "0.1", time.Now().Unix())
			tx.From = util.String(sender.Addr())
			sig, err := wire.TxSign(tx, sender.PrivKey().Base58())
			Expect(err).To(BeNil())
			tx.Sig = sig
			logic.TransactionAdd(tx, errCh)
			err = <-errCh
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("index:0, field:value, error:value must be greater than zero"))
		})

		It("should return err when tx type is wire.TxTypeBalance and fee is less than min fee", func() {
			address, _ := crypto.NewKey(nil)
			sender, _ := crypto.NewKey(nil)
			tx := wire.NewTransaction(wire.TxTypeBalance, 1, util.String(address.Addr()), util.String(sender.PubKey().Base58()), "10", "0.0000000001", time.Now().Unix())
			tx.From = util.String(sender.Addr())
			tx.Hash = tx.ComputeHash()
			sig, err := wire.TxSign(tx, sender.PrivKey().Base58())
			Expect(err).To(BeNil())
			tx.Sig = sig
			logic.TransactionAdd(tx, errCh)
			err = <-errCh
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("index:0, field:fee, error:fee cannot be below the minimum balance transaction fee {0.0100000000000000}"))
		})

		It("should return err when tx type is unknown", func() {
			address, _ := crypto.NewKey(nil)
			sender, _ := crypto.NewKey(nil)
			tx := wire.NewTransaction(0x300, 1, util.String(address.Addr()), util.String(sender.PubKey().Base58()), "10", "100", time.Now().Unix())
			tx.From = util.String(sender.Addr())
			tx.Hash = tx.ComputeHash()
			sig, err := wire.TxSign(tx, sender.PrivKey().Base58())
			Expect(err).To(BeNil())
			tx.Sig = sig
			logic.TransactionAdd(tx, errCh)
			err = <-errCh
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("index:0, field:type, error:unsupported transaction type"))
		})

		It("should successfully add tx to transaction session", func() {
			address, _ := crypto.NewKey(nil)
			sender, _ := crypto.NewKey(nil)
			tx := wire.NewTransaction(wire.TxTypeBalance, 1, util.String(address.Addr()), util.String(sender.PubKey().Base58()), "10", "100", time.Now().Unix())
			tx.From = util.String(sender.Addr())
			tx.Hash = tx.ComputeHash()
			sig, err := wire.TxSign(tx, sender.PrivKey().Base58())
			Expect(err).To(BeNil())
			tx.Sig = sig
			logic.TransactionAdd(tx, errCh)
			err = <-errCh
			Expect(err).To(BeNil())
		})
	})
})
