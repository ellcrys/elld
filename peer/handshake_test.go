package peer

import (
	"time"

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

	Describe(".DoSendHandshake", func() {

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

		Context("With 1 active address in remote peer", func() {

			It("local and remote peer must contain 2 active addresses", func() {
				lp, err := NewPeer(config, "127.0.0.1:40000", 0)
				Expect(err).To(BeNil())
				lpProtoc := NewInception(lp)

				rp, err := NewPeer(config, "127.0.0.1:40001", 1)
				Expect(err).To(BeNil())
				rpProtoc := NewInception(rp)
				rp.SetProtocolHandler(util.HandshakeVersion, rpProtoc.OnHandshake)

				// add 1 recent peers remote peer
				p1, _ := NewPeer(config, "127.0.0.1:40002", 2)
				err = rp.PM().AddOrUpdatePeer(p1)
				Expect(err).To(BeNil())

				err = lpProtoc.SendHandshake(rp)
				Expect(err).To(BeNil())

				activePeerRp := rp.PM().GetActivePeers(0)
				activePeerLp := lp.PM().GetActivePeers(0)
				Expect(len(activePeerLp)).To(Equal(2))
				Expect(len(activePeerRp)).To(Equal(2))
			})
		})

		Context("With 1 inactive/old address in remote peer", func() {

			It("local and remote peer must contain 1 active address", func() {
				lp, err := NewPeer(config, "127.0.0.1:4000", 0)
				Expect(err).To(BeNil())
				lpProtoc := NewInception(lp)

				rp, err := NewPeer(config, "127.0.0.1:40001", 1)
				Expect(err).To(BeNil())
				rpProtoc := NewInception(rp)
				rp.SetProtocolHandler(util.HandshakeVersion, rpProtoc.OnHandshake)

				// add 1 recent peers remote peer
				p1, _ := NewPeer(config, "127.0.0.1:40002", 2)
				err = rp.PM().AddOrUpdatePeer(p1)
				Expect(err).To(BeNil())
				p1.Timestamp = p1.Timestamp.Add(-5 * time.Hour)

				err = lpProtoc.SendHandshake(rp)
				Expect(err).To(BeNil())

				activePeerRp := rp.PM().GetActivePeers(0)
				activePeerLp := lp.PM().GetActivePeers(0)
				Expect(len(activePeerLp)).To(Equal(1))
				Expect(len(activePeerRp)).To(Equal(1))
			})
		})

		Context("With 2 inactive/old address in remote peer", func() {

			It("local and remote peer must contain 1 active address", func() {
				lp, err := NewPeer(config, "127.0.0.1:40000", 0)
				Expect(err).To(BeNil())
				lpProtoc := NewInception(lp)

				rp, err := NewPeer(config, "127.0.0.1:40001", 1)
				Expect(err).To(BeNil())
				rpProtoc := NewInception(rp)
				rp.SetProtocolHandler(util.HandshakeVersion, rpProtoc.OnHandshake)

				// add 1 recent peers remote peer
				p1, _ := NewPeer(config, "127.0.0.1:40002", 2)
				err = rp.PM().AddOrUpdatePeer(p1)
				Expect(err).To(BeNil())
				defer p1.host.Close()

				p2, _ := NewPeer(config, "127.0.0.1:40003", 3)
				err = rp.PM().AddOrUpdatePeer(p1)
				Expect(err).To(BeNil())
				defer p2.host.Close()

				p1.Timestamp = p1.Timestamp.Add(-5 * time.Hour)
				p2.Timestamp = p1.Timestamp.Add(-5 * time.Hour)

				err = lpProtoc.SendHandshake(rp)
				Expect(err).To(BeNil())

				activePeerRp := rp.PM().GetActivePeers(0)
				activePeerLp := lp.PM().GetActivePeers(0)
				Expect(len(activePeerLp)).To(Equal(1))
				Expect(len(activePeerRp)).To(Equal(1))
			})
		})

		Context("With 2 active address in remote peer when MaxAddrsExpected is 1", func() {

			config := &configdir.Config{
				Peer: &configdir.PeerConfig{
					Dev:              true,
					MaxAddrsExpected: 1,
				},
			}

			It("should return error 'too many addresses received. Ignoring addresses'", func() {
				lp, err := NewPeer(config, "127.0.0.1:40000", 0)
				Expect(err).To(BeNil())
				lpProtoc := NewInception(lp)

				rp, err := NewPeer(config, "127.0.0.1:40001", 1)
				Expect(err).To(BeNil())
				rpProtoc := NewInception(rp)
				rp.SetProtocolHandler(util.HandshakeVersion, rpProtoc.OnHandshake)

				// add 1 recent peers remote peer
				p1, _ := NewPeer(config, "127.0.0.1:40002", 2)
				err = rp.PM().AddOrUpdatePeer(p1)
				Expect(err).To(BeNil())
				defer p1.host.Close()

				p2, _ := NewPeer(config, "127.0.0.1:40003", 3)
				err = rp.PM().AddOrUpdatePeer(p2)
				Expect(err).To(BeNil())
				defer p2.host.Close()
				err = lpProtoc.SendHandshake(rp)

				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("too many addresses received. Ignoring addresses"))
			})
		})
	})
})
