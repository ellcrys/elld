package gossip_test

import (
	"github.com/ellcrys/elld/node"
	"github.com/ellcrys/elld/node/gossip"
	"github.com/ellcrys/elld/types/core"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SelfAdv", func() {

	var lp, rp *node.Node
	var lpPort, rpPort int

	BeforeEach(func() {
		lpPort = getPort()
		rpPort = getPort()

		lp = makeTestNode(lpPort)
		Expect(lp.GetBlockchain().Up()).To(BeNil())

		rp = makeTestNode(rpPort)
		Expect(rp.GetBlockchain().Up()).To(BeNil())
	})

	AfterEach(func() {
		closeNode(lp)
		closeNode(rp)
	})

	Describe(".SelfAdvertise", func() {
		It("should successfully self advertise peer; remote peer must add the advertised peer", func(done Done) {
			wait := make(chan bool)

			go func() {
				<-rp.GetEventEmitter().Once(gossip.EventAddrProcessed)
				knownPeers := rp.PM().GetPeers()
				Expect(knownPeers).To(HaveLen(1))
				Expect(knownPeers[0].StringID()).To(Equal(lp.StringID()))
				close(wait)
			}()

			n := lp.Gossip().SelfAdvertise([]core.Engine{rp})
			Expect(n).To(Equal(1))

			<-wait
			close(done)
		})
	})

})
