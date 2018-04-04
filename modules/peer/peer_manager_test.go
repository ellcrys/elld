package peer_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/ellcrys/gcoin/modules/peer"
)

var _ = Describe("PeerManager", func() {
	var mgr *Manager

	BeforeSuite(func() {
		p, err := NewPeer("127.0.0.1:40000", 0)
		defer p.Host().Close()
		mgr = p.PM()
		Expect(err).To(BeNil())
	})

	Describe(".AddPeer", func() {
		It("return error.Error(nil received as *Peer) if nil is passed as param", func() {
			err := mgr.AddPeer(nil)
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(Equal("nil received as *Peer"))
		})

		It("return nil when peer is successfully added and peers list increases to 1", func() {
			p, err := NewPeer("127.0.0.1:40001", 1)
			defer p.Host().Close()
			err = mgr.AddPeer(p)
			Expect(err).To(BeNil())
			Expect(mgr.Peers()).To(HaveLen(1))
		})
	})

	Describe(".PeerExist", func() {
		It("peer does not exist, must return false", func() {
			p, err := NewPeer("127.0.0.1:40002", 2)
			defer p.Host().Close()
			Expect(err).To(BeNil())
			Expect(mgr.PeerExist(p)).To(BeFalse())
		})

		It("peer exists, must return true", func() {
			p, err := NewPeer("127.0.0.1:40001", 1)
			defer p.Host().Close()
			Expect(err).To(BeNil())
			Expect(mgr.PeerExist(p)).To(BeTrue())
		})
	})

	Describe(".CreatePeerFromAddress", func() {

		Context("with invalid address", func() {
			It("peer return error.Error('failed to create peer from address. Peer address is invalid') when address is /ip4/127.0.0.1/tcp/4000", func() {
				err := mgr.CreatePeerFromAddress("/ip4/127.0.0.1/tcp/4000")
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(Equal("failed to create peer from address. Peer address is invalid"))
			})

			It("peer return error.Error('failed to create peer from address. Peer address is invalid') when address is /ip4/127.0.0.1/tcp/4000/ipfs", func() {
				err := mgr.CreatePeerFromAddress("/ip4/127.0.0.1/tcp/4000/ipfs")
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(Equal("failed to create peer from address. Peer address is invalid"))
			})
		})

		Context("with valid address", func() {

			address := "/ip4/127.0.0.1/tcp/40004/ipfs/12D3KooWHHzSeKaY8xuZVzkLbKFfvNgPPeKhFBGrMbNzbm5akpqd"

			It("peer with address '/ip4/127.0.0.1/tcp/40004/ipfs/12D3KooWHHzSeKaY8xuZVzkLbKFfvNgPPeKhFBGrMbNzbm5akpqd' must be added", func() {
				err := mgr.CreatePeerFromAddress(address)
				Expect(err).To(BeNil())
				Expect(mgr.Peers()["12D3KooWHHzSeKaY8xuZVzkLbKFfvNgPPeKhFBGrMbNzbm5akpqd"]).NotTo(BeNil())
			})

			It("duplicate peer should be not be recreated", func() {
				Expect(len(mgr.Peers())).To(Equal(2))
				mgr.CreatePeerFromAddress(address)
				Expect(len(mgr.Peers())).To(Equal(2))
			})
		})
	})
})
