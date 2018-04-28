package node

import (
	"time"

	"github.com/ellcrys/druid/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Ping", func() {

	BeforeEach(func() {
		Expect(setTestCfg()).To(BeNil())
	})

	AfterEach(func() {
		Expect(removeTestCfgDir()).To(BeNil())
	})

	Describe(".sendPing", func() {
		It("should return error.Error('ping failed. failed to connect to peer. dial to self attempted')", func() {
			rp, err := NewNode(cfg, "127.0.0.1:30000", 0, log)
			Expect(err).To(BeNil())
			rpProtoc := NewInception(rp, log)
			rp.Host().Close()
			err = rpProtoc.sendPing(rp)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("ping failed. failed to connect to peer. dial to self attempted"))
		})

		It("should return error.Error('failed to verify message signature') when remote peer signature is invalid", func() {
			lp, err := NewNode(cfg, "127.0.0.1:30001", 1, log)
			Expect(err).To(BeNil())
			lpProtoc := NewInception(lp, log)

			rp, err := NewNode(cfg, "127.0.0.1:30002", 2, log)
			Expect(err).To(BeNil())
			rpProtoc := NewInception(lp, log) // lp should be rp, as such, will cause the protocol to use lp's private key
			rp.SetProtocolHandler(util.PingVersion, rpProtoc.OnPing)

			err = lpProtoc.sendPing(rp)
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(Equal("failed to verify message signature"))
		})

		It("should return nil and update remote peer timestamp locally", func() {
			lp, err := NewNode(cfg, "127.0.0.1:30001", 1, log)
			Expect(err).To(BeNil())
			lpProtoc := NewInception(lp, log)

			rp, err := NewNode(cfg, "127.0.0.1:30002", 2, log)
			Expect(err).To(BeNil())
			rpProtoc := NewInception(rp, log)
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
