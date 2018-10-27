package gossip_test

import (
	"testing"
	"time"

	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/node"
	"github.com/ellcrys/elld/params"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/util"
	"github.com/shopspring/decimal"

	. "github.com/ncodes/goblin"
	. "github.com/onsi/gomega"
)

func TestGetAddr(t *testing.T) {
	params.FeePerByte = decimal.NewFromFloat(0.01)
	g := Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("GetAddr", func() {

		var lp, rp *node.Node
		var sender, _ = crypto.NewKey(nil)
		var lpPort, rpPort int

		g.BeforeEach(func() {
			lpPort = getPort()
			rpPort = getPort()

			lp = makeTestNode(lpPort)
			Expect(lp.GetBlockchain().Up()).To(BeNil())

			rp = makeTestNode(rpPort)
			Expect(rp.GetBlockchain().Up()).To(BeNil())

			// Create sender account on the remote peer
			Expect(rp.GetBlockchain().CreateAccount(1, rp.GetBlockchain().GetBestChain(), &core.Account{
				Type:    core.AccountTypeBalance,
				Address: util.String(sender.Addr()),
				Balance: "100",
			})).To(BeNil())

			// Create sender account on the local peer
			Expect(lp.GetBlockchain().CreateAccount(1, lp.GetBlockchain().GetBestChain(), &core.Account{
				Type:    core.AccountTypeBalance,
				Address: util.String(sender.Addr()),
				Balance: "100",
			})).To(BeNil())
		})

		g.AfterEach(func() {
			closeNode(lp)
			closeNode(rp)
		})

		g.Describe(".SendGetAddrToPeer", func() {

			g.It("should return error if connection fail", func() {
				_, err := rp.Gossip().SendGetAddrToPeer(rp)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("dial to self attempted"))
			})

			g.Context("when a remote peer knowns an address that is the same as the requesting peer", func() {

				g.BeforeEach(func() {
					rp.PM().AddOrUpdateNode(lp)
				})

				g.It("should not be returned", func() {
					addrs, err := lp.Gossip().SendGetAddrToPeer(rp)
					Expect(err).To(BeNil())
					Expect(addrs).To(HaveLen(0))
				})
			})

			g.Context("when a remote peer knowns an address that is valid", func() {
				var remoteAddr *node.Node

				g.BeforeEach(func() {
					remoteAddr = makeTestNode(getPort())
					remoteAddr.SetLastSeen(time.Now())
					rp.PM().AddOrUpdateNode(remoteAddr)
				})

				g.AfterEach(func() {
					closeNode(remoteAddr)
				})

				g.It("should return the address", func() {
					addrs, err := lp.Gossip().SendGetAddrToPeer(rp)
					Expect(err).To(BeNil())
					Expect(addrs).To(HaveLen(1))
					Expect(addrs[0].Address).To(Equal(remoteAddr.GetAddress()))
				})
			})

			g.Context("when a remote peer knowns an address that is hardcoded", func() {
				var remoteAddr *node.Node

				g.BeforeEach(func() {
					remoteAddr = makeTestNode(getPort())
					remoteAddr.MakeHardcoded()
					remoteAddr.SetLastSeen(time.Now())
					rp.PM().AddOrUpdateNode(remoteAddr)
				})

				g.AfterEach(func() {
					closeNode(remoteAddr)
				})

				g.It("should not return the address", func() {
					addrs, err := lp.Gossip().SendGetAddrToPeer(rp)
					Expect(err).To(BeNil())
					Expect(addrs).To(HaveLen(0))
				})
			})

			g.Context("when a remote peer returns addresses greater than MaxAddrsExpected", func() {
				var remoteAddr *node.Node

				g.BeforeEach(func() {
					lp.GetCfg().Node.MaxAddrsExpected = 0
					remoteAddr = makeTestNode(getPort())
					remoteAddr.SetLastSeen(time.Now())
					rp.PM().AddOrUpdateNode(remoteAddr)
				})

				g.AfterEach(func() {
					closeNode(remoteAddr)
				})

				g.It("should not return the address", func() {
					addrs, err := lp.Gossip().SendGetAddrToPeer(rp)
					Expect(err).ToNot(BeNil())
					Expect(err.Error()).To(Equal("too many addresses received. Ignoring addresses"))
					Expect(addrs).To(HaveLen(0))
				})
			})
		})
	})
}
