package gossip_test

import (
	"time"

	"github.com/olebedev/emitter"

	"github.com/ellcrys/elld/blockchain/txpool"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/node"
	"github.com/ellcrys/elld/types/core"
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

		// On the remote node blockchain,
		// Create the sender's account
		// with some initial balance
		Expect(rp.GetBlockchain().CreateAccount(1, rp.GetBlockchain().GetBestChain(), &core.Account{
			Type:    core.AccountTypeBalance,
			Address: util.String(sender.Addr()),
			Balance: "100",
		})).To(BeNil())
	})

	AfterEach(func() {
		closeNode(lp)
		closeNode(rp)
	})

	Describe(".BroadcastTx", func() {
		tx := core.NewTransaction(core.TxTypeBalance, 1, util.String(receiver.Addr()), util.String(sender.PubKey().Base58()), "1", "2.4", time.Now().Unix())
		tx.From = util.String(sender.Addr())
		tx.Hash = tx.ComputeHash()
		sig, _ := core.TxSign(tx, sender.PrivKey().Base58())
		tx.Sig = sig

		BeforeEach(func() {
			Expect(lp.Connect(rp)).To(BeNil())
		})

		Context("when a transaction is successfully relayed", func() {

			var evt emitter.Event
			BeforeEach(func() {
				err := lp.Gossip().BroadcastTx(tx, []core.Engine{rp})
				Expect(err).To(BeNil())

				done := make(chan bool)
				go func() {
					evt = <-rp.GetEventEmitter().On(core.EventTransactionPooled)
					Expect(evt.Args).ToNot(BeEmpty())
					close(done)
				}()
				<-done
			})

			Specify("remote peer must have the transaction in its pool", func() {
				Expect(rp.GetTxPool().Has(tx)).To(BeTrue())
			})
		})

		Context("when transaction failed remote peer's transaction validation", func() {

			var evt emitter.Event
			BeforeEach(func() {
				var tx2 = *tx
				tx2.Sig = []byte("invalid signature")
				err := lp.Gossip().BroadcastTx(&tx2, []core.Engine{rp})
				Expect(err).To(BeNil())

				done := make(chan bool)
				go func() {
					evt = <-rp.GetEventEmitter().On(core.EventTransactionInvalid)
					close(done)
				}()
				<-done
			})

			It("should return error about the transaction's invalid signature", func() {
				Expect(evt.Args).To(HaveLen(2))
				Expect(evt.Args[1].(error).Error()).To(Equal("index:0, field:sig, error:signature is not valid"))
			})
		})

		Context("when transaction type is TypeTxAlloc", func() {

			var evt emitter.Event
			BeforeEach(func() {
				var tx2 = *tx
				tx2.Type = core.TxTypeAlloc

				done := make(chan bool)
				go func() {
					evt = <-rp.GetEventEmitter().On(core.EventTransactionInvalid)
					close(done)
				}()

				err := lp.Gossip().BroadcastTx(&tx2, []core.Engine{rp})
				Expect(err).To(BeNil())

				<-done
			})

			It("should return error about unexpected allocation transaction", func() {
				Expect(evt.Args).To(HaveLen(2))
				Expect(evt.Args[1].(error).Error()).To(Equal("allocation transaction type is not allowed"))
			})
		})

		Context("when the remote peer's transaction pool is full", func() {

			var eventArgs emitter.Event
			BeforeEach(func() {
				rp.SetTxsPool(txpool.New(0))
				err := lp.Gossip().BroadcastTx(tx, []core.Engine{rp})
				Expect(err).To(BeNil())

				done := make(chan bool)
				go func() {
					eventArgs = <-rp.GetEventEmitter().On(core.EventTransactionInvalid)
					close(done)
				}()
				<-done
			})

			It("should not add the transaction to the remote peer's transaction pool", func() {
				Expect(eventArgs.Args).To(HaveLen(2))
				Expect(eventArgs.Args[1].(error).Error()).To(Equal("container is full"))
				Expect(rp.GetTxPool().Has(tx)).To(BeFalse())
			})
		})
	})

})
