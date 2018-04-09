package peer

import (
	"reflect"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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
			Expect(mgr.KnownPeers()).To(HaveLen(1))
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
				Expect(mgr.KnownPeers()["12D3KooWHHzSeKaY8xuZVzkLbKFfvNgPPeKhFBGrMbNzbm5akpqd"]).NotTo(BeNil())
			})

			It("duplicate peer should be not be recreated", func() {
				Expect(len(mgr.KnownPeers())).To(Equal(2))
				mgr.CreatePeerFromAddress(address)
				Expect(len(mgr.KnownPeers())).To(Equal(2))
			})
		})
	})

	Describe(".isActive", func() {
		It("should return false when Timestamp is zero", func() {
			peer := &Peer{}
			Expect(mgr.isActive(peer)).To(BeFalse())
		})

		It("should return false when Timestamp is 3 hours 10 seconds ago", func() {
			peer := &Peer{
				Timestamp: time.Now().UTC().Add((-3 * (60 * 60) * time.Second) + 10),
			}
			Expect(mgr.isActive(peer)).To(BeFalse())
		})

		It("should return false when Timestamp is 3 hours ago", func() {
			peer := &Peer{
				Timestamp: time.Now().UTC().Add(-3 * (60 * 60) * time.Second),
			}
			Expect(mgr.isActive(peer)).To(BeFalse())
		})

		It("should return true when Timestamp is 2 hours, 59 minute ago", func() {
			peer := &Peer{
				Timestamp: time.Now().UTC().Add((-2 * (60 * 60) * time.Second) + 59*time.Minute),
			}
			Expect(mgr.isActive(peer)).To(BeTrue())
		})
	})

	Describe(".getActivePeers", func() {
		var mgr = new(Manager)
		mgr.knownPeers = make(map[string]*Peer)
		peer1 := &Peer{Timestamp: time.Now().UTC().Add(-1 * (60 * 60) * time.Second)}
		peer2 := &Peer{Timestamp: time.Now().UTC().Add(-2 * (60 * 60) * time.Second)}
		peer3 := &Peer{Timestamp: time.Now().UTC().Add(-3 * (60 * 60) * time.Second)}
		mgr.knownPeers = map[string]*Peer{
			"peer1": peer1,
			"peer2": peer2,
			"peer3": peer3,
		}

		It("should return a map with 2 elements with id peer1 and peer2", func() {
			actual := mgr.getActivePeers()
			Expect(actual).To(HaveLen(2))
			Expect(actual).To(ContainElement(peer1))
			Expect(actual).To(ContainElement(peer2))
		})
	})

	Describe(".GetRandomActivePeers", func() {

		It("should shuffle the slice of peers if the number of known/active peers is equal to the limit requested", func() {
			var mgr = new(Manager)
			mgr.knownPeers = make(map[string]*Peer)
			peer1 := &Peer{Timestamp: time.Now().UTC().Add(-1 * (60 * 60) * time.Second)}
			peer2 := &Peer{Timestamp: time.Now().UTC().Add(-2 * (60 * 60) * time.Second)}
			peer3 := &Peer{Timestamp: time.Now().UTC().Add(-2 * (60 * 60) * time.Second)}
			mgr.knownPeers = map[string]*Peer{
				"peer1": peer1,
				"peer2": peer2,
				"peer3": peer3,
			}

			result, err := mgr.GetRandomActivePeers(3)
			Expect(err).To(BeNil())
			result2, err := mgr.GetRandomActivePeers(3)
			result3, err := mgr.GetRandomActivePeers(3)

			// test position randomness
			if reflect.DeepEqual(result[0], result2[0]) && reflect.DeepEqual(result[0], result3[0]) {
				Fail("failed to shuffle")
			}
		})
	})
})
