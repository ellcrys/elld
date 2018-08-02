package node

import (
	"time"

	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/logic"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/util"

	evbus "github.com/asaskevich/EventBus"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/wire"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TransactionTest() bool {
	return Describe("Transaction", func() {
		Describe(".RelayTx", func() {

			// var bchain types.Blockchain
			var err error
			var n, rp *Node
			var proto, rpProto types.Gossip
			var sender, address *crypto.Key
			var nBus, rpBus evbus.Bus

			BeforeEach(func() {
				address, _ = crypto.NewKey(nil)
				sender, _ = crypto.NewKey(nil)
			})

			BeforeEach(func() {
				// bchain = blockchain.New()
			})

			BeforeEach(func() {
				n, err = NewNode(cfg, "127.0.0.1:30010", crypto.NewKeyFromIntSeed(0), log)
				Expect(err).To(BeNil())
				proto = NewGossip(n, log)
				n.SetGossipProtocol(proto)
				_, nBus = logic.New(n, log)
				n.SetLogicBus(nBus)
			})

			BeforeEach(func() {
				rp, err = NewNode(cfg, "127.0.0.1:30011", crypto.NewKeyFromIntSeed(1), log)
				Expect(err).To(BeNil())
				rpProto = NewGossip(rp, log)
				rp.SetGossipProtocol(rpProto)
				rp.SetProtocolHandler(config.TxVersion, rpProto.OnTx)
				_, rpBus = logic.New(rp, log)
				rp.SetLogicBus(rpBus)
			})

			AfterEach(func() {
				n.Host().Close()
				rp.Host().Close()
			})

			It("should return nil and history key of transaction should be in HistoryCache", func() {
				tx := wire.NewTransaction(wire.TxTypeBalance, 1, address.Addr(), sender.PubKey().Base58(), "1", "0.1", time.Now().Unix())
				tx.Hash = util.ToHex(tx.ComputeHash())
				sig, err := wire.TxSign(tx, sender.PrivKey().Base58())
				Expect(err).To(BeNil())
				tx.Sig = util.ToHex(sig)
				err = proto.RelayTx(tx, []types.Engine{rp})
				Expect(err).To(BeNil())
				Expect(n.historyCache.Len()).To(Equal(1))
				Expect(n.historyCache.Has(makeTxHistoryKey(tx, rp))).To(BeTrue())
			})

			It("remote node should add tx in its tx pool", func() {

				tx := wire.NewTransaction(wire.TxTypeBalance, 1, address.Addr(), sender.PubKey().Base58(), "1", "0.1", time.Now().Unix())
				tx.From = sender.Addr()
				tx.Hash = util.ToHex(tx.ComputeHash())
				sig, err := wire.TxSign(tx, sender.PrivKey().Base58())
				Expect(err).To(BeNil())
				tx.Sig = util.ToHex(sig)

				err = n.gProtoc.RelayTx(tx, []types.Engine{rp})
				time.Sleep(1 * time.Millisecond)
				Expect(err).To(BeNil())
				Expect(rp.GetTxPool().Has(tx)).To(BeTrue())
			})

			It("remote node will fail to add tx if its transaction pool is full", func() {
				cfg.TxPool.Capacity = 0
				rp, err = NewNode(cfg, "127.0.0.1:30012", crypto.NewKeyFromIntSeed(2), log)
				Expect(err).To(BeNil())
				rp.SetGossipProtocol(proto)

				tx := wire.NewTransaction(wire.TxTypeBalance, 1, address.Addr(), sender.PubKey().Base58(), "1", "0.1", time.Now().Unix())
				tx.From = sender.Addr()
				tx.Hash = util.ToHex(tx.ComputeHash())
				sig, err := wire.TxSign(tx, sender.PrivKey().Base58())
				Expect(err).To(BeNil())
				tx.Sig = util.ToHex(sig)

				err = proto.RelayTx(tx, []types.Engine{rp})
				Expect(err).To(BeNil())

				time.Sleep(1 * time.Millisecond)
				Expect(rp.GetTxPool().Has(tx)).To(BeFalse())
			})
		})
	})
}
