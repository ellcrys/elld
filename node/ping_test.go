package node_test

import (
	"math/big"
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

		It("should return error.Error('ping failed. failed to connect to peer. dial to self attempted')", func() {
			err := rp.Gossip().SendPingToPeer(rp)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("ping failed. failed to connect to peer. dial to self attempted"))
		})

		Context("when remote peer is known to the local peer and the local peer is responsive", func() {

			var rpBeforePingTime int64

			BeforeEach(func() {
				lp.PM().AddOrUpdatePeer(rp)
				rp.SetTimestamp(time.Now().Add(-2 * time.Hour))
				rpBeforePingTime = rp.GetTimestamp().Unix()
			})

			It("should return nil and update remote peer timestamp locally", func() {
				err := lp.Gossip().SendPingToPeer(rp)
				Expect(err).To(BeNil())
				rpAfterPingTime := rp.GetTimestamp().Unix()
				Expect(rpAfterPingTime > rpBeforePingTime).To(BeTrue())
			})
		})

		Context("when remote node has a better (most difficulty, longest chain etc) chain", func() {
			var block2 core.Block
			var err error

			BeforeEach(func() {
				block2, err = rp.GetBlockchain().Generate(&core.GenerateBlockParams{
					Transactions: []core.Transaction{
						objects.NewTx(objects.TxTypeBalance, 1, util.String(sender.Addr()), sender, "0.1", "0.001", time.Now().UnixNano()),
					},
					Creator:                 sender,
					Nonce:                   core.EncodeNonce(1),
					Difficulty:              new(big.Int).SetInt64(131072),
					OverrideTotalDifficulty: new(big.Int).SetInt64(1000000),
					AddFeeAlloc:             true,
				})
				Expect(err).To(BeNil())
				_, err = rp.GetBlockchain().ProcessBlock(block2)
				Expect(err).To(BeNil())
			})

			Specify("local peer should send block hashes request with the current block as the locator", func(done Done) {
				go func() {
					defer GinkgoRecover()
					err := lp.Gossip().SendPingToPeer(rp)
					Expect(err).To(BeNil())
				}()

				lpCurBlock, err := lp.GetBlockchain().ChainReader().Current()
				Expect(err).To(BeNil())

				evt := <-lp.GetEventEmitter().Once(node.EventRequestedBlockHashes)
				Expect(evt.Args[0].(util.Hash)).To(Equal(lpCurBlock.GetHash()))
				close(done)
			})
		})

		Context("when local node has a better (most difficulty, longest chain etc) chain", func() {

			var block2 core.Block
			var err error

			BeforeEach(func() {
				block2, err = lp.GetBlockchain().Generate(&core.GenerateBlockParams{
					Transactions: []core.Transaction{
						objects.NewTx(objects.TxTypeBalance, 1, util.String(sender.Addr()), sender, "0.1", "0.001", time.Now().UnixNano()),
					},
					Creator:                 sender,
					Nonce:                   core.EncodeNonce(1),
					Difficulty:              new(big.Int).SetInt64(131072),
					OverrideTotalDifficulty: new(big.Int).SetInt64(1000000),
					AddFeeAlloc:             true,
				})
				Expect(err).To(BeNil())
				_, err = lp.GetBlockchain().ProcessBlock(block2)
				Expect(err).To(BeNil())
			})

			Specify("remote peer should send block hashes request with its current block as the locator", func(done Done) {
				go func() {
					defer GinkgoRecover()
					err := lp.Gossip().SendPingToPeer(rp)
					Expect(err).To(BeNil())
				}()

				rpCurBlock, err := rp.GetBlockchain().ChainReader().Current()
				Expect(err).To(BeNil())

				evt := <-rp.GetEventEmitter().Once(node.EventRequestedBlockHashes)
				Expect(evt.Args[0].(util.Hash)).To(Equal(rpCurBlock.GetHash()))
				close(done)
			})
		})
	})
})
