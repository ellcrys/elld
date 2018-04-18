package peer

import (
	"github.com/ellcrys/druid/configdir"
	"github.com/ellcrys/druid/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Handshake", func() {
	var config = &configdir.Config{
		Peer: &configdir.PeerConfig{
			Dev:              true,
			MaxAddrsExpected: 100,
		},
	}

	Describe(".SendHandshake", func() {

		Context("With 0 addresses in local and remote peers", func() {

			It("should return error.Error('handshake failed. failed to connect to peer. dial to self attempted')", func() {
				rp, err := NewPeer(config, "127.0.0.1:40001", 1)
				Expect(err).To(BeNil())
				rpProtoc := NewInception(rp)
				rp.Host().Close()
				err = rpProtoc.SendHandshake(rp)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("handshake failed. failed to connect to peer. dial to self attempted"))
			})

			It("should return error.Error('failed to verify message signature') when remote peer signature is invalid", func() {
				lp, err := NewPeer(config, "127.0.0.1:40000", 0)
				Expect(err).To(BeNil())
				lpProtoc := NewInception(lp)

				rp, err := NewPeer(config, "127.0.0.1:40001", 1)
				Expect(err).To(BeNil())
				rpProtoc := NewInception(lp) // lp should be rp, as such, will cause the protocol to use lp's private key
				rp.SetProtocolHandler(util.HandshakeVersion, rpProtoc.OnHandshake)

				err = lpProtoc.SendHandshake(rp)
				Expect(err).NotTo(BeNil())
			})

			It("should return nil when good connection is established, local and remote peer should have 1 active peer each", func() {
				lp, err := NewPeer(config, "127.0.0.1:40000", 0)
				Expect(err).To(BeNil())
				lpProtoc := NewInception(lp)

				rp, err := NewPeer(config, "127.0.0.1:40001", 1)
				Expect(err).To(BeNil())
				rpProtoc := NewInception(rp)
				rp.SetProtocolHandler(util.HandshakeVersion, rpProtoc.OnHandshake)

				err = lpProtoc.SendHandshake(rp)
				Expect(err).To(BeNil())

				activePeerRp := rp.PM().GetActivePeers(0)
				activePeerLp := lp.PM().GetActivePeers(0)
				Expect(len(activePeerRp)).To(Equal(1))
				Expect(len(activePeerLp)).To(Equal(1))
			})
		})
	})
})
