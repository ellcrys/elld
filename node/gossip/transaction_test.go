package gossip_test

import (
	"testing"
	"time"

	"github.com/olebedev/emitter"
	"github.com/shopspring/decimal"

	"github.com/ellcrys/elld/blockchain/txpool"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/node"
	"github.com/ellcrys/elld/node/gossip"
	"github.com/ellcrys/elld/params"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/util"
	. "github.com/ncodes/goblin"
	. "github.com/onsi/gomega"
)

func TestTransaction(t *testing.T) {
	params.FeePerByte = decimal.NewFromFloat(0.01)
	g := Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("Transaction", func() {

		var lp, rp *node.Node
		var sender, _ = crypto.NewKey(nil)
		var receiver, _ = crypto.NewKey(nil)
		var lpPort, rpPort int

		g.BeforeEach(func() {
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

		g.AfterEach(func() {
			closeNode(lp)
			closeNode(rp)
		})

		g.Describe(".RelayTx", func() {
			tx := core.NewTransaction(core.TxTypeBalance, 1, util.String(receiver.Addr()), util.String(sender.PubKey().Base58()), "1", "2.4", time.Now().Unix())
			tx.From = util.String(sender.Addr())
			tx.Hash = tx.ComputeHash()
			sig, _ := core.TxSign(tx, sender.PrivKey().Base58())
			tx.Sig = sig

			g.Context("when a transaction is successfully relayed", func() {

				var evt emitter.Event
				g.BeforeEach(func() {
					err := lp.Gossip().RelayTx(tx, []core.Engine{rp})
					Expect(err).To(BeNil())

					done := make(chan bool)
					go func() {
						evt = <-rp.GetEventEmitter().On(gossip.EventTransactionProcessed)
						close(done)
					}()
					<-done
				})

				g.It("expects the history cache to have an item for the transaction", func() {
					Expect(evt.Args).To(BeEmpty())
					Expect(lp.GetHistory().Len()).To(Equal(1))
					Expect(lp.GetHistory().HasMulti(gossip.MakeTxHistoryKey(tx, rp)...)).To(BeTrue())
				})

				g.Specify("remote peer's must have the transaction in its pool", func() {
					Expect(rp.GetTxPool().Has(tx)).To(BeTrue())
				})
			})

			g.Context("when transaction failed remote peer's transaction validation", func() {

				var evt emitter.Event
				g.BeforeEach(func() {
					var tx2 = *tx
					tx2.Sig = []byte("invalid signature")
					err := lp.Gossip().RelayTx(&tx2, []core.Engine{rp})
					Expect(err).To(BeNil())

					done := make(chan bool)
					go func() {
						evt = <-rp.GetEventEmitter().On(gossip.EventTransactionProcessed)
						close(done)
					}()
					<-done
				})

				g.It("should return error about the transaction's invalid signature", func() {
					Expect(evt.Args).To(HaveLen(1))
					Expect(evt.Args[0].(error).Error()).To(Equal("index:0, field:sig, error:signature is not valid"))
				})
			})

			g.Context("when transaction type is TypeTxAlloc", func() {

				var evt emitter.Event
				g.BeforeEach(func() {
					var tx2 = *tx
					tx2.Type = core.TxTypeAlloc

					done := make(chan bool)
					go func() {
						evt = <-rp.GetEventEmitter().On(gossip.EventTransactionProcessed)
						close(done)
					}()

					err := lp.Gossip().RelayTx(&tx2, []core.Engine{rp})
					Expect(err).To(BeNil())

					<-done
				})

				g.It("should return error about unexpected allocation transaction", func() {
					Expect(evt.Args).To(HaveLen(1))
					Expect(evt.Args[0].(error).Error()).To(Equal("Allocation transaction type is not allowed"))
				})
			})

			g.Context("when the remote peer's transaction pool is full", func() {

				var eventArgs emitter.Event
				g.BeforeEach(func() {
					rp.SetTxsPool(txpool.New(0))
					err := lp.Gossip().RelayTx(tx, []core.Engine{rp})
					Expect(err).To(BeNil())

					done := make(chan bool)
					go func() {
						eventArgs = <-rp.GetEventEmitter().On(gossip.EventTransactionProcessed)
						close(done)
					}()
					<-done
				})

				g.It("should not add the transaction to the remote peer's transaction pool", func() {
					Expect(eventArgs.Args).To(HaveLen(1))
					Expect(eventArgs.Args[0].(error).Error()).To(Equal("container is full"))
					Expect(rp.GetTxPool().Has(tx)).To(BeFalse())
				})
			})
		})
	})
}
