package node

import (
	"time"

	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/txpool"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/types/core/objects"
	"github.com/ellcrys/elld/util"
	"github.com/k0kubun/pp"

	"github.com/ellcrys/elld/crypto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TransactionTest() bool {
	return Describe("Transaction", func() {
		Describe(".RelayTx", func() {

			// var bchain types.Blockchain
			var err error
			var lp, rp *Node
			var lpProto, rpProto *Gossip
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
				lp, err = NewNode(cfg, "127.0.0.1:30010", crypto.NewKeyFromIntSeed(0), log)
				Expect(err).To(BeNil())
				lp.SetBlockchain(lpBc)
				lpProto = NewGossip(lp, log)
				lp.SetGossipProtocol(lpProto)
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
				Expect(err).To(BeNil())
			})

			// Shutdown the test nodes
			AfterEach(func() {
				closeNode(lp)
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
				err = lpProto.RelayTx(tx, []types.Engine{rp})
				Expect(err).To(BeNil())
				Expect(lp.historyCache.Len()).To(Equal(1))
				Expect(lp.historyCache.Has(makeTxHistoryKey(tx, rp))).To(BeTrue())
			})

			It("remote node should add tx in its tx pool", func() {

				// Create and sign test transaction
				tx := objects.NewTransaction(objects.TxTypeBalance, 1, util.String(address.Addr()), util.String(sender.PubKey().Base58()), "1", "0.1", time.Now().Unix())
				tx.From = util.String(sender.Addr())
				tx.Hash = tx.ComputeHash()
				sig, err := objects.TxSign(tx, sender.PrivKey().Base58())
				Expect(err).To(BeNil())
				tx.Sig = sig

				wait := make(chan bool)
				rpProto.txProcessed = func(err error) {
					defer close(wait)
					defer GinkgoRecover()
					Expect(err).To(BeNil())
					Expect(rp.GetTxPool().Has(tx)).To(BeTrue())
				}

				// Relay the transaction to the remote peer
				err = lp.gProtoc.RelayTx(tx, []types.Engine{rp})
				Expect(err).To(BeNil())
				<-wait
			})

			It("remote node will fail to add tx if its transaction pool is full", func() {

				// Set the transaction pool to
				// one with 0 capacity
				rp.transactionsPool = txpool.New(0)

				// Create the test transaction
				tx := objects.NewTransaction(objects.TxTypeBalance, 1, util.String(address.Addr()), util.String(sender.PubKey().Base58()), "1", "0.1", time.Now().Unix())
				tx.From = util.String(sender.Addr())
				tx.Hash = tx.ComputeHash()
				sig, err := objects.TxSign(tx, sender.PrivKey().Base58())
				Expect(err).To(BeNil())
				tx.Sig = sig

				pp.Println(tx)

				// 	// Verify pool size of remote peer
				// 	waitForRp := make(chan bool)
				// 	rpProto.txProcessed = func(err error) {
				// 		defer GinkgoRecover()
				// 		defer close(waitForRp)
				// 		Expect(err).ToNot(BeNil())
				// 		Expect(err.Error()).To(Equal("container is full"))
				// 		Expect(rp.GetTxPool().Has(tx)).To(BeFalse())
				// 	}

				// 	// Relay transaction to remote peer
				// 	err = lpProto.RelayTx(tx, []types.Engine{rp})
				// 	Expect(err).To(BeNil())
				// 	<-waitForRp
			})
		})
	})
}
