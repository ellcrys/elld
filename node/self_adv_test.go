package node_test

import (
	"testing"

	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/node"
	"github.com/ellcrys/elld/params"
	"github.com/ellcrys/elld/types"
	"github.com/shopspring/decimal"

	. "github.com/ncodes/goblin"
	. "github.com/onsi/gomega"
)

func TestSelfAdv(t *testing.T) {
	params.FeePerByte = decimal.NewFromFloat(0.01)
	g := Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("SelfAdv", func() {

		var lp, rp *node.Node
		var lpPort, rpPort int

		g.BeforeEach(func() {
			lpPort = getPort()
			rpPort = getPort()

			lp = makeTestNode(lpPort)
			Expect(lp.GetBlockchain().Up()).To(BeNil())
			lp.SetProtocolHandler(config.AddrVersion, lp.Gossip().OnAddr)

			rp = makeTestNode(rpPort)
			Expect(rp.GetBlockchain().Up()).To(BeNil())
			rp.SetProtocolHandler(config.AddrVersion, rp.Gossip().OnAddr)
		})

		g.AfterEach(func() {
			closeNode(lp)
			closeNode(rp)
		})

		g.Describe(".SelfAdvertise", func() {
			g.It("should successfully self advertise peer; remote peer must add the advertised peer", func(done Done) {
				n := lp.Gossip().SelfAdvertise([]types.Engine{rp})
				Expect(n).To(Equal(1))

				go func() {
					<-rp.GetEventEmitter().Once(node.EventAddrProcessed)
					knownPeers := rp.PM().GetPeers()
					Expect(knownPeers).To(HaveLen(1))
					Expect(knownPeers[0].StringID()).To(Equal(lp.StringID()))
					done()
				}()
			})
		})
	})
}
