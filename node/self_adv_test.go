package node_test

import (
	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/node"
	"github.com/ellcrys/elld/types"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("GetAddr", func() {

	var lp, rp *node.Node
	var lpPort, rpPort int

	BeforeEach(func() {
		lpPort = getPort()
		rpPort = getPort()

		lp = makeTestNode(lpPort)
		Expect(lp.GetBlockchain().Up()).To(BeNil())
		lp.SetProtocolHandler(config.AddrVersion, lp.Gossip().OnAddr)

		rp = makeTestNode(rpPort)
		Expect(rp.GetBlockchain().Up()).To(BeNil())
		rp.SetProtocolHandler(config.AddrVersion, rp.Gossip().OnAddr)
	})

	AfterEach(func() {
		closeNode(lp)
		closeNode(rp)
	})

	Describe(".SelfAdvertise", func() {
		It("should successfully self advertise peer; remote peer must add the advertised peer", func(done Done) {
			go func() {
				defer GinkgoRecover()
				n := lp.Gossip().SelfAdvertise([]types.Engine{rp})
				Expect(n).To(Equal(1))
			}()

			<-rp.GetEventEmitter().Once(node.EventAddrProcessed)
			knownPeers := rp.PM().GetKnownPeers()
			Expect(knownPeers).To(HaveLen(1))
			Expect(knownPeers[0].StringID()).To(Equal(lp.StringID()))
			close(done)
		})
	})
})
