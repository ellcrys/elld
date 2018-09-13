package node_test

import (
	"time"

	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/node"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/types/core/objects"
	"github.com/ellcrys/elld/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Node", func() {
	var lp, rp *node.Node
	var sender, _ = crypto.NewKey(nil)
	var receiver, _ = crypto.NewKey(nil)
	var lpPort, rpPort int

	BeforeEach(func() {
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

	AfterEach(func() {
		closeNode(lp)
		closeNode(rp)
	})

	Describe(".ProcessBlockHashes", func() {

		var block2, block3 core.Block

		// Target shape:
		// Remote Peer
		// [1]-[2]-[3]
		//
		// Local Peer
		// [1]
		Context("when remote blockchain shape is [1]-[2]-[3] and local blockchain shape: [1]", func() {

			BeforeEach(func(done Done) {
				block2 = makeBlock(rp.GetBlockchain(), sender, receiver, time.Now().Unix()-1)
				_, err := rp.GetBlockchain().ProcessBlock(block2)
				Expect(err).To(BeNil())

				block3 = makeBlock(rp.GetBlockchain(), sender, receiver, time.Now().Unix())
				_, err = rp.GetBlockchain().ProcessBlock(block3)
				Expect(err).To(BeNil())

				go func() {
					defer GinkgoRecover()
					err := lp.Gossip().SendGetBlockHashes(rp, util.Hash{})
					Expect(err).To(BeNil())
				}()

				<-lp.GetEventEmitter().On(node.EventReceivedBlockHashes)
				close(done)
			})

			Context("when block hash queue includes hashes of block [2] and [3] of remote peer", func() {
				BeforeEach(func(done Done) {
					lp.GetBlockHashQueue().Append(&node.BlockHash{Hash: block2.GetHash(), Broadcaster: rp})
					lp.GetBlockHashQueue().Append(&node.BlockHash{Hash: block3.GetHash(), Broadcaster: rp})
					go lp.ProcessBlockHashes()
					<-lp.GetEventEmitter().Once(node.EventBlockBodiesProcessed)
					close(done)
				})

				Specify("local peer blockchain height should equal remote peer blockchain height", func() {
					lpCurBlock, err := lp.GetBlockchain().ChainReader().Current()
					Expect(err).To(BeNil())
					rpCurBlock, err := rp.GetBlockchain().ChainReader().Current()
					Expect(err).To(BeNil())
					Expect(lpCurBlock.GetNumber()).To(Equal(rpCurBlock.GetNumber()))
					Expect(lpCurBlock.GetHash()).To(Equal(rpCurBlock.GetHash()))
				})
			})
		})

	})
})
