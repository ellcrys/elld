package node

import (
	"time"

	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/crypto"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func PingTest() bool {
	return Describe("Ping", func() {

		var lp, rp *Node
		var err error
		var lpGossip, rpGossip *Gossip

		BeforeEach(func() {
			err := lpBc.Up()
			Expect(err).To(BeNil())
			err = rpBc.Up()
			Expect(err).To(BeNil())
		})

		BeforeEach(func() {
			lp, err = NewNode(cfg, "127.0.0.1:30011", crypto.NewKeyFromIntSeed(4), log)
			Expect(err).To(BeNil())
			lpGossip = NewGossip(lp, log)
			lp.SetGossipProtocol(lpGossip)
			lp.SetBlockchain(lpBc)
		})

		BeforeEach(func() {
			rp, err = NewNode(cfg, "127.0.0.1:30012", crypto.NewKeyFromIntSeed(5), log)
			Expect(err).To(BeNil())
			rpGossip = NewGossip(rp, log)
			rp.SetGossipProtocol(rpGossip)
			rp.SetProtocolHandler(config.PingVersion, rpGossip.OnPing)
			rp.SetBlockchain(rpBc)
		})

		AfterEach(func() {
			closeNode(lp)
			closeNode(rp)
		})

		Describe(".sendPing", func() {
			It("should return error.Error('ping failed. failed to connect to peer. dial to self attempted')", func() {
				rp, err := NewNode(cfg, "127.0.0.1:30000", crypto.NewKeyFromIntSeed(0), log)
				Expect(err).To(BeNil())
				rpProtoc := NewGossip(rp, log)
				defer closeNode(rp)
				err = rpProtoc.sendPing(rp)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("ping failed. failed to connect to peer. dial to self attempted"))
			})

			It("should return nil and update remote peer timestamp locally", func() {
				lp.PM().AddOrUpdatePeer(rp)
				rp.SetTimestamp(rp.Timestamp.Add(-2 * time.Hour))

				rpBeforePingTime := rp.GetTimestamp().Unix()
				err = lpGossip.sendPing(rp)
				Expect(err).To(BeNil())
				rpAfterPingTime := rp.GetTimestamp().Unix()

				Expect(rpAfterPingTime > rpBeforePingTime).To(BeTrue())
			})
		})
	})
}
