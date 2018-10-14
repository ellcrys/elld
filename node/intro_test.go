package node_test

import (
	"context"

	. "github.com/ellcrys/elld/blockchain/testutil"
	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/node"
	"github.com/ellcrys/elld/types/core/objects"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/wire"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Intro", func() {

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
		rp.SetProtocolHandler(config.IntroVersion, rp.Gossip().OnIntro)

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

	Describe(".SendIntro", func() {
		_ = MakeBlock

		// Connect lp and rp as peers
		BeforeEach(func() {
			lp.Host().Peerstore().AddAddr(rp.ID(), rp.GetAddress().DecapIPFS(), pstore.PermanentAddrTTL)
			err := lp.Host().Connect(context.TODO(), lp.Host().Peerstore().PeerInfo(rp.ID()))
			Expect(err).To(BeNil())
			lp.PM().UpdateLastSeenTime(rp)
			rp.SetLocalNode(lp)
		})

		When("intro is successfully sent, remote peer should receive intro", func() {

			BeforeEach(func(done Done) {
				go func() {
					<-rp.GetEventEmitter().On(node.EventIntroReceived)
					close(done)
				}()
				lp.Gossip().SendIntro(nil)
			})

			Specify("remote peer's intro count must be 1", func() {
				Expect(rp.CountIntros()).To(Equal(1))
			})
		})

		When("the intro message is explicitly passed", func() {
			When("intro is successfully sent, remote peer should receive intro", func() {

				BeforeEach(func(done Done) {
					go func() {
						<-rp.GetEventEmitter().On(node.EventIntroReceived)
						close(done)
					}()
					lp.Gossip().SendIntro(&wire.Intro{
						PeerID: "12D3KooWPR29KSgCH9QkgUEkxyEkKo6Ehg6ubZxD3T74No97RW6L",
					})
				})

				Specify("remote peer's intro count must be 1", func() {
					Expect(rp.CountIntros()).To(Equal(1))
				})
			})
		})
	})

	AfterEach(func() {
		closeNode(lp)
		closeNode(rp)
	})
})
