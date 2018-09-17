package node_test

import (
	"time"

	"github.com/olebedev/emitter"

	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/node"
	"github.com/ellcrys/elld/txpool"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/types/core/objects"
	"github.com/ellcrys/elld/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Transaction", func() {

	var lp, rp *node.Node
	var sender, _ = crypto.NewKey(nil)
	var receiver, _ = crypto.NewKey(nil)
	var lpPort, rpPort int

	BeforeEach(func() {
		lpPort = getPort()
		rpPort = getPort()

		lp = makeTestNode(lpPort)
		Expect(lp.GetBlockchain().Up()).To(BeNil())

		rp = makeTestNode(rpPort)
		Expect(rp.GetBlockchain().Up()).To(BeNil())
		rp.SetProtocolHandler(config.TxVersion, rp.Gossip().OnTx)

		// On the remote node blockchain,
		// Create the sender's account
		// with some initial balance
		Expect(rp.GetBlockchain().CreateAccount(1, rp.GetBlockchain().GetBestChain(), &objects.Account{
			Type:    objects.AccountTypeBalance,
			Address: util.String(sender.Addr()),
			Balance: "100",
		})).To(BeNil())
	})

	AfterEach(func() {
		closeNode(lp)
		closeNode(rp)
	})

	Describe(".RelayTx", func() {
		tx := objects.NewTransaction(objects.TxTypeBalance, 1, util.String(receiver.Addr()), util.String(sender.PubKey().Base58()), "1", "2.4", time.Now().Unix())
		tx.From = util.String(sender.Addr())
		tx.Hash = tx.ComputeHash()
		sig, _ := objects.TxSign(tx, sender.PrivKey().Base58())
		tx.Sig = sig

		Context("when a transaction is successfully relayed", func() {

			var evt emitter.Event
			BeforeEach(func(done Done) {
				err := lp.Gossip().RelayTx(tx, []types.Engine{rp})
				Expect(err).To(BeNil())

				go func() {
					evt = <-rp.GetEventEmitter().On(node.EventTransactionProcessed)
					close(done)
				}()
			})

			It("expects the history cache to have an item for the transaction", func() {
				Expect(evt.Args).To(BeEmpty())
				Expect(lp.GetHistory().Len()).To(Equal(1))
				Expect(lp.GetHistory().HasMulti(node.MakeTxHistoryKey(tx, rp)...)).To(BeTrue())
			})

			Specify("remote peer's must have the transaction in its pool", func() {
				Expect(rp.GetTxPool().Has(tx)).To(BeTrue())
			})
		})

		Context("when transaction failed remote peer's transaction validation", func() {

			var evt emitter.Event
			BeforeEach(func(done Done) {
				var tx2 = *tx
				tx2.Sig = []byte("invalid signature")
				err := lp.Gossip().RelayTx(&tx2, []types.Engine{rp})
				Expect(err).To(BeNil())
				go func() {
					evt = <-rp.GetEventEmitter().On(node.EventTransactionProcessed)
					close(done)
				}()
			})

			It("should return error about the transaction's invalid signature", func() {
				Expect(evt.Args).To(HaveLen(1))
				Expect(evt.Args[0].(error).Error()).To(Equal("index:0, field:sig, error:signature is not valid"))
			})
		})

		Context("when transaction type is TypeTxAlloc", func() {

			var evt emitter.Event
			BeforeEach(func(done Done) {
				var tx2 = *tx
				tx2.Type = objects.TxTypeAlloc

				go func() {
					evt = <-rp.GetEventEmitter().On(node.EventTransactionProcessed)
					close(done)
				}()

				err := lp.Gossip().RelayTx(&tx2, []types.Engine{rp})
				Expect(err).To(BeNil())

			})

			It("should return error about unexpected allocation transaction", func() {
				Expect(evt.Args).To(HaveLen(1))
				Expect(evt.Args[0].(error).Error()).To(Equal("unexpected allocation transaction received"))
			})
		})

		Context("when the remote peer's transaction pool is full", func() {

			var eventArgs emitter.Event
			BeforeEach(func(done Done) {
				rp.SetTransactionPool(txpool.New(0))
				err := lp.Gossip().RelayTx(tx, []types.Engine{rp})
				Expect(err).To(BeNil())

				go func() {
					eventArgs = <-rp.GetEventEmitter().On(node.EventTransactionProcessed)
					close(done)
				}()
			})

			It("should not add the transaction to the remote peer's transaction pool", func() {
				Expect(eventArgs.Args).To(HaveLen(1))
				Expect(eventArgs.Args[0].(error).Error()).To(Equal("container is full"))
				Expect(rp.GetTxPool().Has(tx)).To(BeFalse())
			})
		})
	})
})
