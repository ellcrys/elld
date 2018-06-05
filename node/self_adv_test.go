package node

import (
	"time"

	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/testutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/ellcrys/elld/util"
)

var _ = Describe("SelfAdv", func() {

	BeforeEach(func() {
		var err error
		cfg, err = testutil.SetTestCfg()
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		Expect(testutil.RemoveTestCfgDir()).To(BeNil())
	})
	Describe(".SelfAdvertise", func() {

		var err error
		var lp *Node
		var lpProtoc *Inception

		BeforeEach(func() {
			lp, err = NewNode(cfg, "127.0.0.1:30010", crypto.NewKeyFromIntSeed(0), log)
			Expect(err).To(BeNil())
			lpProtoc = NewInception(lp, log)
			lp.SetProtocol(lpProtoc)
			lp.SetProtocolHandler(util.AddrVersion, lpProtoc.OnAddr)
		})

		It("should successfully self advertise peer; remote peer must add the advertised peer", func() {
			p2, err := NewNode(cfg, "127.0.0.1:30011", crypto.NewKeyFromIntSeed(1), log)
			Expect(err).To(BeNil())
			p2.Timestamp = time.Now()
			pt := NewInception(p2, log)
			p2.SetProtocol(pt)
			p2.SetProtocolHandler(util.AddrVersion, pt.OnAddr)

			Expect(p2.PM().knownPeers).To(HaveLen(0))
			n := lpProtoc.SelfAdvertise([]*Node{p2})
			Expect(n).To(Equal(1))
			time.Sleep(5 * time.Millisecond)
			Expect(p2.PM().knownPeers).To(HaveLen(1))
			Expect(p2.PM().knownPeers).To(HaveKey(lp.StringID()))
		})

		AfterEach(func() {
			lp.host.Close()
		})
	})
})
