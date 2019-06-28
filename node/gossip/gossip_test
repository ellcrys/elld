package gossip_test

import (
	"context"

	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/node"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/util"
	net "github.com/libp2p/go-libp2p-net"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Gossip", func() {

	var lp, rp *node.Node

	BeforeEach(func() {
		lp = makeTestNode(getPort())
		rp = makeTestNode(getPort())
	})

	AfterEach(func() {
		closeNode(lp)
		closeNode(rp)
	})

	Describe(".checkRemotePeer", func() {

		var stream net.Stream
		var ws *core.WrappedStream
		var cc context.CancelFunc
		var err error

		When("protocol id is Handshake", func() {

			BeforeEach(func() {
				stream, cc, err = lp.Gossip().NewStream(rp, config.GetVersions().Handshake)
				Expect(err).To(BeNil())
				defer cc()
				defer stream.Close()
				ws = &core.WrappedStream{Stream: stream, Extra: map[string]interface{}{}}
			})

			It("should return nil", func() {
				err := lp.Gossip().CheckRemotePeer(ws, rp)
				Expect(err).To(BeNil())
			})
		})

		When("not in test mode and remote peer is unacquainted and message is not Addr", func() {

			BeforeEach(func() {
				lp.GetCfg().Node.Mode = config.ModeProd
			})

			BeforeEach(func() {
				stream, cc, err = lp.Gossip().NewStream(rp, config.GetVersions().GetAddr)
				Expect(err).To(BeNil())
				defer cc()
				defer stream.Close()
				ws = &core.WrappedStream{Stream: stream, Extra: map[string]interface{}{}}
			})

			It("should return err='unacquainted node'", func() {
				err := lp.Gossip().CheckRemotePeer(ws, rp)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("unacquainted node"))
			})
		})

		When("not in test mode and remote peer is unacquainted and message is Addr", func() {

			BeforeEach(func() {
				lp.GetCfg().Node.Mode = config.ModeProd
			})

			BeforeEach(func() {
				stream, cc, err = lp.Gossip().NewStream(rp, config.GetVersions().Addr)
				Expect(err).To(BeNil())
				defer cc()
				defer stream.Close()
				ws = &core.WrappedStream{Stream: stream, Extra: map[string]interface{}{}}
			})

			It("should return nil", func() {
				err := lp.Gossip().CheckRemotePeer(ws, rp)
				Expect(err).To(BeNil())
			})
		})
	})

	Describe(".PickBroadcasters", func() {

		var candidateAddrs []*core.Address
		var cache *core.BroadcastPeers

		BeforeEach(func() {
			cache = core.NewBroadcastPeers()
			candidateAddrs = []*core.Address{
				{Address: "/ip4/172.16.238.10/tcp/9000/ipfs/12D3KooWHHzSeKaY8xuZVzkLbKFfvNgPPeKhFBGrMbNzbm5akpqu"},
				{Address: "/ip4/172.16.238.11/tcp/9000/ipfs/12D3KooWB1b3qZxWJanuhtseF3DmPggHCtG36KZ9ixkqHtdKH9fh"},
				{Address: "/ip4/172.16.238.12/tcp/9000/ipfs/12D3KooWPgam4TzSVCRa4AbhxQnM9abCYR4E9hV57SN7eAjEYn1j"},
				{Address: "/ip4/172.16.238.13/tcp/9000/ipfs/12D3KooWKRyzVWW6ChFjQjK4miCty85Niy49tpPV95XdKu1BcvMA"},
				{Address: "/ip4/172.16.238.14/tcp/9000/ipfs/12D3KooWE4qDcRrueTuRYWUdQZgcy7APZqBngVeXRt4Y6ytHizKV"},
			}
		})

		Describe("cache contains no address", func() {
			Describe("many candidate addresses is passed", func() {
				It("should return N nodes", func() {
					broadcasters := lp.Gossip().PickBroadcasters(cache, candidateAddrs, 2)
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
					broadcasters := lp.Gossip().PickBroadcasters(cache, candidateAddrs, 2)
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
				broadcasters = lp.Gossip().PickBroadcasters(cache, candidateAddrs, 2)
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
				cache.Add(n)
			})

			It("should leave the cache untouched and add the only candidate address", func() {
				broadcasters := lp.Gossip().PickBroadcasters(cache, candidateAddrs, 2)
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
				broadcasters = lp.Gossip().PickBroadcasters(cache, candidateAddrs, 2)
				Expect(broadcasters.Len()).To(Equal(2))
			})

			It("should return current cache values", func() {
				broadcasters2 := lp.Gossip().PickBroadcasters(cache, candidateAddrs2, 2)
				Expect(broadcasters2).To(Equal(broadcasters))
			})
		})
	})

})
