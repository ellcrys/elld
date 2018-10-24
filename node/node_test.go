package node_test

import (
	"testing"

	. "github.com/ellcrys/elld/blockchain/testutil"
	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/node"
	"github.com/ellcrys/elld/params"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/types/core/objects"
	"github.com/ellcrys/elld/util"
	. "github.com/ncodes/goblin"
	. "github.com/onsi/gomega"
	"github.com/shopspring/decimal"
)

func TestNode(t *testing.T) {
	params.FeePerByte = decimal.NewFromFloat(0.01)
	g := Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("Node", func() {
		var lp, rp *node.Node
		var sender, _ = crypto.NewKey(nil)
		var receiver, _ = crypto.NewKey(nil)
		var lpPort, rpPort int

		g.BeforeEach(func() {
			lpPort = getPort()
			rpPort = getPort()

			lp = makeTestNode(lpPort)
			Expect(lp.GetBlockchain().Up()).To(BeNil())
			lp.SetProtocolHandler(config.HandshakeVersion, lp.Gossip().OnHandshake)
			lp.SetProtocolHandler(config.GetBlockHashesVersion, lp.Gossip().OnGetBlockHashes)

			rp = makeTestNode(rpPort)
			Expect(rp.GetBlockchain().Up()).To(BeNil())
			rp.SetProtocolHandler(config.HandshakeVersion, rp.Gossip().OnHandshake)
			rp.SetProtocolHandler(config.GetBlockHashesVersion, rp.Gossip().OnGetBlockHashes)
			rp.SetProtocolHandler(config.GetBlockBodiesVersion, rp.Gossip().OnGetBlockBodies)

			// Create sender account on the remote peer
			Expect(rp.GetBlockchain().CreateAccount(1, rp.GetBlockchain().GetBestChain(), &objects.Account{
				Type:    objects.AccountTypeBalance,
				Address: util.String(sender.Addr()),
				Balance: "100",
			})).To(BeNil())

			// Create sender account on the local peer
			Expect(lp.GetBlockchain().CreateAccount(1, lp.GetBlockchain().GetBestChain(), &objects.Account{
				Type:    objects.AccountTypeBalance,
				Address: util.String(sender.Addr()),
				Balance: "100",
			})).To(BeNil())
		})

		g.AfterEach(func() {
			closeNode(lp)
			closeNode(rp)
		})

		g.Describe(".ProcessBlockHashes", func() {
			var block2, block3 core.Block

			// Target shape:
			// Remote Peer
			// [1]-[2]-[3]
			//
			// Local Peer
			// [1]
			g.Context("when remote blockchain shape is [1]-[2]-[3] and local blockchain shape: [1]", func() {

				g.BeforeEach(func() {
					block2 = MakeBlockWithSingleTx(rp.GetBlockchain(), rp.GetBlockchain().GetBestChain(), sender, receiver, 1)
					_, err := rp.GetBlockchain().ProcessBlock(block2)
					Expect(err).To(BeNil())

					block3 = MakeBlockWithSingleTx(rp.GetBlockchain(), rp.GetBlockchain().GetBestChain(), sender, receiver, 2)
					_, err = rp.GetBlockchain().ProcessBlock(block3)
					Expect(err).To(BeNil())

					err = lp.Gossip().SendGetBlockHashes(rp, nil)
					Expect(err).To(BeNil())
				})

				g.Context("when block hash queue includes hashes of block [2] and [3] of remote peer", func() {

					g.BeforeEach(func() {
						Expect(lp.GetBlockHashQueue().Size()).To(Equal(2))
					})

					g.Specify("local peer blockchain height should equal remote peer blockchain height", func(done Done) {

						go func() {
							<-lp.GetEventEmitter().Once(node.EventBlockBodiesProcessed)

							lpCurBlock, err := lp.GetBlockchain().ChainReader().Current()
							Expect(err).To(BeNil())

							rpCurBlock, err := rp.GetBlockchain().ChainReader().Current()
							Expect(err).To(BeNil())

							Expect(lpCurBlock.GetNumber()).To(Equal(rpCurBlock.GetNumber()))
							Expect(lpCurBlock.GetHash()).To(Equal(rpCurBlock.GetHash()))
							done()
						}()

						lp.ProcessBlockHashes()
					})
				})
			})
		})
	})
}
