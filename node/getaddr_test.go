package node

import (
	"time"

	"github.com/ellcrys/druid/crypto"
	"github.com/ellcrys/druid/testutil"

	"github.com/ellcrys/druid/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Getaddr", func() {

	BeforeEach(func() {
		var err error
		cfg, err = testutil.SetTestCfg()
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		Expect(testutil.RemoveTestCfgDir()).To(BeNil())
	})

	Describe(".sendGetAddr", func() {
		It("should return error.Error('getaddr failed. failed to connect to peer. dial to self attempted')", func() {
			rp, err := NewNode(cfg, "127.0.0.1:30010", crypto.NewAddressFromIntSeed(0), log)
			Expect(err).To(BeNil())
			rpProtoc := NewInception(rp, log)
			rp.Host().Close()
			_, err = rpProtoc.sendGetAddr(rp)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("getaddr failed. failed to connect to peer. dial to self attempted"))
		})

		It("when rp2 timestamp is 3 hours ago, it should not be returned", func() {
			lp, err := NewNode(cfg, "127.0.0.1:30011", crypto.NewAddressFromIntSeed(4), log)
			Expect(err).To(BeNil())
			lpProtoc := NewInception(lp, log)
			defer lp.Host().Close()

			rp, err := NewNode(cfg, "127.0.0.1:30012", crypto.NewAddressFromIntSeed(5), log)
			Expect(err).To(BeNil())
			rpProtoc := NewInception(rp, log)
			rp.SetProtocolHandler(util.GetAddrVersion, rpProtoc.OnGetAddr)
			defer rp.Host().Close()

			rp2, err := NewNode(cfg, "127.0.0.1:30013", crypto.NewAddressFromIntSeed(5), log)
			Expect(err).To(BeNil())
			rp2.Timestamp = time.Now().Add(-3 * time.Hour)
			rp.PM().AddOrUpdatePeer(rp2)
			defer rp2.Host().Close()

			addrs, err := lpProtoc.sendGetAddr(rp)
			Expect(err).To(BeNil())
			Expect(addrs).To(HaveLen(0))
		})

		It("hardcoded seed peer should not be returned", func() {
			lp, err := NewNode(cfg, "127.0.0.1:30011", crypto.NewAddressFromIntSeed(4), log)
			Expect(err).To(BeNil())
			lpProtoc := NewInception(lp, log)
			defer lp.Host().Close()

			rp, err := NewNode(cfg, "127.0.0.1:30012", crypto.NewAddressFromIntSeed(5), log)
			Expect(err).To(BeNil())
			rpProtoc := NewInception(rp, log)
			rp.SetProtocolHandler(util.GetAddrVersion, rpProtoc.OnGetAddr)
			defer rp.Host().Close()

			rp2, _ := NewNode(cfg, "127.0.0.1:30013", crypto.NewAddressFromIntSeed(6), log)
			rp2.isHardcodedSeed = true
			err = rp.PM().AddOrUpdatePeer(rp2)
			Expect(err).To(BeNil())
			defer rp2.Host().Close()
			addrs, err := lpProtoc.sendGetAddr(rp)
			Expect(err).To(BeNil())
			Expect(addrs).To(HaveLen(0))
		})

		It("when address returned is more than MaxAddrsExpected, error must be returned and none of the addresses are added", func() {

			cfg.Node.MaxAddrsExpected = 1

			lp, err := NewNode(cfg, "127.0.0.1:30011", crypto.NewAddressFromIntSeed(4), log)
			Expect(err).To(BeNil())
			lpProtoc := NewInception(lp, log)
			defer lp.Host().Close()

			rp, err := NewNode(cfg, "127.0.0.1:30012", crypto.NewAddressFromIntSeed(5), log)
			Expect(err).To(BeNil())
			rpProtoc := NewInception(rp, log)
			rp.SetProtocolHandler(util.GetAddrVersion, rpProtoc.OnGetAddr)
			defer rp.Host().Close()

			rp2, err := NewNode(cfg, "127.0.0.1:30013", crypto.NewAddressFromIntSeed(6), log)
			Expect(err).To(BeNil())
			err = rp.PM().AddOrUpdatePeer(rp2)
			Expect(err).To(BeNil())
			defer rp2.Host().Close()

			rp3, err := NewNode(cfg, "127.0.0.1:30014", crypto.NewAddressFromIntSeed(7), log)
			Expect(err).To(BeNil())
			rp.PM().AddOrUpdatePeer(rp3)

			addrs, err := lpProtoc.sendGetAddr(rp)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("too many addresses received. Ignoring addresses"))
			Expect(addrs).To(HaveLen(0))
		})
	})
})
