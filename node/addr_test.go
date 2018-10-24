package node_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/node"
	"github.com/ellcrys/elld/params"
	"github.com/ellcrys/elld/types/core/objects"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/wire"
	. "github.com/ncodes/goblin"
	"github.com/olebedev/emitter"
	. "github.com/onsi/gomega"
	"github.com/shopspring/decimal"
)

func TestAddr(t *testing.T) {
	params.FeePerByte = decimal.NewFromFloat(0.01)
	g := Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("TestAddr", func() {

		var lp, rp *node.Node
		var sender, _ = crypto.NewKey(nil)
		var lpPort, rpPort int

		g.BeforeEach(func() {
			lpPort = getPort()
			rpPort = getPort()

			lp = makeTestNodeWith(lpPort, 399)
			Expect(lp.GetBlockchain().Up()).To(BeNil())

			rp = makeTestNodeWith(rpPort, 382)
			Expect(rp.GetBlockchain().Up()).To(BeNil())

			// On the remote node blockchain,
			// Create the sender's account
			// with some initial balance
			Expect(rp.GetBlockchain().CreateAccount(1, rp.GetBlockchain().GetBestChain(), &objects.Account{
				Type:    objects.AccountTypeBalance,
				Address: util.String(sender.Addr()),
				Balance: "100",
			})).To(BeNil())
		})

		g.AfterEach(func() {
			closeNode(lp)
			closeNode(rp)
		})

		g.Describe(".PickBroadcasters", func() {

			var candidateAddrs = []*wire.Address{
				{Address: "/ip4/172.16.238.10/tcp/9000/ipfs/12D3KooWHHzSeKaY8xuZVzkLbKFfvNgPPeKhFBGrMbNzbm5akpqu"},
				{Address: "/ip4/172.16.238.11/tcp/9000/ipfs/12D3KooWB1b3qZxWJanuhtseF3DmPggHCtG36KZ9ixkqHtdKH9fh"},
				{Address: "/ip4/172.16.238.12/tcp/9000/ipfs/12D3KooWPgam4TzSVCRa4AbhxQnM9abCYR4E9hV57SN7eAjEYn1j"},
				{Address: "/ip4/172.16.238.13/tcp/9000/ipfs/12D3KooWKRyzVWW6ChFjQjK4miCty85Niy49tpPV95XdKu1BcvMA"},
				{Address: "/ip4/172.16.238.14/tcp/9000/ipfs/12D3KooWE4qDcRrueTuRYWUdQZgcy7APZqBngVeXRt4Y6ytHizKV"},
			}

			g.Describe("cache contains no address", func() {
				g.Describe("many candidate addresses is passed", func() {
					g.It("should return N nodes", func() {
						broadcasters := lp.Gossip().PickBroadcasters(candidateAddrs, 2)
						Expect(broadcasters.Len()).To(Equal(2))
						Expect(broadcasters.PeersID()).To(ContainElement(candidateAddrs[2].Address.StringID()))
						Expect(broadcasters.PeersID()).To(ContainElement(candidateAddrs[1].Address.StringID()))
					})
				})

				g.Describe("only one candidate address is passed", func() {
					candidateAddrs := []*wire.Address{
						{Address: "/ip4/172.16.238.12/tcp/9000/ipfs/12D3KooWPgam4TzSVCRa4AbhxQnM9abCYR4E9hV57SN7eAjEYn1j"},
					}

					g.It("Should return the only candidate node", func() {
						broadcasters := lp.Gossip().PickBroadcasters(candidateAddrs, 2)
						Expect(broadcasters.Len()).To(Equal(1))
						Expect(broadcasters.PeersID()).To(ContainElement(candidateAddrs[0].Address.StringID()))
					})
				})
			})

			g.Describe("cache has one address and multiple candidate addresses", func() {
				var addr = util.NodeAddr("/ip4/172.16.238.14/tcp/9000/ipfs/12D3KooWE4qDcRrueTuRYWUdQZgcy7APZqBngVeXRt4Y6ytHizKV")
				var broadcasters *node.BroadcastPeers

				g.BeforeEach(func() {
					n, _ := rp.NodeFromAddr(addr, true)
					broadcasters = lp.Gossip().GetBroadcasters()
					broadcasters.Add(n)
					Expect(broadcasters.Len()).To(Equal(1))
				})

				g.It("should clear the cache and select new addresses", func() {
					broadcasters = lp.Gossip().PickBroadcasters(candidateAddrs, 2)
					Expect(broadcasters.Len()).To(Equal(2))
					Expect(broadcasters.PeersID()).ToNot(ContainElement(addr.StringID()))
				})
			})

			g.Describe("cache has one address and one candidate addresses", func() {
				var addr = util.NodeAddr("/ip4/172.16.238.14/tcp/9000/ipfs/12D3KooWE4qDcRrueTuRYWUdQZgcy7APZqBngVeXRt4Y6ytHizKV")
				var candidateAddrs = []*wire.Address{
					{Address: "/ip4/172.16.238.10/tcp/9000/ipfs/12D3KooWHHzSeKaY8xuZVzkLbKFfvNgPPeKhFBGrMbNzbm5akpqu"},
				}

				g.BeforeEach(func() {
					n, _ := rp.NodeFromAddr(addr, true)
					lp.Gossip().GetBroadcasters().Add(n)
				})

				g.It("should leave the cache untouched and add the only candidate address", func() {
					broadcasters := lp.Gossip().PickBroadcasters(candidateAddrs, 2)
					Expect(broadcasters.Len()).To(Equal(2))
					Expect(broadcasters.PeersID()).To(ContainElement(addr.StringID()))
					Expect(broadcasters.PeersID()).To(ContainElement(candidateAddrs[0].Address.StringID()))
				})
			})

			g.Describe("cache has not expired", func() {
				var broadcasters *node.BroadcastPeers
				var candidateAddrs = []*wire.Address{
					{Address: "/ip4/172.16.238.10/tcp/9000/ipfs/12D3KooWHHzSeKaY8xuZVzkLbKFfvNgPPeKhFBGrMbNzbm5akpqu"},
					{Address: "/ip4/172.16.238.12/tcp/9000/ipfs/12D3KooWPgam4TzSVCRa4AbhxQnM9abCYR4E9hV57SN7eAjEYn1j"},
				}
				var candidateAddrs2 = []*wire.Address{
					{Address: "/ip4/172.16.238.11/tcp/9000/ipfs/12D3KooWHHzSeKaY8xuZVzkLbKFfvNgPPeKhFBGrMbNzbm5akpqd"},
				}

				g.BeforeEach(func() {
					broadcasters = lp.Gossip().PickBroadcasters(candidateAddrs, 2)
					Expect(broadcasters.Len()).To(Equal(2))
				})

				g.It("should return current cache values", func() {
					broadcasters2 := lp.Gossip().PickBroadcasters(candidateAddrs2, 2)
					Expect(broadcasters2).To(Equal(broadcasters))
				})
			})
		})

		g.Describe(".RelayAddresses", func() {

			g.It("should return err='too many items in addr message' when address is more than 10", func() {
				addrs := []*wire.Address{
					{Address: ""}, {Address: ""}, {Address: ""}, {Address: ""}, {Address: ""}, {Address: ""}, {Address: ""},
					{Address: ""}, {Address: ""}, {Address: ""}, {Address: ""},
				}
				errs := lp.Gossip().RelayAddresses(addrs)
				Expect(errs).ToNot(BeEmpty())
				Expect(errs).To(ContainElement(fmt.Errorf("too many addresses in the message")))
			})

			g.It("should return err='no addr to relay' if non of the addresses where relayable", func() {
				addrs := []*wire.Address{
					{Address: ""},
					{Address: ""},
				}
				errs := lp.Gossip().RelayAddresses(addrs)
				Expect(errs).ToNot(BeEmpty())
				Expect(errs).To(ContainElement(fmt.Errorf("no addr to relay")))
			})

			g.Describe("in production mode", func() {
				g.It("should return err when an address is not routable", func() {
					lp.GetCfg().Node.Mode = config.ModeProd
					addrs := []*wire.Address{
						{Address: "/ip4/0.0.0.0/tcp/9000/ipfs/12D3KooWHHzSeKaY8xuZVzkLbKFfvNgPPeKhFBGrMbNzbm5akpqu"},
					}
					errs := lp.Gossip().RelayAddresses(addrs)
					Expect(errs).ToNot(BeEmpty())
					Expect(errs).To(ContainElement(fmt.Errorf("address {/ip4/0.0.0.0/tcp/9000/ipfs/12D3KooWHHzSeKaY8xuZVzkLbKFfvNgPPeKhFBGrMbNzbm5akpqu} is not routable")))
				})
			})

			g.It("should return err='no addr to relay' if address timestamp over 60 minutes", func() {
				addrs := []*wire.Address{
					{Address: "/ip4/127.0.0.1/tcp/9000/ipfs/12D3KooWHHzSeKaY8xuZVzkLbKFfvNgPPeKhFBGrMbNzbm5akpqu", Timestamp: time.Now().Add(61 * time.Minute).Unix()},
				}
				errs := lp.Gossip().RelayAddresses(addrs)
				Expect(errs).ToNot(BeEmpty())
				Expect(errs).To(ContainElement(fmt.Errorf("no addr to relay")))
			})

			g.Describe("with 3 address to relay", func() {

				var p, p2, p3 *node.Node

				g.Before(func() {
					p = makeTestNode(getPort())
					p2 = makeTestNode(getPort())
					p3 = makeTestNode(getPort())
					lp.PM().AddPeer(rp)
				})

				g.After(func() {
					closeNode(p)
					closeNode(p2)
					closeNode(p3)
				})

				g.It("should successfully select 2 relay peers", func() {
					addrs := []*wire.Address{
						{Address: p.GetAddress(), Timestamp: time.Now().Unix()},
						{Address: p2.GetAddress(), Timestamp: time.Now().Unix()},
						{Address: p3.GetAddress(), Timestamp: time.Now().Unix()},
					}
					Expect(p.Gossip().GetBroadcasters().Len()).To(Equal(0))
					p.Gossip().RelayAddresses(addrs)
					Expect(p.Gossip().GetBroadcasters().Len()).To(Equal(2))
				})
			})
		})

		g.Describe(".OnAddr", func() {

			var p, p2, p3 *node.Node
			var evt emitter.Event
			var addrMsg *wire.Addr

			g.BeforeEach(func() {
				p = makeTestNode(getPort())
				p2 = makeTestNode(getPort())
				p3 = makeTestNode(getPort())
				addrMsg = &wire.Addr{
					Addresses: []*wire.Address{
						{Address: p.GetAddress(), Timestamp: time.Now().Unix()},
						{Address: p2.GetAddress(), Timestamp: time.Now().Unix()},
						{Address: p3.GetAddress(), Timestamp: time.Now().Unix()},
					},
				}
			})

			g.AfterEach(func() {
				closeNode(p)
				closeNode(p2)
				closeNode(p3)
			})

			g.Describe("when the number of addresses is below max address expected", func() {

				g.It("should select relay peers from the relayed addresses", func(done Done) {
					stream, c, err := lp.Gossip().NewStream(rp, config.AddrVersion)
					Expect(err).To(BeNil())

					err = node.WriteStream(stream, addrMsg)
					Expect(err).To(BeNil())
					defer c()
					defer stream.Close()

					go func() {
						evt = <-rp.GetEventEmitter().On(node.EventAddressesRelayed)
						Expect(evt.Args).To(BeEmpty())
						Expect(rp.Gossip().GetBroadcasters().Len()).To(Equal(2))
						done()
					}()
				})
			})

			g.XContext("when the number of addresses is above max address expected", func() {
				g.It("should return no error", func(done Done) {
					rp.GetCfg().Node.MaxAddrsExpected = 1
					stream, c, err := lp.Gossip().NewStream(rp, config.AddrVersion)

					Expect(err).To(BeNil())
					defer c()
					defer stream.Close()

					err = node.WriteStream(stream, addrMsg)
					Expect(err).To(BeNil())

					go func() {
						evt = <-rp.GetEventEmitter().On(node.EventAddrProcessed)
						Expect(evt.Args).ToNot(BeEmpty())
						Expect(evt.Args[0].(error).Error()).To(Equal("too many addresses received. Ignoring addresses"))
						done()
					}()
				})
			})
		})
	})

}
