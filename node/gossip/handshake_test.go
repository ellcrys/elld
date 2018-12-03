package gossip_test

import (
	"math/big"
	"testing"

	. "github.com/ellcrys/elld/blockchain/testutil"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/node"
	"github.com/ellcrys/elld/params"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/util"
	. "github.com/ncodes/goblin"
	. "github.com/onsi/gomega"
	"github.com/shopspring/decimal"
)

func TestHandshake(t *testing.T) {
	params.FeePerByte = decimal.NewFromFloat(0.01)
	g := Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("Handshake", func() {

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

		g.Describe(".SendHandshake", func() {

			g.It("should return err when connection to peer failed", func() {
				err := rp.Gossip().SendHandshake(rp)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("dial to self attempted"))
			})

			g.Context("when local and remote peer have no active addresses", func() {

				var err error

				g.BeforeEach(func() {
					err = lp.Gossip().SendHandshake(rp)
				})

				g.It("should return nil when good connection is established", func() {
					Expect(err).To(BeNil())
				})

				g.Specify("local and remote peer should have 1 active peer each", func() {
					activePeerRp := rp.PM().GetActivePeers(0)
					Expect(activePeerRp).To(HaveLen(1))
					activePeerLp := lp.PM().GetActivePeers(0)
					Expect(activePeerLp).To(HaveLen(1))
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

					err := lp.Gossip().SendHandshake(rp)
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

					err := lp.Gossip().SendHandshake(rp)
					Expect(err).To(BeNil())
				})
			})
		})
	})
}
