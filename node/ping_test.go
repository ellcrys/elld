package node

import (
	"time"

	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/testutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Ping", func() {

	BeforeEach(func() {
		var err error
		cfg, err = testutil.SetTestCfg()
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		Expect(testutil.RemoveTestCfgDir()).To(BeNil())
	})

	Describe(".sendPing", func() {
		It("should return error.Error('ping failed. failed to connect to peer. dial to self attempted')", func() {
			rp, err := NewNode(cfg, "127.0.0.1:30000", crypto.NewKeyFromIntSeed(0), log)
			Expect(err).To(BeNil())
			rpProtoc := NewGossip(rp, log)
			rp.Host().Close()
			err = rpProtoc.sendPing(rp)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("ping failed. failed to connect to peer. dial to self attempted"))
		})

		It("should return nil and update remote peer timestamp locally", func() {
			lp, err := NewNode(cfg, "127.0.0.1:30001", crypto.NewKeyFromIntSeed(1), log)
			Expect(err).To(BeNil())
			lpProtoc := NewGossip(lp, log)

			rp, err := NewNode(cfg, "127.0.0.1:30002", crypto.NewKeyFromIntSeed(2), log)
			Expect(err).To(BeNil())
			rpProtoc := NewGossip(rp, log)
			rp.SetProtocolHandler(config.PingVersion, rpProtoc.OnPing)

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
