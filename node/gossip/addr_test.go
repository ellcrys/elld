package gossip_test

import (
	"fmt"
	"time"

	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/node"
	"github.com/ellcrys/elld/node/gossip"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/util"
	"github.com/olebedev/emitter"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TestAddr", func() {

	var lp, rp *node.Node
	var sender, _ = crypto.NewKey(nil)
	var lpPort, rpPort int

	BeforeEach(func() {
		lpPort = getPort()
		rpPort = getPort()

		lp = makeTestNodeWith(lpPort, 399)
		Expect(lp.GetBlockchain().Up()).To(BeNil())

		rp = makeTestNodeWith(rpPort, 382)
		Expect(rp.GetBlockchain().Up()).To(BeNil())

		// On the remote node blockchain,
		// Create the sender's account
		// with some initial balance
		Expect(rp.GetBlockchain().CreateAccount(1, rp.GetBlockchain().GetBestChain(), &core.Account{
			Type:    core.AccountTypeBalance,
			Address: util.String(sender.Addr()),
			Balance: "100",
		})).To(BeNil())
	})

	AfterEach(func() {
		closeNode(lp)
		closeNode(rp)
	})

	Describe(".PickBroadcasters", func() {

		var candidateAddrs = []*core.Address{
			{Address: "/ip4/172.16.238.10/tcp/9000/ipfs/12D3KooWHHzSeKaY8xuZVzkLbKFfvNgPPeKhFBGrMbNzbm5akpqu"},
			{Address: "/ip4/172.16.238.11/tcp/9000/ipfs/12D3KooWB1b3qZxWJanuhtseF3DmPggHCtG36KZ9ixkqHtdKH9fh"},
			{Address: "/ip4/172.16.238.12/tcp/9000/ipfs/12D3KooWPgam4TzSVCRa4AbhxQnM9abCYR4E9hV57SN7eAjEYn1j"},
			{Address: "/ip4/172.16.238.13/tcp/9000/ipfs/12D3KooWKRyzVWW6ChFjQjK4miCty85Niy49tpPV95XdKu1BcvMA"},
			{Address: "/ip4/172.16.238.14/tcp/9000/ipfs/12D3KooWE4qDcRrueTuRYWUdQZgcy7APZqBngVeXRt4Y6ytHizKV"},
		}

		Describe("cache contains no address", func() {
			Describe("many candidate addresses is passed", func() {
				It("should return N nodes", func() {
					broadcasters := lp.Gossip().PickBroadcasters(candidateAddrs, 2)
					Expect(broadcasters.Len()).To(Equal(2))
					Expect(broadcasters.PeersID()).To(ContainElement(candidateAddrs[2].Address.StringID()))
					Expect(broadcasters.PeersID()).To(ContainElement(candidateAddrs[1].Address.StringID()))
				})
			})

			Describe("only one candidate address is passed", func() {
				candidateAddrs := []*core.Address{
					{Address: "/ip4/172.16.238.12/tcp/9000/ipfs/12D3KooWPgam4TzSVCRa4AbhxQnM9abCYR4E9hV57SN7eAjEYn1j"},
				}

				It("Should return the only candidate node", func() {
					broadcasters := lp.Gossip().PickBroadcasters(candidateAddrs, 2)
					Expect(broadcasters.Len()).To(Equal(1))
					Expect(broadcasters.PeersID()).To(ContainElement(candidateAddrs[0].Address.StringID()))
				})
			})
		})

		Describe("cache has one address and multiple candidate addresses", func() {
			var addr = util.NodeAddr("/ip4/172.16.238.14/tcp/9000/ipfs/12D3KooWE4qDcRrueTuRYWUdQZgcy7APZqBngVeXRt4Y6ytHizKV")
			var broadcasters *core.BroadcastPeers

			BeforeEach(func() {
				n := rp.NewRemoteNode(addr)
				broadcasters = lp.Gossip().GetBroadcasters()
				broadcasters.Add(n)
				Expect(broadcasters.Len()).To(Equal(1))
			})

			It("should clear the cache and select new addresses", func() {
				broadcasters = lp.Gossip().PickBroadcasters(candidateAddrs, 2)
				Expect(broadcasters.Len()).To(Equal(2))
				Expect(broadcasters.PeersID()).ToNot(ContainElement(addr.StringID()))
			})
		})

		Describe("cache has one address and one candidate addresses", func() {
			var addr = util.NodeAddr("/ip4/172.16.238.14/tcp/9000/ipfs/12D3KooWE4qDcRrueTuRYWUdQZgcy7APZqBngVeXRt4Y6ytHizKV")
			var candidateAddrs = []*core.Address{
				{Address: "/ip4/172.16.238.10/tcp/9000/ipfs/12D3KooWHHzSeKaY8xuZVzkLbKFfvNgPPeKhFBGrMbNzbm5akpqu"},
			}

			BeforeEach(func() {
				n := rp.NewRemoteNode(addr)
				lp.Gossip().GetBroadcasters().Add(n)
			})

			It("should leave the cache untouched and add the only candidate address", func() {
				broadcasters := lp.Gossip().PickBroadcasters(candidateAddrs, 2)
				Expect(broadcasters.Len()).To(Equal(2))
				Expect(broadcasters.PeersID()).To(ContainElement(addr.StringID()))
				Expect(broadcasters.PeersID()).To(ContainElement(candidateAddrs[0].Address.StringID()))
			})
		})

		Describe("cache has not expired", func() {
			var broadcasters *core.BroadcastPeers
			var candidateAddrs = []*core.Address{
				{Address: "/ip4/172.16.238.10/tcp/9000/ipfs/12D3KooWHHzSeKaY8xuZVzkLbKFfvNgPPeKhFBGrMbNzbm5akpqu"},
				{Address: "/ip4/172.16.238.12/tcp/9000/ipfs/12D3KooWPgam4TzSVCRa4AbhxQnM9abCYR4E9hV57SN7eAjEYn1j"},
			}
			var candidateAddrs2 = []*core.Address{
				{Address: "/ip4/172.16.238.11/tcp/9000/ipfs/12D3KooWHHzSeKaY8xuZVzkLbKFfvNgPPeKhFBGrMbNzbm5akpqd"},
			}

			BeforeEach(func() {
				broadcasters = lp.Gossip().PickBroadcasters(candidateAddrs, 2)
				Expect(broadcasters.Len()).To(Equal(2))
			})

			It("should return current cache values", func() {
				broadcasters2 := lp.Gossip().PickBroadcasters(candidateAddrs2, 2)
				Expect(broadcasters2).To(Equal(broadcasters))
			})
		})
	})

	Describe(".RelayAddresses", func() {

		It("should return err='too many items in addr message' when address is more than 10", func() {
			addrs := []*core.Address{
				{Address: ""}, {Address: ""}, {Address: ""}, {Address: ""}, {Address: ""}, {Address: ""}, {Address: ""},
				{Address: ""}, {Address: ""}, {Address: ""}, {Address: ""},
			}
			errs := lp.Gossip().RelayAddresses(addrs)
			Expect(errs).ToNot(BeEmpty())
			Expect(errs).To(ContainElement(fmt.Errorf("too many addresses in the message")))
		})

		It("should return err='no addr to relay' if non of the addresses where relayable", func() {
			addrs := []*core.Address{
				{Address: ""},
				{Address: ""},
			}
			errs := lp.Gossip().RelayAddresses(addrs)
			Expect(errs).ToNot(BeEmpty())
			Expect(errs).To(ContainElement(fmt.Errorf("no addr to relay")))
		})

		Describe("in production mode", func() {
			It("should return err when an address is not routable", func() {
				lp.GetCfg().Node.Mode = config.ModeProd
				addrs := []*core.Address{
					{Address: "/ip4/0.0.0.0/tcp/9000/ipfs/12D3KooWHHzSeKaY8xuZVzkLbKFfvNgPPeKhFBGrMbNzbm5akpqu"},
				}
				errs := lp.Gossip().RelayAddresses(addrs)
				Expect(errs).ToNot(BeEmpty())
				Expect(errs).To(ContainElement(fmt.Errorf("address {/ip4/0.0.0.0/tcp/9000/ipfs/12D3KooWHHzSeKaY8xuZVzkLbKFfvNgPPeKhFBGrMbNzbm5akpqu} is not routable")))
			})
		})

		It("should return err='no addr to relay' if address timestamp over 60 minutes", func() {
			addrs := []*core.Address{
				{Address: "/ip4/127.0.0.1/tcp/9000/ipfs/12D3KooWHHzSeKaY8xuZVzkLbKFfvNgPPeKhFBGrMbNzbm5akpqu", Timestamp: time.Now().Add(61 * time.Minute).Unix()},
			}
			errs := lp.Gossip().RelayAddresses(addrs)
			Expect(errs).ToNot(BeEmpty())
			Expect(errs).To(ContainElement(fmt.Errorf("no addr to relay")))
		})

		Describe("with 3 address to relay", func() {

			var p, p2, p3 *node.Node

			BeforeEach(func() {
				p = makeTestNode(getPort())
				p2 = makeTestNode(getPort())
				p3 = makeTestNode(getPort())
				lp.PM().AddPeer(rp)
			})

			AfterEach(func() {
				closeNode(p)
				closeNode(p2)
				closeNode(p3)
			})

			It("should successfully select 2 relay peers", func() {
				addrs := []*core.Address{
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

	Describe(".OnAddr", func() {

		var p, p2, p3 *node.Node
		var evt emitter.Event
		var addrMsg *core.Addr

		BeforeEach(func() {
			p = makeTestNode(getPort())
			p2 = makeTestNode(getPort())
			p3 = makeTestNode(getPort())
			addrMsg = &core.Addr{
				Addresses: []*core.Address{
					{Address: p.GetAddress(), Timestamp: time.Now().Unix()},
					{Address: p2.GetAddress(), Timestamp: time.Now().Unix()},
					{Address: p3.GetAddress(), Timestamp: time.Now().Unix()},
				},
			}
		})

		AfterEach(func() {
			closeNode(p)
			closeNode(p2)
			closeNode(p3)
		})

		Describe("when the number of addresses is below max address expected", func() {

			It("should select relay peers from the relayed addresses", func(done Done) {
				wait := make(chan bool)

				stream, c, err := lp.Gossip().NewStream(rp, config.Versions.Addr)
				Expect(err).To(BeNil())

				err = gossip.WriteStream(stream, addrMsg)
				Expect(err).To(BeNil())
				defer c()
				defer stream.Close()

				go func() {
					defer GinkgoRecover()
					evt = <-rp.GetEventEmitter().On(gossip.EventAddressesRelayed)
					Expect(evt.Args).To(BeEmpty())
					Expect(rp.Gossip().GetBroadcasters().Len()).To(Equal(2))
					close(wait)
				}()

				<-wait
				close(done)
			})
		})

		Context("when the number of addresses is above max address expected", func() {
			It("should return no error", func(done Done) {
				wait := make(chan bool)

				rp.GetCfg().Node.MaxAddrsExpected = 1
				stream, c, err := lp.Gossip().NewStream(rp, config.Versions.Addr)

				Expect(err).To(BeNil())
				defer c()
				defer stream.Close()

				err = gossip.WriteStream(stream, addrMsg)
				Expect(err).To(BeNil())

				go func() {
					defer GinkgoRecover()
					evt = <-rp.GetEventEmitter().On(gossip.EventAddrProcessed)
					Expect(evt.Args).ToNot(BeEmpty())
					Expect(evt.Args[0].(error).Error()).To(Equal("too many addresses received. Ignoring addresses"))
					close(wait)
				}()

				<-wait
				close(done)
			})
		})
	})

})
