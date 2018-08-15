package node

import (
	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/crypto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func HandshakeTest() bool {
	return Describe("Handshake", func() {
		Describe(".SendHandshake", func() {
			Context("With 0 addresses in local and remote peers", func() {

				It("should return error.Error('handshake failed. failed to connect to peer. dial to self attempted')", func() {
					rp, err := NewNode(cfg, "127.0.0.1:40001", crypto.NewKeyFromIntSeed(1), log)
					Expect(err).To(BeNil())
					rpGossip := NewGossip(rp, log)
					rp.Host().Close()
					err = rpGossip.SendHandshake(rp)
					Expect(err).ToNot(BeNil())
					Expect(err.Error()).To(Equal("handshake failed. failed to connect to peer. dial to self attempted"))
				})

				It("should return nil when good connection is established, local and remote peer should have 1 active peer each", func() {
					lp, err := NewNode(cfg, "127.0.0.1:40000", crypto.NewKeyFromIntSeed(0), log)
					Expect(err).To(BeNil())
					lpGossip := NewGossip(lp, log)
					lp.SetGossipProtocol(lpGossip)

					rp, err := NewNode(cfg, "127.0.0.1:40001", crypto.NewKeyFromIntSeed(1), log)
					Expect(err).To(BeNil())
					rpGossip := NewGossip(rp, log)
					rp.SetProtocolHandler(config.HandshakeVersion, rpGossip.OnHandshake)

					err = lpGossip.SendHandshake(rp)
					Expect(err).To(BeNil())

					activePeerRp := rp.PM().GetActivePeers(0)
					activePeerLp := lp.PM().GetActivePeers(0)
					Expect(len(activePeerRp)).To(Equal(1))
					Expect(len(activePeerLp)).To(Equal(1))
				})
			})
		})
	})
}
