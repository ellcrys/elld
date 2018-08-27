package node

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/wire"
)

func AddrTest() bool {
	return Describe("Addr", func() {
		Describe(".getAddrRelayPeers", func() {

			var err error
			var lp *Node
			var gossip *Gossip

			BeforeEach(func() {
				lp, err = NewNode(cfg, "127.0.0.1:30010", crypto.NewKeyFromIntSeed(0), log)
				Expect(err).To(BeNil())
				gossip = NewGossip(lp, log)
				lp.SetGossipProtocol(gossip)
			})

			When("no relay peer have been stored", func() {

				It("should return a slice of length equal 2 and each index containing *Peer objects", func() {

					candidateAddrs := []*wire.Address{
						{Address: "/ip4/172.16.238.10/tcp/9000/ipfs/12D3KooWHHzSeKaY8xuZVzkLbKFfvNgPPeKhFBGrMbNzbm5akpqu"},
						{Address: "/ip4/172.16.238.11/tcp/9000/ipfs/12D3KooWB1b3qZxWJanuhtseF3DmPggHCtG36KZ9ixkqHtdKH9fh"},
						{Address: "/ip4/172.16.238.12/tcp/9000/ipfs/12D3KooWPgam4TzSVCRa4AbhxQnM9abCYR4E9hV57SN7eAjEYn1j"},
						{Address: "/ip4/172.16.238.13/tcp/9000/ipfs/12D3KooWKRyzVWW6ChFjQjK4miCty85Niy49tpPV95XdKu1BcvMA"},
						{Address: "/ip4/172.16.238.14/tcp/9000/ipfs/12D3KooWE4qDcRrueTuRYWUdQZgcy7APZqBngVeXRt4Y6ytHizKV"},
					}

					peers := gossip.selectPeersToRelayTo(candidateAddrs)
					Expect(len(peers)).To(Equal(2))
					Expect(peers[0]).ToNot(BeNil())
					Expect(peers[1]).ToNot(BeNil())
				})

				It("should return a slice of length equal 2; index 0 index containing a *Peer object and 1 contains nil", func() {

					candidateAddrs := []*wire.Address{
						{Address: "/ip4/172.16.238.10/tcp/9000/ipfs/12D3KooWHHzSeKaY8xuZVzkLbKFfvNgPPeKhFBGrMbNzbm5akpqu"},
					}

					peers := gossip.selectPeersToRelayTo(candidateAddrs)
					Expect(len(peers)).To(Equal(2))
					Expect(peers[0]).ToNot(BeNil())
					Expect(peers[1]).To(BeNil())
				})
			})

			AfterEach(func() {
				closeNode(lp)
			})
		})

		Describe(".RelayAddr", func() {

			var err error
			var lp *Node
			var gossip *Gossip

			BeforeEach(func() {
				lp, err = NewNode(cfg, "127.0.0.1:30010", crypto.NewKeyFromIntSeed(0), log)
				Expect(err).To(BeNil())
				gossip = NewGossip(lp, log)
				lp.SetGossipProtocol(gossip)
			})

			It("should return err.Error(too many items in addr message) when address is more than 10", func() {
				addrs := []*wire.Address{
					{Address: ""},
					{Address: ""},
					{Address: ""},
					{Address: ""},
					{Address: ""},
					{Address: ""},
					{Address: ""},
					{Address: ""},
					{Address: ""},
					{Address: ""},
					{Address: ""},
				}
				errs := gossip.RelayAddr(addrs)
				Expect(errs).ToNot(BeEmpty())
				Expect(errs).To(ContainElement(fmt.Errorf("too many addresses in the message")))
			})

			It("should return err.Error(no addr to relay) if non of the addresses where relayable", func() {
				addrs := []*wire.Address{
					{Address: ""},
					{Address: ""},
				}
				errs := gossip.RelayAddr(addrs)
				Expect(errs).ToNot(BeEmpty())
				Expect(errs).To(ContainElement(fmt.Errorf("no addr to relay")))
			})

			It("should return err.Error(no addr to relay) if address timestamp over 60 minutes", func() {
				addrs := []*wire.Address{
					{Address: "/ip4/127.0.0.1/tcp/9000/ipfs/12D3KooWHHzSeKaY8xuZVzkLbKFfvNgPPeKhFBGrMbNzbm5akpqu", Timestamp: time.Now().Add(61 * time.Minute).Unix()},
				}
				errs := gossip.RelayAddr(addrs)
				Expect(errs).ToNot(BeEmpty())
				Expect(errs).To(ContainElement(fmt.Errorf("no addr to relay")))
			})

			Context("with relay peers", func() {

				var err error
				var p, p2, p3 *Node
				var pt, pt2, pt3 *Gossip

				BeforeEach(func() {
					p, err = NewNode(cfg, "127.0.0.1:30011", crypto.NewKeyFromIntSeed(1), log)
					Expect(err).To(BeNil())
					pt = NewGossip(p, log)
					p.SetGossipProtocol(pt)

					p2, err = NewNode(cfg, "127.0.0.1:30012", crypto.NewKeyFromIntSeed(2), log)
					Expect(err).To(BeNil())
					pt2 = NewGossip(p2, log)
					p2.SetGossipProtocol(pt2)

					p3, err = NewNode(cfg, "127.0.0.1:30013", crypto.NewKeyFromIntSeed(3), log)
					Expect(err).To(BeNil())
					pt3 = NewGossip(p3, log)
					p3.SetGossipProtocol(pt3)
				})

				It("should successfully choose relay peers", func() {
					addrs := []*wire.Address{
						{Address: p2.GetMultiAddr(), Timestamp: time.Now().Unix()},
						{Address: p3.GetMultiAddr(), Timestamp: time.Now().Unix()},
					}
					pt.RelayAddr(addrs)
					Expect(pt.addrRelayPeers[0]).ToNot(BeNil())
					Expect(pt.addrRelayPeers[1]).ToNot(BeNil())
				})

				AfterEach(func() {
					closeNode(p)
					closeNode(p2)
					closeNode(p3)
				})
			})

			AfterEach(func() {
				closeNode(lp)
			})
		})
	})
}
