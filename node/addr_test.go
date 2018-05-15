package node

import (
	"time"

	"github.com/ellcrys/druid/testutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/ellcrys/druid/crypto"
	"github.com/ellcrys/druid/wire"
)

var _ = Describe("Addr", func() {

	BeforeEach(func() {
		var err error
		cfg, err = testutil.SetTestCfg()
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		Expect(testutil.RemoveTestCfgDir()).To(BeNil())
	})

	Describe(".getAddrRelayPeers", func() {

		var err error
		var lp *Node
		var lpProtoc *Inception

		BeforeEach(func() {
			lp, err = NewNode(cfg, "127.0.0.1:30010", crypto.NewKeyFromIntSeed(0), log)
			Expect(err).To(BeNil())
			lpProtoc = NewInception(lp, log)
			lp.SetProtocol(lpProtoc)
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

				peers := lpProtoc.getAddrRelayPeers(candidateAddrs)
				Expect(len(peers)).To(Equal(2))
				Expect(peers[0]).ToNot(BeNil())
				Expect(peers[1]).ToNot(BeNil())
			})

			It("should return a slice of length equal 2; index 0 index containing a *Peer object and 1 contains nil", func() {

				candidateAddrs := []*wire.Address{
					{Address: "/ip4/172.16.238.10/tcp/9000/ipfs/12D3KooWHHzSeKaY8xuZVzkLbKFfvNgPPeKhFBGrMbNzbm5akpqu"},
				}

				peers := lpProtoc.getAddrRelayPeers(candidateAddrs)
				Expect(len(peers)).To(Equal(2))
				Expect(peers[0]).ToNot(BeNil())
				Expect(peers[1]).To(BeNil())
			})
		})

		AfterEach(func() {
			lp.host.Close()
		})
	})

	Describe(".RelayAddr", func() {

		var err error
		var lp *Node
		var lpProtoc *Inception

		BeforeEach(func() {
			lp, err = NewNode(cfg, "127.0.0.1:30010", crypto.NewKeyFromIntSeed(0), log)
			Expect(err).To(BeNil())
			lpProtoc = NewInception(lp, log)
			lp.SetProtocol(lpProtoc)
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
			err := lpProtoc.RelayAddr(addrs)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("too many items in addr message"))
		})

		It("should return err.Error(no addr to relay) if non of the addresses where relayable", func() {
			addrs := []*wire.Address{
				{Address: ""},
				{Address: ""},
			}
			err := lpProtoc.RelayAddr(addrs)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("no addr to relay"))
		})

		It("should return err.Error(no addr to relay) if address timestamp over 60 minutes", func() {
			addrs := []*wire.Address{
				{Address: "/ip4/127.0.0.1/tcp/9000/ipfs/12D3KooWHHzSeKaY8xuZVzkLbKFfvNgPPeKhFBGrMbNzbm5akpqu", Timestamp: time.Now().Add(61 * time.Minute).Unix()},
			}
			err := lpProtoc.RelayAddr(addrs)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("no addr to relay"))
		})

		Context("with relay peers", func() {

			var err error
			var p, p2, p3 *Node
			var pt, pt2, pt3 *Inception

			BeforeEach(func() {
				p, err = NewNode(cfg, "127.0.0.1:30011", crypto.NewKeyFromIntSeed(1), log)
				Expect(err).To(BeNil())
				pt = NewInception(p, log)
				p.SetProtocol(pt)

				p2, err = NewNode(cfg, "127.0.0.1:30012", crypto.NewKeyFromIntSeed(2), log)
				Expect(err).To(BeNil())
				pt2 = NewInception(p2, log)
				p2.SetProtocol(pt2)

				p3, err = NewNode(cfg, "127.0.0.1:30013", crypto.NewKeyFromIntSeed(3), log)
				Expect(err).To(BeNil())
				pt3 = NewInception(p3, log)
				p3.SetProtocol(pt3)
			})

			It("should successfully choose relay peers", func() {
				addrs := []*wire.Address{
					{Address: p2.GetMultiAddr(), Timestamp: time.Now().Unix()},
					{Address: p3.GetMultiAddr(), Timestamp: time.Now().Unix()},
				}
				err := pt.RelayAddr(addrs)
				Expect(err).To(BeNil())
				Expect(pt.addrRelayPeers[0]).ToNot(BeNil())
				Expect(pt.addrRelayPeers[1]).ToNot(BeNil())
			})

			AfterEach(func() {
				p.host.Close()
				p2.host.Close()
				p3.host.Close()
			})
		})

		AfterEach(func() {
			lp.host.Close()
		})
	})
})
