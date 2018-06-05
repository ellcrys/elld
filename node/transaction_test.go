package node

import (
	"time"

	"github.com/ellcrys/elld/util"

	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/testutil"
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

	Describe(".RelayTx", func() {

		var err error
		var n, rp *Node
		var proto Protocol
		var sender, address *crypto.Key

		BeforeEach(func() {
			address, _ = crypto.NewKey(nil)
			sender, _ = crypto.NewKey(nil)
		})

		BeforeEach(func() {
			n, err = NewNode(cfg, "127.0.0.1:30010", crypto.NewKeyFromIntSeed(0), log)
			Expect(err).To(BeNil())
			proto = NewInception(n, log)
		})

		BeforeEach(func() {
			rp, err = NewNode(cfg, "127.0.0.1:30011", crypto.NewKeyFromIntSeed(1), log)
			Expect(err).To(BeNil())
			rpProto := NewInception(rp, log)
			rp.SetProtocolHandler(util.TxVersion, rpProto.OnTx)
		})

		AfterEach(func() {
			n.Host().Close()
			rp.Host().Close()
		})

		It("should return nil and history key of transaction should be in HistoryCache", func() {
			tx := wire.NewTransaction(wire.TxTypeA2A, 1, address.Addr(), sender.PubKey().Base58(), "1", "0.1", time.Now().Unix())
			tx.Sig, err = wire.TxSign(tx, sender.PrivKey().Base58())
			err = proto.RelayTx(tx, []*Node{rp})
			Expect(err).To(BeNil())
			Expect(n.historyCache.Len()).To(Equal(1))
			Expect(n.historyCache.Has(makeTxHistoryKey(tx, rp))).To(BeTrue())
		})

		It("remote node should add tx in its tx pool", func() {
			tx := wire.NewTransaction(wire.TxTypeA2A, 1, address.Addr(), sender.PubKey().Base58(), "1", "0.1", time.Now().Unix())
			tx.Sig, err = wire.TxSign(tx, sender.PrivKey().Base58())
			Expect(err).To(BeNil())
			err = proto.RelayTx(tx, []*Node{rp})
			Expect(err).To(BeNil())
			time.Sleep(1 * time.Millisecond)
			Expect(rp.txPool.Has(tx)).To(BeTrue())
		})

		It("remote node should add tx in its tx pool", func() {
			tx := wire.NewTransaction(wire.TxTypeA2A, 1, address.Addr(), sender.PubKey().Base58(), "1", "0.1", time.Now().Unix())
			tx.Sig, err = wire.TxSign(tx, sender.PrivKey().Base58())
			Expect(err).To(BeNil())
			err = proto.RelayTx(tx, []*Node{rp})
			Expect(err).To(BeNil())
			time.Sleep(1 * time.Millisecond)
			Expect(rp.txPool.Has(tx)).To(BeTrue())
		})

		It("remote node will fail to add tx if its transaction pool is full", func() {
			cfg.TxPool.Capacity = 0
			rp, err = NewNode(cfg, "127.0.0.1:30012", crypto.NewKeyFromIntSeed(2), log)
			Expect(err).To(BeNil())

			tx := wire.NewTransaction(wire.TxTypeA2A, 1, address.Addr(), sender.PubKey().Base58(), "1", "0.1", time.Now().Unix())
			tx.Sig, err = wire.TxSign(tx, sender.PrivKey().Base58())
			Expect(err).To(BeNil())
			err = proto.RelayTx(tx, []*Node{rp})
			Expect(err).To(BeNil())

			time.Sleep(1 * time.Millisecond)
			Expect(rp.txPool.Has(tx)).To(BeFalse())
		})
	})
})
