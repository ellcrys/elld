package peer

import (
	"time"

	"github.com/ellcrys/druid/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Ping", func() {

	Describe(".sendPing", func() {
		It("should return error.Error('handshake failed. failed to connect to peer. dial to self attempted')", func() {
			rp, err := NewPeer(nil, "127.0.0.1:40000", 0)
			Expect(err).To(BeNil())
			rpProtoc := NewInception(rp)
			rp.Host().Close()
			err = rpProtoc.sendPing(rp)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("handshake failed. failed to connect to peer. dial to self attempted"))
		})

		It("should return error.Error('failed to verify message signature') when remote peer signature is invalid", func() {
			lp, err := NewPeer(nil, "127.0.0.1:40001", 1)
			Expect(err).To(BeNil())
			lpProtoc := NewInception(lp)

			rp, err := NewPeer(nil, "127.0.0.1:40002", 2)
			Expect(err).To(BeNil())
			rpProtoc := NewInception(lp) // lp should be rp, as such, will cause the protocol to use lp's private key
			rp.SetProtocolHandler(util.PingVersion, rpProtoc.OnPing)

			err = lpProtoc.sendPing(rp)
			Expect(err).NotTo(BeNil())
		})

		It("should return nil and update remote peer timestamp locally", func() {
			lp, err := NewPeer(nil, "127.0.0.1:40001", 1)
			Expect(err).To(BeNil())
			lpProtoc := NewInception(lp)

			rp, err := NewPeer(nil, "127.0.0.1:40002", 2)
			Expect(err).To(BeNil())

			lp.PM().AddOrUpdatePeer(rp)
			rp.Timestamp = rp.Timestamp.Add(-1 * time.Hour)
			currentTime := rp.Timestamp.Unix()

			rpProtoc := NewInception(rp)
			rp.SetProtocolHandler(util.PingVersion, rpProtoc.OnPing)

			err = lpProtoc.sendPing(rp)
			Expect(err).To(BeNil())

			latestTime := rp.Timestamp.Unix()
			Expect(currentTime < latestTime).To(BeTrue())
			Expect(latestTime - currentTime).To(Equal(int64(3600)))
		})
	})
})
