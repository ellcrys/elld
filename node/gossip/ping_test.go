package gossip_test

import (
	"math/big"
	"time"

	. "github.com/ellcrys/elld/blockchain/testutil"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/node"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/types/core"
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
				// make rp a peer to lp
				lp.PM().AddOrUpdateNode(rp)
				rp.SetLocalNode(lp)
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

		Context("check that core.EventPeerChainInfo is emitted", func() {
			var block2 types.Block

			BeforeEach(func() {
				block2 = MakeBlockWithTotalDifficulty(rp.GetBlockchain(), rp.GetBlockchain().GetBestChain(), sender, sender, new(big.Int).SetInt64(20000000000))
				_, err := rp.GetBlockchain().ProcessBlock(block2)
				Expect(err).To(BeNil())
			})

			Specify("local peer should emit core.EventPeerChainInfo event", func(done Done) {
				wait := make(chan bool)

				go func() {
					evt := <-lp.GetEventEmitter().Once(core.EventPeerChainInfo)
					peerChainInfo := evt.Args[0].(*types.SyncPeerChainInfo)
					Expect(peerChainInfo.PeerID).To(Equal(rp.StringID()))

					rpCurBlock, err := rp.GetBlockchain().ChainReader().Current()
					Expect(err).To(BeNil())
					Expect(peerChainInfo.PeerChainHeight).To(Equal(rpCurBlock.GetNumber()))

					close(wait)
				}()

				err := lp.Gossip().SendPingToPeer(rp)
				Expect(err).To(BeNil())
				<-wait
				close(done)
			})

			Specify("remote peer should emit core.EventPeerChainInfo event", func(done Done) {
				wait := make(chan bool)

				go func() {
					evt := <-rp.GetEventEmitter().Once(core.EventPeerChainInfo)
					peerChainInfo := evt.Args[0].(*types.SyncPeerChainInfo)
					Expect(peerChainInfo.PeerID).To(Equal(lp.StringID()))

					lpCurBlock, err := lp.GetBlockchain().ChainReader().Current()
					Expect(err).To(BeNil())
					Expect(peerChainInfo.PeerChainHeight).To(Equal(lpCurBlock.GetNumber()))
					close(wait)
				}()

				err := lp.Gossip().SendPingToPeer(rp)
				Expect(err).To(BeNil())

				<-wait
				close(done)
			})
		})
	})

})
