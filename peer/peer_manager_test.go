package peer

import (
	"reflect"
	"sync"
	"time"

	"github.com/ellcrys/druid/configdir"

	"github.com/ellcrys/druid/util"
	ma "github.com/multiformats/go-multiaddr"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func NewMgr(config *configdir.Config) *Manager {
	var mgr = new(Manager)
	mgr.kpm = &sync.Mutex{}
	mgr.gm = &sync.Mutex{}
	mgr.config = config
	mgr.knownPeers = make(map[string]*Peer)
	mgr.log = util.NewNopLogger()
	return mgr
}

var _ = Describe("PeerManager", func() {
	var mgr *Manager

	var config = &configdir.Config{
		Peer: &configdir.PeerConfig{
			Dev: true,
		},
	}

	BeforeSuite(func() {
		p, err := NewPeer(config, "127.0.0.1:40000", 0)
		defer p.Host().Close()
		mgr = p.PM()
		Expect(err).To(BeNil())
	})

	Describe(".AddPeer", func() {
		It("return error.Error(nil received as *Peer) if nil is passed as param", func() {
			err := mgr.AddOrUpdatePeer(nil)
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(Equal("nil received as *Peer"))
		})

		It("return nil when peer is successfully added and peers list increases to 1", func() {
			p, err := NewPeer(config, "127.0.0.1:40001", 1)
			defer p.Host().Close()
			err = mgr.AddOrUpdatePeer(p)
			Expect(err).To(BeNil())
			Expect(mgr.KnownPeers()).To(HaveLen(1))
		})
	})

	Describe(".GetKnownPeer", func() {
		It("should return nil when peer is not in known peer list", func() {
			mgr := NewMgr(config)
			p, err := NewPeer(config, "127.0.0.1:40001", 1)
			Expect(err).To(BeNil())
			Expect(mgr.GetKnownPeer(p.StringID())).To(BeNil())
		})

		It("should return peer when peer is in known peer list", func() {
			mgr := NewMgr(config)
			p, err := NewPeer(config, "127.0.0.1:40001", 1)
			defer p.host.Close()
			mgr.AddOrUpdatePeer(p)
			Expect(err).To(BeNil())
			actual := mgr.GetKnownPeer(p.StringID())
			Expect(actual).NotTo(BeNil())
			Expect(actual).To(Equal(p))
		})
	})

	Describe(".PeerExist", func() {
		It("peer does not exist, must return false", func() {
			p, err := NewPeer(config, "127.0.0.1:40002", 2)
			defer p.Host().Close()
			Expect(err).To(BeNil())
			Expect(mgr.PeerExist(p.StringID())).To(BeFalse())
		})

		It("peer exists, must return true", func() {
			p, err := NewPeer(config, "127.0.0.1:40001", 1)
			defer p.Host().Close()
			Expect(err).To(BeNil())
			Expect(mgr.PeerExist(p.StringID())).To(BeTrue())
		})
	})

	Describe(".onPeerConnect", func() {
		It("should increment active connection count", func() {
			mgr := NewMgr(config)
			p, err := NewPeer(config, "127.0.0.1:40001", 1)
			Expect(err).To(BeNil())
			defer p.host.Close()
			addr, _ := ma.NewMultiaddr(p.GetMultiAddr())
			mgr.onPeerConnect(addr)
			Expect(mgr.activeConnections).To(Equal(1))
		})
	})

	Describe(".onPeerDisconnet", func() {

		It("should set the disconnected peer's timestamp to at least 2 hour ago", func() {
			mgr := NewMgr(config)
			mgr.activeConnections = 1
			p, err := NewPeer(config, "127.0.0.1:40001", 1)
			Expect(err).To(BeNil())
			defer p.host.Close()
			mgr.AddOrUpdatePeer(p)
			currentTimestamp := p.Timestamp.Unix()

			addr, _ := ma.NewMultiaddr(p.GetMultiAddr())
			mgr.onPeerDisconnect(addr)
			newTimestamp := p.Timestamp.Unix()
			Expect(newTimestamp).ToNot(Equal(currentTimestamp))
			if newTimestamp >= currentTimestamp {
				Fail("newTimestamp must be lesser")
			}
			Expect(currentTimestamp - newTimestamp).To(Equal(int64(7200)))
		})

		It("should increment decrement connection count", func() {
			mgr := NewMgr(config)
			mgr.activeConnections = 3
			p, err := NewPeer(config, "127.0.0.1:40001", 1)
			Expect(err).To(BeNil())
			defer p.host.Close()
			addr, _ := ma.NewMultiaddr(p.GetMultiAddr())
			mgr.onPeerDisconnect(addr)
			Expect(mgr.activeConnections).To(Equal(2))
		})
	})

	Describe(".GetKnownPeers", func() {
		It("should return peer1 as the only known peer", func() {
			mgr := NewMgr(config)
			peer1, err := NewPeer(config, "127.0.0.1:40001", 1)
			Expect(err).To(BeNil())
			mgr.AddOrUpdatePeer(peer1)
			actual := mgr.GetKnownPeers()
			Expect(actual).To(HaveLen(1))
			Expect(actual).To(ContainElement(peer1))
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

	Describe(".GetActivePeers", func() {
		var mgr = NewMgr(config)
		mgr.knownPeers = make(map[string]*Peer)
		peer1 := &Peer{Timestamp: time.Now().UTC().Add(-1 * (60 * 60) * time.Second)}
		peer2 := &Peer{Timestamp: time.Now().UTC().Add(-2 * (60 * 60) * time.Second)}
		peer3 := &Peer{Timestamp: time.Now().UTC().Add(-3 * (60 * 60) * time.Second)}
		mgr.knownPeers = map[string]*Peer{
			"peer1": peer1,
			"peer2": peer2,
			"peer3": peer3,
		}

		It("should return a map with 2 elements with id peer1 and peer2 when limit is set to 0", func() {
			actual := mgr.GetActivePeers(0)
			Expect(actual).To(HaveLen(2))
			Expect(actual).To(ContainElement(peer1))
			Expect(actual).To(ContainElement(peer2))
		})

		It("should return a map with 2 elements with id peer1 and peer2 when limit is set to a negative number", func() {
			actual := mgr.GetActivePeers(-1)
			Expect(actual).To(HaveLen(2))
			Expect(actual).To(ContainElement(peer1))
			Expect(actual).To(ContainElement(peer2))
		})

		It("should return a map with 1 elements with either peer1 or peer2 when limit is set to 1", func() {
			actual := mgr.GetActivePeers(1)
			Expect(actual).To(HaveLen(1))
			Expect([]*Peer{peer1, peer2}).To(ContainElement(actual[0]))
		})
	})

	Describe(".CopyActivePeers", func() {
		var mgr = NewMgr(config)
		mgr.knownPeers = make(map[string]*Peer)
		peer1 := &Peer{Timestamp: time.Now().UTC().Add(-1 * (60 * 60) * time.Second)}
		mgr.knownPeers = map[string]*Peer{
			"peer1": peer1,
		}

		It("should return a different slice from the original knownPeer slice", func() {
			actual := mgr.CopyActivePeers(1)
			Expect(actual).To(HaveLen(1))
			Expect(actual).NotTo(Equal(mgr.knownPeers))
		})
	})

	Describe(".GetRandomActivePeers", func() {

		It("should shuffle the slice of peers if the number of known/active peers is equal to the limit requested", func() {
			var mgr = NewMgr(config)
			mgr.knownPeers = make(map[string]*Peer)
			peer1 := &Peer{Timestamp: time.Now().UTC().Add(-1 * (60 * 60) * time.Second)}
			peer2 := &Peer{Timestamp: time.Now().UTC().Add(-2 * (60 * 60) * time.Second)}
			peer3 := &Peer{Timestamp: time.Now().UTC().Add(-2 * (60 * 60) * time.Second)}
			mgr.knownPeers = map[string]*Peer{
				"peer1": peer1,
				"peer2": peer2,
				"peer3": peer3,
			}

			var result []*Peer
			var sameIndexCount = 0
			result = mgr.GetRandomActivePeers(3)

			// test for randomness
			for i := 0; i < 10; i++ {
				result2 := mgr.GetRandomActivePeers(3)
				if reflect.DeepEqual(result[0], result2[0]) {
					sameIndexCount++
				}
			}

			Expect(sameIndexCount).NotTo(Equal(10))
		})

		It("Should return the limit requested if known active peers are more than limit", func() {
			var mgr = NewMgr(config)
			mgr.knownPeers = make(map[string]*Peer)
			peer1 := &Peer{Timestamp: time.Now().UTC().Add(-1 * (60 * 60) * time.Second)}
			peer2 := &Peer{Timestamp: time.Now().UTC().Add(-2 * (60 * 60) * time.Second)}
			peer3 := &Peer{Timestamp: time.Now().UTC().Add(-2 * (60 * 60) * time.Second)}
			mgr.knownPeers = map[string]*Peer{
				"peer1": peer1,
				"peer2": peer2,
				"peer3": peer3,
			}

			var result []*Peer
			var sameIndexCount = 0
			result = mgr.GetRandomActivePeers(2)

			// test for randomness
			for i := 0; i < 10; i++ {
				result2 := mgr.GetRandomActivePeers(3)
				if reflect.DeepEqual(result[0], result2[0]) {
					sameIndexCount++
				}
			}

			Expect(sameIndexCount).NotTo(Equal(10))
		})
	})

	Describe(".NeedMorePeers", func() {
		It("should return true if peer manager does not have upto 1000 peers", func() {
			var mgr = NewMgr(config)
			mgr.knownPeers = make(map[string]*Peer)
			peer1 := &Peer{Timestamp: time.Now().UTC().Add(-1 * (60 * 60) * time.Second)}
			mgr.knownPeers = map[string]*Peer{
				"peer1": peer1,
			}
			Expect(mgr.NeedMorePeers()).To(BeTrue())
		})
	})

	Describe(".TimestampPunishment", func() {
		It("return err.Error('nil passed') when nil is passed as peer", func() {
			var mgr = NewMgr(config)
			err := mgr.TimestampPunishment(nil)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("nil passed"))
		})

		It("reduce timestamp by an 3600 seconds", func() {
			var mgr = NewMgr(config)
			mgr.knownPeers = make(map[string]*Peer)
			peer1 := &Peer{Timestamp: time.Now().UTC()}
			mgr.knownPeers = map[string]*Peer{
				"peer1": peer1,
			}
			currentTime := peer1.Timestamp.Unix()
			err := mgr.TimestampPunishment(peer1)
			Expect(err).To(BeNil())
			actual := peer1.Timestamp.Unix()
			Expect(currentTime > actual).To(BeTrue())
			Expect(currentTime - actual).To(Equal(int64(3600)))
		})
	})

	Describe(".IsLocalPeer", func() {
		It("should return false if nil is passed", func() {
			var mgr = NewMgr(config)
			Expect(mgr.IsLocalPeer(nil)).To(BeFalse())
		})

		It("should return false if local peer is nil", func() {
			var mgr = NewMgr(config)
			mgr.localPeer = nil
			peer1 := &Peer{Timestamp: time.Now().UTC()}
			Expect(mgr.IsLocalPeer(peer1)).To(BeFalse())
		})

		It("should return false if not local peer", func() {
			var mgr = NewMgr(config)
			peer1, err := NewPeer(config, "127.0.0.1:40010", 1)
			Expect(err).To(BeNil())
			defer peer1.host.Close()
			mgr.localPeer = peer1
			peer2, err := NewPeer(config, "127.0.0.1:40011", 2)
			defer peer2.host.Close()
			Expect(err).To(BeNil())
			Expect(mgr.IsLocalPeer(peer2)).To(BeFalse())
		})

		It("should return true if peer is the local peer", func() {
			var mgr = NewMgr(config)
			peer1, err := NewPeer(config, "127.0.0.1:40010", 1)
			Expect(err).To(BeNil())
			defer peer1.host.Close()
			mgr.localPeer = peer1
			Expect(mgr.IsLocalPeer(peer1)).To(BeTrue())
		})
	})
})
