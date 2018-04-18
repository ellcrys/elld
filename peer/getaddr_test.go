package peer

import (
	"time"

	"github.com/ellcrys/druid/configdir"
	"github.com/ellcrys/druid/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Getaddr", func() {
	var config = &configdir.Config{
		Peer: &configdir.PeerConfig{
			Dev:              true,
			MaxAddrsExpected: 5,
		},
	}

	Describe(".sendGetAddr", func() {
		It("should return error.Error('getaddr failed. failed to connect to peer. dial to self attempted')", func() {
			rp, err := NewPeer(config, "127.0.0.1:30010", 0)
			Expect(err).To(BeNil())
			rpProtoc := NewInception(rp)
			rp.Host().Close()
			err = rpProtoc.sendGetAddr(rp)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("getaddr failed. failed to connect to peer. dial to self attempted"))
		})

		It("should return error.Error('failed to verify message signature') when remote peer signature is invalid", func() {
			lp, err := NewPeer(config, "127.0.0.1:30011", 1)
			Expect(err).To(BeNil())
			lpProtoc := NewInception(lp)

			rp, err := NewPeer(config, "127.0.0.1:30012", 2)
			Expect(err).To(BeNil())
			rpProtoc := NewInception(lp) // lp should be rp, as such, will cause the protocol to use lp's private key
			rp.SetProtocolHandler(util.GetAddrVersion, rpProtoc.OnGetAddr)

			err = lpProtoc.sendGetAddr(rp)
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(Equal("failed to verify message signature"))
			lp.Host().Close()
			rp.Host().Close()
		})

		It("rp2 timestamp must not be updated", func() {
			lp, err := NewPeer(config, "127.0.0.1:30011", 4)
			Expect(err).To(BeNil())
			lpProtoc := NewInception(lp)
			defer lp.Host().Close()

			rp, err := NewPeer(config, "127.0.0.1:30012", 5)
			Expect(err).To(BeNil())
			rpProtoc := NewInception(rp)
			rp.SetProtocolHandler(util.GetAddrVersion, rpProtoc.OnGetAddr)
			defer rp.Host().Close()

			rp2, err := NewPeer(config, "127.0.0.1:30013", 6)
			Expect(err).To(BeNil())
			rp2.Timestamp = time.Now().Add(-1 * time.Hour)
			rp2Time := rp2.Timestamp.Unix()
			err = rp.PM().AddOrUpdatePeer(rp2)
			Expect(err).To(BeNil())
			defer rp2.Host().Close()

			err = lpProtoc.sendGetAddr(rp)
			Expect(err).To(BeNil())
			Expect(lpProtoc.PM().KnownPeers()).To(HaveLen(2))
			Expect(rp2Time).To(Equal(rp2.Timestamp.Unix()))
			Expect(rp2Time == rp2.Timestamp.Unix()).To(BeTrue())
		})

		It("when rp2 timestamp is 3 hours ago, it should not be returned", func() {
			lp, err := NewPeer(config, "127.0.0.1:30011", 4)
			Expect(err).To(BeNil())
			lpProtoc := NewInception(lp)
			defer lp.Host().Close()

			rp, err := NewPeer(config, "127.0.0.1:30012", 5)
			Expect(err).To(BeNil())
			rpProtoc := NewInception(rp)
			rp.SetProtocolHandler(util.GetAddrVersion, rpProtoc.OnGetAddr)
			defer rp.Host().Close()

			rp2, err := NewPeer(config, "127.0.0.1:30013", 6)
			Expect(err).To(BeNil())
			rp2.Timestamp = time.Now().Add(-3 * time.Hour)
			err = rp.PM().AddOrUpdatePeer(rp2)
			Expect(err).To(BeNil())
			defer rp2.Host().Close()

			err = lpProtoc.sendGetAddr(rp)
			Expect(err).To(BeNil())

			knownPeers := lpProtoc.PM().KnownPeers()
			Expect(knownPeers).To(HaveLen(1))
			Expect(knownPeers[rp2.StringID()]).To(BeNil())
		})

		It("hardcoded seed peer should not be returned", func() {
			lp, err := NewPeer(config, "127.0.0.1:30011", 4)
			Expect(err).To(BeNil())
			lpProtoc := NewInception(lp)
			defer lp.Host().Close()

			rp, err := NewPeer(config, "127.0.0.1:30012", 5)
			Expect(err).To(BeNil())
			rpProtoc := NewInception(rp)
			rp.SetProtocolHandler(util.GetAddrVersion, rpProtoc.OnGetAddr)
			defer rp.Host().Close()

			rp2, err := NewPeer(config, "127.0.0.1:30013", 6)
			Expect(err).To(BeNil())
			rp2.isHardcodedSeed = true
			err = rp.PM().AddOrUpdatePeer(rp2)
			Expect(err).To(BeNil())
			defer rp2.Host().Close()
			err = lpProtoc.sendGetAddr(rp)
			Expect(err).To(BeNil())

			knownPeers := lpProtoc.PM().KnownPeers()
			Expect(knownPeers).To(HaveLen(1))
		})

		It("when address returned is more than MaxAddrsExpected", func() {

			config := &configdir.Config{
				Peer: &configdir.PeerConfig{
					Dev:              true,
					MaxAddrsExpected: 1,
				},
			}

			lp, err := NewPeer(config, "127.0.0.1:30011", 4)
			Expect(err).To(BeNil())
			lpProtoc := NewInception(lp)
			defer lp.Host().Close()

			rp, err := NewPeer(config, "127.0.0.1:30012", 5)
			Expect(err).To(BeNil())
			rpProtoc := NewInception(rp)
			rp.SetProtocolHandler(util.GetAddrVersion, rpProtoc.OnGetAddr)
			defer rp.Host().Close()

			rp2, err := NewPeer(config, "127.0.0.1:30013", 6)
			Expect(err).To(BeNil())
			err = rp.PM().AddOrUpdatePeer(rp2)
			Expect(err).To(BeNil())
			defer rp2.Host().Close()

			rp3, err := NewPeer(config, "127.0.0.1:30014", 7)
			Expect(err).To(BeNil())
			rp.PM().AddOrUpdatePeer(rp3)

			err = lpProtoc.sendGetAddr(rp)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("too many addresses received. Ignoring addresses"))
		})
	})
})
