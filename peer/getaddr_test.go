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
			rp, err := NewPeer(config, "127.0.0.1:30010", 0, util.NewNopLogger())
			Expect(err).To(BeNil())
			rpProtoc := NewInception(rp, util.NewNopLogger())
			rp.Host().Close()
			err = rpProtoc.sendGetAddr(rp)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("getaddr failed. failed to connect to peer. dial to self attempted"))
		})

		It("should return error.Error('failed to verify message signature') when remote peer signature is invalid", func() {
			lp, err := NewPeer(config, "127.0.0.1:30011", 1, util.NewNopLogger())
			Expect(err).To(BeNil())
			lpProtoc := NewInception(lp, util.NewNopLogger())

			rp, err := NewPeer(config, "127.0.0.1:30012", 2, util.NewNopLogger())
			Expect(err).To(BeNil())
			rpProtoc := NewInception(lp, util.NewNopLogger()) // lp should be rp, as such, will cause the protocol to use lp's private key
			rp.SetProtocolHandler(util.GetAddrVersion, rpProtoc.OnGetAddr)

			err = lpProtoc.sendGetAddr(rp)
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(Equal("failed to verify message signature"))
			lp.Host().Close()
			rp.Host().Close()
		})

		It("when rp2 timestamp is 3 hours ago, it should not be returned", func() {
			lp, err := NewPeer(config, "127.0.0.1:30011", 4, util.NewNopLogger())
			Expect(err).To(BeNil())
			lpProtoc := NewInception(lp, util.NewNopLogger())
			defer lp.Host().Close()

			rp, err := NewPeer(config, "127.0.0.1:30012", 5, util.NewNopLogger())
			Expect(err).To(BeNil())
			rpProtoc := NewInception(rp, util.NewNopLogger())
			rp.SetProtocolHandler(util.GetAddrVersion, rpProtoc.OnGetAddr)
			defer rp.Host().Close()

			rp2, err := NewPeer(config, "127.0.0.1:30013", 6, util.NewNopLogger())
			Expect(err).To(BeNil())
			rp2.Timestamp = time.Now().Add(-3 * time.Hour)
			rp.PM().AddOrUpdatePeer(rp2)
			defer rp2.Host().Close()

			err = lpProtoc.sendGetAddr(rp)
			Expect(err).To(BeNil())

			knownPeers := lpProtoc.PM().KnownPeers()
			Expect(knownPeers).To(HaveLen(0))
			Expect(knownPeers[rp2.StringID()]).To(BeNil())
		})

		It("hardcoded seed peer should not be returned", func() {
			lp, err := NewPeer(config, "127.0.0.1:30011", 4, util.NewNopLogger())
			Expect(err).To(BeNil())
			lpProtoc := NewInception(lp, util.NewNopLogger())
			defer lp.Host().Close()

			rp, err := NewPeer(config, "127.0.0.1:30012", 5, util.NewNopLogger())
			Expect(err).To(BeNil())
			rpProtoc := NewInception(rp, util.NewNopLogger())
			rp.SetProtocolHandler(util.GetAddrVersion, rpProtoc.OnGetAddr)
			defer rp.Host().Close()

			rp2, _ := NewPeer(config, "127.0.0.1:30013", 6, util.NewNopLogger())
			rp2.isHardcodedSeed = true
			err = rp.PM().AddOrUpdatePeer(rp2)
			Expect(err).To(BeNil())
			defer rp2.Host().Close()
			err = lpProtoc.sendGetAddr(rp)
			Expect(err).To(BeNil())

			knownPeers := lpProtoc.PM().KnownPeers()
			Expect(knownPeers).To(HaveLen(0))
		})

		It("when address returned is more than MaxAddrsExpected, error must be returned and none of the addresses are added", func() {

			config := &configdir.Config{
				Peer: &configdir.PeerConfig{
					Dev:              true,
					MaxAddrsExpected: 1,
				},
			}

			lp, err := NewPeer(config, "127.0.0.1:30011", 4, util.NewNopLogger())
			Expect(err).To(BeNil())
			lpProtoc := NewInception(lp, util.NewNopLogger())
			defer lp.Host().Close()

			rp, err := NewPeer(config, "127.0.0.1:30012", 5, util.NewNopLogger())
			Expect(err).To(BeNil())
			rpProtoc := NewInception(rp, util.NewNopLogger())
			rp.SetProtocolHandler(util.GetAddrVersion, rpProtoc.OnGetAddr)
			defer rp.Host().Close()

			rp2, err := NewPeer(config, "127.0.0.1:30013", 6, util.NewNopLogger())
			Expect(err).To(BeNil())
			err = rp.PM().AddOrUpdatePeer(rp2)
			Expect(err).To(BeNil())
			defer rp2.Host().Close()

			rp3, err := NewPeer(config, "127.0.0.1:30014", 7, util.NewNopLogger())
			Expect(err).To(BeNil())
			rp.PM().AddOrUpdatePeer(rp3)

			err = lpProtoc.sendGetAddr(rp)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("too many addresses received. Ignoring addresses"))

			peers := lp.PM().GetActivePeers(0)
			Expect(len(peers)).To(BeZero())
		})
	})
})
