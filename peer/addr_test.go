package peer

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/ellcrys/druid/configdir"
	"github.com/ellcrys/druid/util"
	"github.com/ellcrys/druid/wire"
)

var _ = Describe("Addr", func() {

	var config = &configdir.Config{
		Peer: &configdir.PeerConfig{
			Dev:              true,
			MaxAddrsExpected: 5,
		},
	}

	Describe(".getAddrRelayPeers", func() {

		var err error
		var lp *Peer
		var lpProtoc *Inception

		BeforeEach(func() {
			lp, err = NewPeer(config, "127.0.0.1:30010", 0, util.NewNopLogger())
			Expect(err).To(BeNil())
			lpProtoc = NewInception(lp, util.NewNopLogger())
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
})
