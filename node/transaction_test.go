package node

import (
	"time"

	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/types/core/objects"
	"github.com/ellcrys/elld/util"

	"github.com/ellcrys/elld/crypto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TransactionTest() bool {
	return Describe("Transaction", func() {
		Describe(".RelayTx", func() {

			// var bchain types.Blockchain
			var err error
			var n, rp *Node
			var proto, rpProto *Gossip
			var sender, address *crypto.Key

			// create test addresses
			BeforeEach(func() {
				address, _ = crypto.NewKey(nil)
				sender, _ = crypto.NewKey(nil)
			})

			// Create test local node
			// Set the nodes blockchain manager
			// Create and set the nodes gossip handler
			BeforeEach(func() {
				n, err = NewNode(cfg, "127.0.0.1:30010", crypto.NewKeyFromIntSeed(0), log)
				Expect(err).To(BeNil())
				n.SetBlockchain(lpBc)
				proto = NewGossip(n, log)
				n.SetGossipProtocol(proto)
			})

			// Create test remote node
			// Set the nodes blockchain manager
			// Create and set the nodes gossip handler
			// Set protocol handler
			BeforeEach(func() {
				rp, err = NewNode(cfg, "127.0.0.1:30011", crypto.NewKeyFromIntSeed(1), log)
				Expect(err).To(BeNil())
				rp.SetBlockchain(rpBc)
				rpProto = NewGossip(rp, log)
				rp.SetGossipProtocol(rpProto)
				rp.SetProtocolHandler(config.TxVersion, rpProto.OnTx)
			})

			// On the remote node blockchain,
			// Create the sender's account
			// with some initial balance
			BeforeEach(func() {
				err = rp.bchain.CreateAccount(1, rp.bchain.GetBestChain(), &objects.Account{
					Type:    objects.AccountTypeBalance,
					Address: util.String(sender.Addr()),
					Balance: "100",
				})
			})

			// Shutdown the test nodes
			AfterEach(func() {
				closeNode(n)
				closeNode(rp)
			})

			It("should return nil and history key of transaction should be in HistoryCache", func() {

				// create and sign test transaction
				tx := objects.NewTransaction(objects.TxTypeBalance, 1, util.String(address.Addr()), util.String(sender.PubKey().Base58()), "1", "0.1", time.Now().Unix())
				tx.Hash = tx.ComputeHash()
				sig, err := objects.TxSign(tx, sender.PrivKey().Base58())
				Expect(err).To(BeNil())
				tx.Sig = sig

				// Call RelayTx on local node's gossip handler
				// and verify expected values
				err = proto.RelayTx(tx, []types.Engine{rp})
				Expect(err).To(BeNil())
				Expect(n.historyCache.Len()).To(Equal(1))
				Expect(n.historyCache.Has(makeTxHistoryKey(tx, rp))).To(BeTrue())
			})

			It("remote node should add tx in its tx pool", func() {

				// Create and sign test transaction
				tx := objects.NewTransaction(objects.TxTypeBalance, 1, util.String(address.Addr()), util.String(sender.PubKey().Base58()), "1", "0.1", time.Now().Unix())
				tx.From = util.String(sender.Addr())
				tx.Hash = tx.ComputeHash()
				sig, err := objects.TxSign(tx, sender.PrivKey().Base58())
				Expect(err).To(BeNil())
				tx.Sig = sig

				// Relay the transaction to the remote peer
				err = n.gProtoc.RelayTx(tx, []types.Engine{rp})
				time.Sleep(100 * time.Millisecond)
				Expect(err).To(BeNil())
				Expect(rp.GetTxPool().Has(tx)).To(BeTrue())
			})

			It("remote node will fail to add tx if its transaction pool is full", func() {

				// Create the remote peer and
				// set pool capacity to zero
				cfg.TxPool.Capacity = 0
				rp, err = NewNode(cfg, "127.0.0.1:30012", crypto.NewKeyFromIntSeed(2), log)
				Expect(err).To(BeNil())
				rp.SetGossipProtocol(proto)

				// Create the test transaction
				tx := objects.NewTransaction(objects.TxTypeBalance, 1, util.String(address.Addr()), util.String(sender.PubKey().Base58()), "1", "0.1", time.Now().Unix())
				tx.From = util.String(sender.Addr())
				tx.Hash = tx.ComputeHash()
				sig, err := objects.TxSign(tx, sender.PrivKey().Base58())
				Expect(err).To(BeNil())
				tx.Sig = sig

				// Relay transaction to remote peer
				err = proto.RelayTx(tx, []types.Engine{rp})
				Expect(err).To(BeNil())

				// Verify pool size of remote peer
				time.Sleep(100 * time.Millisecond)
				Expect(rp.GetTxPool().Has(tx)).To(BeFalse())
			})
		})
	})
}
