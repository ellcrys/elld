package gossip_test

import (
	"context"

	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/node"
	"github.com/ellcrys/elld/node/gossip"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/util"
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

		Expect(lp.Connect(rp)).To(BeNil())
	})

	AfterEach(func() {
		closeNode(lp)
		closeNode(rp)
	})

	Describe(".SendIntro", func() {

		// Connect lp and rp as peers
		BeforeEach(func() {
			lp.GetHost().Peerstore().AddAddr(rp.ID(), rp.GetAddress().DecapIPFS(), pstore.PermanentAddrTTL)
			err := lp.GetHost().Connect(context.TODO(), lp.GetHost().Peerstore().PeerInfo(rp.ID()))
			Expect(err).To(BeNil())
			lp.PM().AddOrUpdateNode(rp)
			rp.SetLocalNode(lp)
		})

		When("intro is successfully sent, remote peer should receive intro", func() {
			Specify("remote peer's intro count must be 1", func(done Done) {
				wait := make(chan bool)
				go func() {
					<-rp.GetEventEmitter().On(gossip.EventIntroReceived)
					Expect(rp.CountIntros()).To(Equal(1))
					close(wait)
				}()

				lp.Gossip().SendIntro(nil)
				<-wait
				close(done)
			})
		})

		When("the intro message is explicitly passed", func() {
			When("intro is successfully sent, remote peer should receive intro", func() {
				Specify("remote peer's intro count must be 1", func(done Done) {
					wait := make(chan bool)
					go func() {
						<-rp.GetEventEmitter().On(gossip.EventIntroReceived)
						Expect(rp.CountIntros()).To(Equal(1))
						close(wait)
					}()

					lp.Gossip().SendIntro(&core.Intro{PeerID: "12D3KooWPR29KSgCH9QkgUEkxyEkKo6Ehg6ubZxD3T74No97RW6L"})
					<-wait

					close(done)
				})
			})
		})
	})

})
