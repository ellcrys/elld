package gossip_test

import (
	"math/big"
	"testing"
	"time"

	. "github.com/ellcrys/elld/blockchain/testutil"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/node"
	"github.com/ellcrys/elld/params"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/util"
	"github.com/shopspring/decimal"

	. "github.com/ncodes/goblin"
	. "github.com/onsi/gomega"
)

func TestPing(t *testing.T) {
	params.FeePerByte = decimal.NewFromFloat(0.01)
	g := Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("Ping", func() {

		var lp, rp *node.Node
		var sender, _ = crypto.NewKey(nil)
		var lpPort, rpPort int

		g.BeforeEach(func() {
			lpPort = getPort()
			rpPort = getPort()

			lp = makeTestNode(lpPort)
			Expect(lp.GetBlockchain().Up()).To(BeNil())

			rp = makeTestNode(rpPort)
			Expect(rp.GetBlockchain().Up()).To(BeNil())

			// Create sender account on the remote peer
			Expect(rp.GetBlockchain().CreateAccount(1, rp.GetBlockchain().GetBestChain(), &core.Account{
				Type:    core.AccountTypeBalance,
				Address: util.String(sender.Addr()),
				Balance: "100",
			})).To(BeNil())

			// Create sender account on the local peer
			Expect(lp.GetBlockchain().CreateAccount(1, lp.GetBlockchain().GetBestChain(), &core.Account{
				Type:    core.AccountTypeBalance,
				Address: util.String(sender.Addr()),
				Balance: "100",
			})).To(BeNil())
		})

		g.AfterEach(func() {
			closeNode(lp)
			closeNode(rp)
		})

		g.Describe(".sendPing", func() {

			g.It("should return err when connection fail", func() {
				err := rp.Gossip().SendPingToPeer(rp)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("dial to self attempted"))
			})

			g.Context("when remote peer is known to the local peer and the local peer is responsive", func() {

				var rpBeforePingTime int64

				g.BeforeEach(func() {
					// make rp a peer to lp
					lp.PM().AddOrUpdateNode(rp)
					rp.SetLocalNode(lp)
					rp.SetLastSeen(time.Now().Add(-2 * time.Hour))
					rpBeforePingTime = rp.GetLastSeen().Unix()
				})

				g.It("should return nil and update remote peer timestamp locally", func() {
					err := lp.Gossip().SendPingToPeer(rp)
					Expect(err).To(BeNil())
					rpAfterPingTime := rp.GetLastSeen().Unix()
					Expect(rpAfterPingTime > rpBeforePingTime).To(BeTrue())
				})
			})

			g.Context("check that core.EventPeerChainInfo is emitted", func() {
				var block2 types.Block

				g.BeforeEach(func() {
					block2 = MakeBlockWithTotalDifficulty(rp.GetBlockchain(), rp.GetBlockchain().GetBestChain(), sender, sender, new(big.Int).SetInt64(20000000000))
					_, err := rp.GetBlockchain().ProcessBlock(block2)
					Expect(err).To(BeNil())
				})

				g.Specify("local peer should emit core.EventPeerChainInfo event", func(done Done) {

					go func() {
						evt := <-lp.GetEventEmitter().Once(core.EventPeerChainInfo)
						peerChainInfo := evt.Args[0].(*types.SyncPeerChainInfo)
						Expect(peerChainInfo.PeerID).To(Equal(rp.StringID()))

						rpCurBlock, err := rp.GetBlockchain().ChainReader().Current()
						Expect(err).To(BeNil())
						Expect(peerChainInfo.PeerChainHeight).To(Equal(rpCurBlock.GetNumber()))
						done()
					}()

					err := lp.Gossip().SendPingToPeer(rp)
					Expect(err).To(BeNil())
				})

				g.Specify("remote peer should emit core.EventPeerChainInfo event", func(done Done) {
					go func() {
						evt := <-rp.GetEventEmitter().Once(core.EventPeerChainInfo)
						peerChainInfo := evt.Args[0].(*types.SyncPeerChainInfo)
						Expect(peerChainInfo.PeerID).To(Equal(lp.StringID()))

						lpCurBlock, err := lp.GetBlockchain().ChainReader().Current()
						Expect(err).To(BeNil())
						Expect(peerChainInfo.PeerChainHeight).To(Equal(lpCurBlock.GetNumber()))
						done()
					}()

					err := lp.Gossip().SendPingToPeer(rp)
					Expect(err).To(BeNil())
				})
			})
		})
	})
}
