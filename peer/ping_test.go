package peer

import (
	"time"

	"github.com/ellcrys/druid/configdir"
	"github.com/ellcrys/druid/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Ping", func() {

	var config = &configdir.Config{
		Peer: &configdir.PeerConfig{
			Dev: true,
		},
	}

	Describe(".sendPing", func() {
		It("should return error.Error('ping failed. failed to connect to peer. dial to self attempted')", func() {
			rp, err := NewPeer(config, "127.0.0.1:30000", 0, util.NewNopLogger())
			Expect(err).To(BeNil())
			rpProtoc := NewInception(rp, util.NewNopLogger())
			rp.Host().Close()
			err = rpProtoc.sendPing(rp)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("ping failed. failed to connect to peer. dial to self attempted"))
		})

		It("should return error.Error('failed to verify message signature') when remote peer signature is invalid", func() {
			lp, err := NewPeer(config, "127.0.0.1:30001", 1, util.NewNopLogger())
			Expect(err).To(BeNil())
			lpProtoc := NewInception(lp, util.NewNopLogger())

			rp, err := NewPeer(config, "127.0.0.1:30002", 2, util.NewNopLogger())
			Expect(err).To(BeNil())
			rpProtoc := NewInception(lp, util.NewNopLogger()) // lp should be rp, as such, will cause the protocol to use lp's private key
			rp.SetProtocolHandler(util.PingVersion, rpProtoc.OnPing)

			err = lpProtoc.sendPing(rp)
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(Equal("failed to verify message signature"))
		})

		It("should return nil and update remote peer timestamp locally", func() {
			lp, err := NewPeer(config, "127.0.0.1:30001", 1, util.NewNopLogger())
			Expect(err).To(BeNil())
			lpProtoc := NewInception(lp, util.NewNopLogger())

			rp, err := NewPeer(config, "127.0.0.1:30002", 2, util.NewNopLogger())
			Expect(err).To(BeNil())
			rpProtoc := NewInception(rp, util.NewNopLogger())
			rp.SetProtocolHandler(util.PingVersion, rpProtoc.OnPing)

			lp.PM().AddOrUpdatePeer(rp)
			rp.Timestamp = rp.Timestamp.Add(-2 * time.Hour)

			rpBeforePingTime := rp.Timestamp.Unix()
			err = lpProtoc.sendPing(rp)
			Expect(err).To(BeNil())
			rpAfterPingTime := rp.Timestamp.Unix()

			Expect(rpAfterPingTime > rpBeforePingTime).To(BeTrue())
		})
	})
})
