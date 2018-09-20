package node_test

import (
	"math/big"

	. "github.com/ellcrys/elld/blockchain/testutil"
	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/node"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/types/core/objects"
	"github.com/ellcrys/elld/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Handshake", func() {

	var lp, rp *node.Node
	var sender, _ = crypto.NewKey(nil)
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

	Describe(".SendHandshake", func() {

		It("should return err when connection to peer failed", func() {
			err := rp.Gossip().SendHandshake(rp)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("handshake failed. failed to connect to peer. dial to self attempted"))
		})

		Context("when local and remote peer have no active addresses", func() {

			var err error

			BeforeEach(func() {
				err = lp.Gossip().SendHandshake(rp)
			})

			It("should return nil when good connection is established", func() {
				Expect(err).To(BeNil())
			})

			Specify("local and remote peer should have 1 active peer each", func() {
				activePeerRp := rp.PM().GetActivePeers(0)
				Expect(activePeerRp).To(HaveLen(1))
				activePeerLp := lp.PM().GetActivePeers(0)
				Expect(activePeerLp).To(HaveLen(1))
			})
		})

		Context("when remote node has a better (most difficulty, longest chain etc) chain", func() {

			var block2 core.Block
			var err error

			BeforeEach(func() {
				block2 = MakeBlockWithTotalDifficulty(rp.GetBlockchain(), rp.GetBlockchain().GetBestChain(), sender, sender, new(big.Int).SetInt64(20000000000))
				Expect(err).To(BeNil())
				_, err = rp.GetBlockchain().ProcessBlock(block2)
				Expect(err).To(BeNil())
			})

			Specify("local peer should send block hashes request with the current block as the locator", func(done Done) {
				go func() {
					defer GinkgoRecover()
					err := lp.Gossip().SendHandshake(rp)
					Expect(err).To(BeNil())
				}()

				lpCurBlock, err := lp.GetBlockchain().ChainReader().Current()
				Expect(err).To(BeNil())

				evt := <-lp.GetEventEmitter().Once(node.EventRequestedBlockHashes)
				locators := evt.Args[0].([]util.Hash)
				Expect(locators).To(HaveLen(1))
				Expect(locators[0]).To(Equal(lpCurBlock.GetHash()))
				close(done)
			})
		})

		Context("when local node has a better (most difficulty, longest chain etc) chain", func() {

			var block2 core.Block
			var err error

			BeforeEach(func() {
				block2 = MakeBlockWithTotalDifficulty(lp.GetBlockchain(), lp.GetBlockchain().GetBestChain(), sender, sender, new(big.Int).SetInt64(20000000000))
				Expect(err).To(BeNil())
				_, err = lp.GetBlockchain().ProcessBlock(block2)
				Expect(err).To(BeNil())
			})

			Specify("remote peer should send block hashes request with its current block as the locator", func(done Done) {
				go func() {
					defer GinkgoRecover()
					err := lp.Gossip().SendHandshake(rp)
					Expect(err).To(BeNil())
				}()

				rpCurBlock, err := rp.GetBlockchain().ChainReader().Current()
				Expect(err).To(BeNil())

				evt := <-rp.GetEventEmitter().Once(node.EventRequestedBlockHashes)
				locators := evt.Args[0].([]util.Hash)
				Expect(locators).To(HaveLen(1))
				Expect(locators[0]).To(Equal(rpCurBlock.GetHash()))
				close(done)
			})
		})
	})
})
