package node

import (
	"time"

	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/testutil"

	"github.com/ellcrys/elld/util"
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

	var lp, rp, rp2 *Node
	var err error
	var lpProtoc, rpProtoc *Inception

	BeforeEach(func() {
		lp, err = NewNode(cfg, "127.0.0.1:30011", crypto.NewKeyFromIntSeed(4), log)
		Expect(err).To(BeNil())
		lpProtoc = NewInception(lp, log)
	})

	BeforeEach(func() {
		rp, err = NewNode(cfg, "127.0.0.1:30012", crypto.NewKeyFromIntSeed(5), log)
		Expect(err).To(BeNil())
		rpProtoc = NewInception(rp, log)
		rp.SetProtocolHandler(util.GetAddrVersion, rpProtoc.OnGetAddr)
	})

	BeforeEach(func() {
		rp2, err = NewNode(cfg, "127.0.0.1:30013", crypto.NewKeyFromIntSeed(6), log)
		Expect(err).To(BeNil())
		err = rp.PM().AddOrUpdatePeer(rp2)
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		lp.Host().Close()
		rp.Host().Close()
		rp2.Host().Close()
	})

	Describe(".sendGetAddr", func() {
		It("should return error.Error('getaddr failed. failed to connect to peer. dial to self attempted')", func() {
			rp, err := NewNode(cfg, "127.0.0.1:30010", crypto.NewKeyFromIntSeed(0), log)
			Expect(err).To(BeNil())
			rpProtoc := NewInception(rp, log)
			rp.Host().Close()
			_, err = rpProtoc.sendGetAddr(rp)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("getaddr failed. failed to connect to peer. dial to self attempted"))
		})

		It("when rp2 timestamp is 3 hours ago, it should not be returned", func() {
			rp2.Timestamp = time.Now().Add(-3 * time.Hour)
			err := rp.PM().AddOrUpdatePeer(rp2)
			Expect(err).To(BeNil())
			addrs, err := lpProtoc.sendGetAddr(rp)
			Expect(err).To(BeNil())
			Expect(addrs).To(HaveLen(0))
		})

		It("hardcoded seed peer should not be returned", func() {
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

			Expect(err).To(BeNil())
			err = rp.PM().AddOrUpdatePeer(rp2)
			Expect(err).To(BeNil())

			rp3, err := NewNode(cfg, "127.0.0.1:30014", crypto.NewKeyFromIntSeed(7), log)
			Expect(err).To(BeNil())
			rp.PM().AddOrUpdatePeer(rp3)

			addrs, err := lpProtoc.sendGetAddr(rp)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("too many addresses received. Ignoring addresses"))
			Expect(addrs).To(HaveLen(0))
		})
	})
})
