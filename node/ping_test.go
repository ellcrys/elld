package node_test

import (
	"math/big"
	"time"

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

var _ = Describe("Ping", func() {

	var lp, rp *node.Node
	var sender, _ = crypto.NewKey(nil)
	var lpPort, rpPort int

	BeforeEach(func() {
		lpPort = getPort()
		rpPort = getPort()

		lp = makeTestNode(lpPort)
		Expect(lp.GetBlockchain().Up()).To(BeNil())
		lp.SetProtocolHandler(config.GetBlockHashesVersion, lp.Gossip().OnGetBlockHashes)
		lp.SetProtocolHandler(config.PingVersion, lp.Gossip().OnPing)

		rp = makeTestNode(rpPort)
		Expect(rp.GetBlockchain().Up()).To(BeNil())
		rp.SetProtocolHandler(config.BlockBodyVersion, rp.Gossip().OnBlockBody)
		rp.SetProtocolHandler(config.GetBlockHashesVersion, rp.Gossip().OnGetBlockHashes)
		rp.SetProtocolHandler(config.GetBlockHashesVersion, rp.Gossip().OnGetBlockHashes)
		rp.SetProtocolHandler(config.GetAddrVersion, rp.Gossip().OnGetAddr)
		rp.SetProtocolHandler(config.PingVersion, rp.Gossip().OnPing)
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

	Describe(".sendPing", func() {

		It("should return err when connection fail", func() {
			err := rp.Gossip().SendPingToPeer(rp)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("dial to self attempted"))
		})

		Context("when remote peer is known to the local peer and the local peer is responsive", func() {

			var rpBeforePingTime int64

			BeforeEach(func() {
				lp.PM().AddOrUpdatePeer(rp)
				rp.SetLastSeen(time.Now().Add(-2 * time.Hour))
				rpBeforePingTime = rp.GetLastSeen().Unix()
			})

			It("should return nil and update remote peer timestamp locally", func() {
				err := lp.Gossip().SendPingToPeer(rp)
				Expect(err).To(BeNil())
				rpAfterPingTime := rp.GetLastSeen().Unix()
				Expect(rpAfterPingTime > rpBeforePingTime).To(BeTrue())
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

				lpCurBlock, err := lp.GetBlockchain().ChainReader().Current()
				Expect(err).To(BeNil())
				Expect(lpCurBlock.GetNumber()).To(Equal(uint64(1)))

				go func() {
					defer GinkgoRecover()
					evt := <-lp.GetEventEmitter().Once(node.EventRequestedBlockHashes)
					locators := evt.Args[0].([]util.Hash)
					Expect(locators).To(HaveLen(1))
					Expect(locators[0]).To(Equal(lpCurBlock.GetHash()))
					close(done)
				}()

				err = lp.Gossip().SendPingToPeer(rp)
				Expect(err).To(BeNil())
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
				rpCurBlock, err := rp.GetBlockchain().ChainReader().Current()
				Expect(err).To(BeNil())
				Expect(rpCurBlock.GetNumber()).To(Equal(uint64(1)))

				go func() {
					defer GinkgoRecover()
					evt := <-rp.GetEventEmitter().Once(node.EventRequestedBlockHashes)
					locators := evt.Args[0].([]util.Hash)
					Expect(locators).To(HaveLen(1))
					Expect(locators[0]).To(Equal(rpCurBlock.GetHash()))
					close(done)
				}()

				err = rp.Gossip().SendPingToPeer(lp)
				Expect(err).To(BeNil())
			})
		})
	})
})
