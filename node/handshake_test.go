package node

import (
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/testutil"
	"github.com/ellcrys/elld/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Handshake", func() {

	BeforeEach(func() {
		var err error
		cfg, err = testutil.SetTestCfg()
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		Expect(testutil.RemoveTestCfgDir()).To(BeNil())
	})

	Describe(".SendHandshake", func() {

		Context("With 0 addresses in local and remote peers", func() {

			It("should return error.Error('handshake failed. failed to connect to peer. dial to self attempted')", func() {
				rp, err := NewNode(cfg, "127.0.0.1:40001", crypto.NewKeyFromIntSeed(1), log)
				Expect(err).To(BeNil())
				rpProtoc := NewInception(rp, log)
				rp.Host().Close()
				err = rpProtoc.SendHandshake(rp)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("handshake failed. failed to connect to peer. dial to self attempted"))
			})

			It("should return nil when good connection is established, local and remote peer should have 1 active peer each", func() {
				lp, err := NewNode(cfg, "127.0.0.1:40000", crypto.NewKeyFromIntSeed(0), log)
				Expect(err).To(BeNil())
				lpProtoc := NewInception(lp, log)
				lp.SetProtocol(lpProtoc)

				rp, err := NewNode(cfg, "127.0.0.1:40001", crypto.NewKeyFromIntSeed(1), log)
				Expect(err).To(BeNil())
				rpProtoc := NewInception(rp, log)
				rp.SetProtocolHandler(util.HandshakeVersion, rpProtoc.OnHandshake)

				err = lpProtoc.SendHandshake(rp)
				Expect(err).To(BeNil())

				activePeerRp := rp.PM().GetActivePeers(0)
				activePeerLp := lp.PM().GetActivePeers(0)
				Expect(len(activePeerRp)).To(Equal(1))
				Expect(len(activePeerLp)).To(Equal(1))
			})
		})
	})
})
