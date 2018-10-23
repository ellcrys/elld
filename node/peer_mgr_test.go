package node_test

import (
	"context"
	"reflect"
	"time"

	"github.com/imdario/mergo"
	pstore "github.com/libp2p/go-libp2p-peerstore"

	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/node"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/util"
	ma "github.com/multiformats/go-multiaddr"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func emptyNode(n *node.Node) *node.Node {
	n2 := node.NewAlmostEmptyNode()
	if n != nil {
		mergo.MergeWithOverwrite(n2, n)
	}
	return n2
}

func emptyNodeWithLastSeenTime(t time.Time) *node.Node {
	n := node.NewAlmostEmptyNode()
	n.SetLastSeen(t)
	return n
}

func NewMgr(cfg *config.EngineConfig, localNode *node.Node) *node.Manager {
	mgr := node.NewManager(cfg, localNode, log)
	return mgr
}

var _ = Describe("Peer Manager", func() {

	var lp *node.Node
	var mgr *node.Manager
	var cfg *config.EngineConfig
	var lpPort int

	BeforeEach(func() {
		lpPort = getPort()
		lp = makeTestNodeWith(lpPort, 1)
		Expect(lp.GetBlockchain().Up()).To(BeNil())
		mgr = lp.PM()
		mgr.SetLocalNode(lp)
		cfg = lp.GetCfg()
	})

	AfterEach(func() {
		closeNode(lp)
	})

	Describe(".UpdateLastSeen", func() {
		var p2 *node.Node
		var err error

		When("peer does not previously exist", func() {

			BeforeEach(func() {
				p2, err = node.NewNode(cfg, "127.0.0.1:40002", crypto.NewKeyFromIntSeed(0), log)
				defer closeNode(p2)
				err = mgr.UpdateLastSeenTime(p2)
				Expect(err).To(BeNil())
			})

			It("should add peer with time set to an hour ago", func() {
				Expect(mgr.Peers()).To(HaveLen(1))
				Expect(p2.GetLastSeen().IsZero()).To(BeFalse())
				Expect(p2.GetLastSeen().Unix()).To(Equal(time.Now().Add(-1 * time.Hour).Unix()))
			})
		})

		When("peer previously existed and not currently connected to the local node", func() {
			var existingPeer *node.Node
			var err error

			BeforeEach(func() {
				existingPeer, err = node.NewNode(cfg, "127.0.0.1:40002", crypto.NewKeyFromIntSeed(0), log)
				Expect(err).To(BeNil())
				defer closeNode(existingPeer)
				mgr.AddPeer(existingPeer)
				err = mgr.UpdateLastSeenTime(existingPeer)
				Expect(err).To(BeNil())
			})

			It("should update the last seen to one hour ago", func() {
				Expect(existingPeer.GetLastSeen().Unix()).To(Equal(time.Now().Add(-1 * time.Hour).Unix()))
			})
		})

		When("peer previously existed and currently connected to the local node", func() {
			var existingPeer *node.Node
			var err error

			BeforeEach(func() {
				existingPeer, err = node.NewNode(cfg, "127.0.0.1:40002", crypto.NewKeyFromIntSeed(0), log)
				Expect(err).To(BeNil())
				existingPeer.SetLocalNode(lp)

				lp.Host().Peerstore().AddAddr(existingPeer.ID(), existingPeer.GetAddress().DecapIPFS(), pstore.PermanentAddrTTL)
				err := lp.Host().Connect(context.TODO(), lp.Host().Peerstore().PeerInfo(existingPeer.ID()))
				Expect(err).To(BeNil())

				mgr.AddPeer(existingPeer)
			})

			AfterEach(func() {
				closeNode(existingPeer)
			})

			It("should update the last seen to current time", func() {
				now := time.Now().Unix()
				err = mgr.UpdateLastSeenTime(existingPeer)
				Expect(err).To(BeNil())
				Expect(existingPeer.GetLastSeen().Unix()).To(Equal(now))
			})
		})
	})

	Describe(".CleanPeers", func() {

		It("should return 0 when number of connected peers is less than 3", func() {
			n := mgr.CleanPeers()
			Expect(n).To(BeZero())
		})

		When("all peers are active", func() {
			It("should remove 0 peers", func() {
				addr, _ := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/9000/ipfs/12D3KooWM4yJB31d4hF2F9Vdwuj9WFo1qonoySyw4bVAQ9a9d21o")
				p2 := node.NewRemoteNodeFromMultiAddr(addr, lp)
				p2.SetLastSeen(time.Now())
				mgr.Peers()[p2.StringID()] = p2

				addr2, _ := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/9000/ipfs/12D3KooWM4yJB31d4hF2F9Vdwuj9WFo1qonoySyw4bVAQ9a9d21d")
				p3 := node.NewRemoteNodeFromMultiAddr(addr2, lp)
				p3.SetLastSeen(time.Now())
				mgr.Peers()[p3.StringID()] = p3

				n := mgr.CleanPeers()
				Expect(n).To(BeZero())
			})
		})

		When("one peer is not active and one is active", func() {

			var p3 *node.Node

			BeforeEach(func() {
				addr, _ := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/9000/ipfs/12D3KooWM4yJB31d4hF2F9Vdwuj9WFo1qonoySyw4bVAQ9a9d21o")
				p2 := node.NewRemoteNodeFromMultiAddr(addr, lp)
				p2.SetLastSeen(time.Now())
				mgr.Peers()[p2.StringID()] = p2

				addr2, _ := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/9000/ipfs/12D3KooWM4yJB31d4hF2F9Vdwuj9WFo1qonoySyw4bVAQ9a9d21d")
				p3 = node.NewRemoteNodeFromMultiAddr(addr2, lp)
				mgr.Peers()[p3.StringID()] = p3
				mgr.AddAcquainted(p3)
				Expect(mgr.IsAcquainted(p3)).To(BeTrue())
			})

			It("should remove 1 peer", func() {
				n := mgr.CleanPeers()
				Expect(n).To(Equal(1))
			})

			It("should remove the peer from acquainted cache", func() {
				n := mgr.CleanPeers()
				Expect(n).To(Equal(1))
				Expect(mgr.IsAcquainted(p3)).To(BeFalse())
			})
		})
	})

	Describe(".SavePeers", func() {

		When("peer was created less than 20 minutes ago", func() {
			It("should not store peer", func() {
				addr, _ := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/9000/ipfs/12D3KooWM4yJB31d4hF2F9Vdwuj9WFo1qonoySyw4bVAQ9a9d21o")
				rp := node.NewRemoteNodeFromMultiAddr(addr, lp)
				rp.SetCreatedAt(time.Now())

				mgr = lp.PM()
				mgr.Peers()[rp.StringID()] = rp

				err := mgr.SavePeers()
				Expect(err).To(BeNil())

				result := lp.DB().GetByPrefix([]byte("address"))
				Expect(result).To(BeEmpty())
			})
		})

		When("peers were created more than 20 minutes ago", func() {
			It("should successfully store peers", func() {

				addr, _ := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/9000/ipfs/12D3KooWM4yJB31d4hF2F9Vdwuj9WFo1qonoySyw4bVAQ9a9d21o")
				addr2, _ := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/9000/ipfs/12D3KooWM4yJB31d4hF2F9Vdwuj9WFo1qonoySyw4bVAQ9a9d21d")

				p2 := node.NewRemoteNodeFromMultiAddr(addr, lp)
				p2.SetLastSeen(time.Now())
				p2.SetCreatedAt(time.Now().Add(-21 * time.Minute))

				p3 := node.NewRemoteNodeFromMultiAddr(addr2, lp)
				p3.SetLastSeen(time.Now())
				p3.SetCreatedAt(time.Now().Add(-21 * time.Minute))

				mgr.Peers()[p2.StringID()] = p2
				mgr.Peers()[p3.StringID()] = p3

				err := mgr.SavePeers()
				Expect(err).To(BeNil())

				result := lp.DB().GetByPrefix([]byte("address"))
				Expect(result).To(HaveLen(2))
			})
		})

	})

	Describe(".loadPeers", func() {

		var lastSeen time.Time
		var createdAt time.Time
		var peer *node.Node

		BeforeEach(func() {
			createdAt = time.Now().Add(-21 * time.Minute)
			lastSeen = time.Now()
			addr, _ := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/9000/ipfs/12D3KooWM4yJB31d4hF2F9Vdwuj9WFo1qonoySyw4bVAQ9a9d21o")

			peer = node.NewRemoteNodeFromMultiAddr(addr, lp)
			peer.SetCreatedAt(createdAt)
			peer.SetLastSeen(lastSeen)
			mgr.AddPeer(peer)
			Expect(mgr.Peers()).To(HaveLen(1))

			err := mgr.SavePeers()
			Expect(err).To(BeNil())
		})

		It("should fetch 1 address", func() {
			mgr.SetPeers(map[string]types.Engine{})
			err := mgr.LoadPeers()
			Expect(err).To(BeNil())
			Expect(mgr.Peers()).To(HaveKey(peer.StringID()))
			peer := mgr.Peers()[peer.StringID()]
			Expect(peer.GetLastSeen().Unix()).To(Equal(lastSeen.Unix()))
			Expect(peer.CreatedAt().Unix()).To(Equal(createdAt.Unix()))
		})
	})

	Describe(".GetKnownPeer", func() {

		It("should return nil when peer is not in known peer list", func() {
			p2, err := node.NewNode(cfg, "127.0.0.1:40002", crypto.NewKeyFromIntSeed(0), log)
			Expect(err).To(BeNil())
			defer closeNode(p2)
			Expect(mgr.GetPeer(p2.StringID())).To(BeNil())
		})

		It("should return peer when peer is in known peer list", func() {
			p2, err := node.NewNode(cfg, "127.0.0.1:40003", crypto.NewKeyFromIntSeed(3), log)
			defer closeNode(p2)
			mgr.UpdateLastSeenTime(p2)
			Expect(err).To(BeNil())
			actual := mgr.GetPeer(p2.StringID())
			Expect(actual).NotTo(BeNil())
			Expect(actual).To(Equal(p2))
		})
	})

	Describe(".PeerExist", func() {

		It("peer does not exist, must return false", func() {
			p2, err := node.NewNode(cfg, "127.0.0.1:40002", crypto.NewKeyFromIntSeed(0), log)
			defer closeNode(p2)
			Expect(err).To(BeNil())
			Expect(mgr.PeerExist(lp.StringID())).To(BeFalse())
			closeNode(p2)
		})

		It("peer exists, must return true", func() {
			p2, err := node.NewNode(cfg, "127.0.0.1:40002", crypto.NewKeyFromIntSeed(0), log)
			defer closeNode(p2)
			mgr.UpdateLastSeenTime(p2)
			Expect(err).To(BeNil())
			Expect(mgr.PeerExist(p2.StringID())).To(BeTrue())
			closeNode(p2)
		})
	})

	Describe(".GetKnownPeers", func() {
		It("should return peer1 as the only known peer", func() {
			p2, err := node.NewNode(cfg, "127.0.0.1:40002", crypto.NewKeyFromIntSeed(0), log)
			Expect(err).To(BeNil())
			defer closeNode(p2)
			mgr.UpdateLastSeenTime(p2)
			actual := mgr.GetPeers()
			Expect(actual).To(HaveLen(1))
			Expect(actual).To(ContainElement(p2))
		})

	})

	Describe(".IsActive", func() {

		It("should return false when Timestamp is zero", func() {
			peer := emptyNode(nil)
			Expect(mgr.IsActive(peer)).To(BeFalse())
		})

		It("should return false when Timestamp is 3 hours 10 seconds ago", func() {
			peer := emptyNode(nil)
			peer.SetLastSeen(time.Now().Add((-3 * (60 * 60) * time.Second) + 10))
			Expect(mgr.IsActive(peer)).To(BeFalse())
		})

		It("should return false when Timestamp is 3 hours ago", func() {
			peer := emptyNode(nil)
			peer.SetLastSeen(time.Now().Add(-3 * (60 * 60) * time.Second))
			Expect(mgr.IsActive(peer)).To(BeFalse())
		})

		It("should return true when Timestamp is 2 hours, 59 minute ago", func() {
			peer := emptyNode(nil)
			peer.SetLastSeen(time.Now().Add((-2 * (60 * 60) * time.Second) + 59*time.Minute))
			Expect(mgr.IsActive(peer)).To(BeTrue())
		})

	})

	Describe(".RandomPeers", func() {

		var peer1, peer2, peer3 *node.Node
		var mgr *node.Manager

		BeforeEach(func() {
			mgr = lp.PM()
			mgr.SetLocalNode(lp)
			peer1 = emptyNodeWithLastSeenTime(time.Now())
			peer2 = emptyNodeWithLastSeenTime(time.Now())
			peer3 = emptyNodeWithLastSeenTime(time.Now())
			mgr.SetPeers(map[string]types.Engine{
				"peer1": peer1,
				"peer2": peer2,
				"peer3": peer3,
			})
		})

		It("", func() {

		})
	})

	Describe(".GetActivePeers", func() {

		var peer1, peer2, peer3 *node.Node
		var mgr *node.Manager

		BeforeEach(func() {
			mgr = lp.PM()
			mgr.SetLocalNode(lp)
			peer1 = emptyNodeWithLastSeenTime(time.Now().Add(-1 * (60 * 60) * time.Second))
			peer2 = emptyNodeWithLastSeenTime(time.Now().Add(-2 * (60 * 60) * time.Second))
			peer3 = emptyNodeWithLastSeenTime(time.Now().Add(-3 * (60 * 60) * time.Second))
			mgr.SetPeers(map[string]types.Engine{
				"peer1": peer1,
				"peer2": peer2,
				"peer3": peer3,
			})
		})

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
			Expect([]*node.Node{peer1, peer2}).To(ContainElement(actual[0]))
		})

	})

	Describe(".CopyActivePeers", func() {

		var mgr *node.Manager

		BeforeEach(func() {
			mgr = NewMgr(cfg, lp)
			peer1 := emptyNodeWithLastSeenTime(time.Now().Add(-1 * (60 * 60) * time.Second))
			mgr.SetPeers(map[string]types.Engine{
				"peer1": peer1,
			})
		})

		It("should return a different slice from the original knownPeer slice", func() {
			actual := mgr.CopyActivePeers(1)
			Expect(actual).To(HaveLen(1))
			Expect(actual).NotTo(Equal(mgr.Peers))
		})
	})

	Describe(".GetRandomActivePeers", func() {

		It("should shuffle the slice of peers if the number of known/active peers is equal to the limit requested", func() {
			peer1 := emptyNodeWithLastSeenTime(time.Now().Add(-1 * (60 * 60) * time.Second))
			peer2 := emptyNodeWithLastSeenTime(time.Now().Add(-2 * (60 * 60) * time.Second))
			peer3 := emptyNodeWithLastSeenTime(time.Now().Add(-2 * (60 * 60) * time.Second))
			mgr.SetPeers(map[string]types.Engine{
				"peer1": peer1,
				"peer2": peer2,
				"peer3": peer3,
			})

			var result []types.Engine
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
			peer1 := emptyNodeWithLastSeenTime(time.Now().Add(-1 * (60 * 60) * time.Second))
			peer2 := emptyNodeWithLastSeenTime(time.Now().Add(-2 * (60 * 60) * time.Second))
			peer3 := emptyNodeWithLastSeenTime(time.Now().Add(-2 * (60 * 60) * time.Second))
			mgr.SetPeers(map[string]types.Engine{
				"peer1": peer1,
				"peer2": peer2,
				"peer3": peer3,
			})

			var result []types.Engine
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

	Describe(".RequirePeers", func() {

		Context("when the node does not have up to 1000 active addresses", func() {
			Context("when max outbound connection has not been reached", func() {
				It("should return true", func() {
					cfg.Node.MaxOutboundConnections = 120
					peer1 := emptyNodeWithLastSeenTime(time.Now().Add(-1 * (60 * 60) * time.Second))
					mgr.SetPeers(map[string]types.Engine{"peer1": peer1})
					Expect(mgr.RequirePeers()).To(BeTrue())
				})
			})

			Context("when max outbound connection has been reached", func() {
				It("should return false", func() {
					cfg.Node.MaxOutboundConnections = 10
					mgr.ConnMgr().SetConnsInfo(node.NewConnsInfo(0, 10))
					peer1 := emptyNodeWithLastSeenTime(time.Now().Add(-1 * (60 * 60) * time.Second))
					mgr.SetPeers(map[string]types.Engine{"peer1": peer1})
					Expect(mgr.RequirePeers()).To(BeFalse())
				})
			})

		})
	})

	Describe(".HasDisconnected", func() {

		Context("when address does not contain a known peer id", func() {
			It("should return err='unknown peer'", func() {
				addr := "/ip4/127.0.0.1/tcp/9000/ipfs/12D3KooWM4yJB31d4hF2F9Vdwuj9WFo1qonoySyw4bVAQ9a9d212"
				err := mgr.HasDisconnected(util.NodeAddr(addr))
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("unknown peer"))
			})
		})

		Context("when address contain a known peer id", func() {
			It("should set 'last seen' time to 1 hour ago", func() {
				lp.SetLastSeen(time.Now())
				mgr.SetPeers(map[string]types.Engine{lp.GetAddress().StringID(): lp})
				currentTime := lp.GetLastSeen().Unix()
				err := mgr.HasDisconnected(lp.GetAddress())
				Expect(err).To(BeNil())
				actual := lp.GetLastSeen().Unix()
				Expect(currentTime > actual).To(BeTrue())
				Expect(currentTime - actual).To(Equal(int64(3600)))
			})
		})

	})

	Describe(".IsLocalPeer", func() {

		It("should return false if nil is passed", func() {
			Expect(mgr.IsLocalNode(nil)).To(BeFalse())
		})

		It("should return false if local peer is nil", func() {
			mgr.SetLocalNode(nil)
			peer1 := emptyNodeWithLastSeenTime(time.Now())
			Expect(mgr.IsLocalNode(peer1)).To(BeFalse())
		})

		It("should return false if not local peer", func() {
			peer1, err := node.NewNode(cfg, "127.0.0.1:40002", crypto.NewKeyFromIntSeed(0), log)
			Expect(err).To(BeNil())
			defer closeNode(peer1)
			Expect(mgr.IsLocalNode(peer1)).To(BeFalse())
		})

		It("should return true if peer is the local peer", func() {
			peer1, err := node.NewNode(cfg, "127.0.0.1:40002", crypto.NewKeyFromIntSeed(0), log)
			Expect(err).To(BeNil())
			defer closeNode(peer1)
			mgr.SetLocalNode(peer1)
			Expect(mgr.IsLocalNode(peer1)).To(BeTrue())
		})
	})

	Describe(".establishConnection", func() {
		It("should return nil when peer does not exist", func() {
			var mgr = NewMgr(cfg, lp)
			err := mgr.ConnectToPeer("invalid")
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("peer not found"))
		})
	})

})
