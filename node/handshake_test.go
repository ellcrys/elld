package node

import (
	"time"

	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/crypto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func HandshakeTest() bool {
	return Describe("Handshake", func() {

		var lp, rp *Node
		var err error
		var lpGossip, rpGossip *Gossip

		BeforeEach(func() {
			lp, err = NewNode(cfg, "127.0.0.1:31100", crypto.NewKeyFromIntSeed(0), log)
			lpGossip = NewGossip(lp, log)
			lp.SetGossipProtocol(lpGossip)
			lp.SetBlockchain(lpBc)
		})

		BeforeEach(func() {
			rp, err = NewNode(cfg, "127.0.0.1:31101", crypto.NewKeyFromIntSeed(1), log)
			rpGossip = NewGossip(rp, log)
			rp.SetProtocolHandler(config.HandshakeVersion, rpGossip.OnHandshake)
			rp.SetBlockchain(rpBc)
		})

		AfterEach(func() {
			closeNode(lp)
			closeNode(rp)
		})

		Describe(".SendHandshake", func() {

			It("should return err when connection to peer failed", func() {
				err = rpGossip.SendHandshake(rp)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("handshake failed. failed to connect to peer. dial to self attempted"))
			})

			Context("With 0 addresses in local and remote peers", func() {

				It("should return nil when good connection is established, local and remote peer should have 1 active peer each", func() {
					err = lpGossip.SendHandshake(rp)
					Expect(err).To(BeNil())
					time.Sleep(100 * time.Millisecond)
					activePeerRp := rp.PM().GetActivePeers(0)
					activePeerLp := lp.PM().GetActivePeers(0)
					Expect(len(activePeerRp)).To(Equal(1))
					Expect(len(activePeerLp)).To(Equal(1))
				})
			})
		})
	})
}
