package node_test

import (
	"context"
	"reflect"
	"time"

	"github.com/imdario/mergo"

	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/node"
	"github.com/ellcrys/elld/testutil"
	"github.com/ellcrys/elld/types"
	host "github.com/libp2p/go-libp2p-host"
	pstore "github.com/libp2p/go-libp2p-peerstore"
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

func NewMgr(cfg *config.EngineConfig, localNode *node.Node) *node.Manager {
	mgr := node.NewManager(cfg, localNode, log)
	return mgr
}

var _ = Describe("GetAddr", func() {

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

	Describe(".AddOrUpdatePeer", func() {

		It("return err='nil received' when nil is passed", func() {
			err := mgr.AddOrUpdatePeer(nil)
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(Equal("nil received"))
		})

		It("return nil when peer is successfully added and peers list increases to 1", func() {
			p2, err := node.NewNode(cfg, "127.0.0.1:40002", crypto.NewKeyFromIntSeed(0), log)
			defer closeNode(p2)
			err = mgr.AddOrUpdatePeer(p2)
			Expect(err).To(BeNil())
			Expect(mgr.KnownPeers()).To(HaveLen(1))
		})

		It("when peer exist but has a different address, return error", func() {
			p2, err := node.NewNode(cfg, "127.0.0.1:40003", crypto.NewKeyFromIntSeed(3), log)
			Expect(err).To(BeNil())
			defer closeNode(p2)

			err = mgr.AddOrUpdatePeer(p2)
			Expect(err).To(BeNil())

			p3, err := node.NewNode(cfg, "127.0.0.1:40004", crypto.NewKeyFromIntSeed(3), log)
			Expect(err).To(BeNil())
			defer closeNode(p2)

			err = mgr.AddOrUpdatePeer(p3)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("existing peer address do not match"))
		})

		It("should return err='peer is the local peer' when the local peer is passed", func() {
			err := mgr.AddOrUpdatePeer(lp)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("peer is the local peer"))
		})
	})

	Describe(".CleanKnownPeers", func() {

		It("should return 0 when number of connected peers is less than 3", func() {
			n := mgr.CleanKnownPeers()
			Expect(n).To(BeZero())
		})

		It("should return 0 when no peer was removed", func() {
			mgr.SetNumActiveConnections(3)
			addr, _ := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/9000/ipfs/12D3KooWM4yJB31d4hF2F9Vdwuj9WFo1qonoySyw4bVAQ9a9d21o")

			p2 := node.NewRemoteNode(addr, lp)
			p2.Timestamp = time.Now()
			addr2, _ := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/9000/ipfs/12D3KooWM4yJB31d4hF2F9Vdwuj9WFo1qonoySyw4bVAQ9a9d21d")

			p3 := node.NewRemoteNode(addr2, lp)
			p3.Timestamp = time.Now()
			mgr.KnownPeers()[p2.StringID()] = p2
			mgr.KnownPeers()[p3.StringID()] = p3

			n := mgr.CleanKnownPeers()
			Expect(n).To(BeZero())
		})

		It("should return 1 when a peer was removed", func() {
			mgr.SetNumActiveConnections(3)
			addr, _ := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/9000/ipfs/12D3KooWM4yJB31d4hF2F9Vdwuj9WFo1qonoySyw4bVAQ9a9d21o")

			p2 := node.NewRemoteNode(addr, lp)
			p2.Timestamp = time.Now()
			addr2, _ := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/9000/ipfs/12D3KooWM4yJB31d4hF2F9Vdwuj9WFo1qonoySyw4bVAQ9a9d21d")

			p3 := node.NewRemoteNode(addr2, lp)
			mgr.KnownPeers()[p2.StringID()] = p2
			mgr.KnownPeers()[p3.StringID()] = p3

			n := mgr.CleanKnownPeers()
			Expect(n).To(Equal(1))
		})

		AfterEach(func() {
			lp.Stop()
		})
	})

	Describe(".SavePeers", func() {

		It("should not store peers less than 20 minutes old", func() {
			addr, _ := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/9000/ipfs/12D3KooWM4yJB31d4hF2F9Vdwuj9WFo1qonoySyw4bVAQ9a9d21o")
			rp := node.NewRemoteNode(addr, lp)
			rp.Timestamp = time.Now()

			mgr = lp.PM()
			mgr.KnownPeers()[rp.StringID()] = rp

			err := mgr.SavePeers()
			Expect(err).To(BeNil())

			result := lp.DB().GetByPrefix([]byte("address"))
			Expect(result).To(BeEmpty())
		})

		It("should successfully store 2 peer addresses", func() {

			mgr.SetNumActiveConnections(3)

			addr, _ := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/9000/ipfs/12D3KooWM4yJB31d4hF2F9Vdwuj9WFo1qonoySyw4bVAQ9a9d21o")
			addr2, _ := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/9000/ipfs/12D3KooWM4yJB31d4hF2F9Vdwuj9WFo1qonoySyw4bVAQ9a9d21d")

			p2 := node.NewRemoteNode(addr, lp)
			p2.Timestamp = time.Now().Add(21 * time.Minute)

			p3 := node.NewRemoteNode(addr2, lp)
			p3.Timestamp = time.Now().Add(21 * time.Minute)

			mgr.KnownPeers()[p2.StringID()] = p2
			mgr.KnownPeers()[p3.StringID()] = p3

			err := mgr.SavePeers()
			Expect(err).To(BeNil())

			result := lp.DB().GetByPrefix([]byte("address"))
			Expect(result).To(HaveLen(2))
		})

	})

	Describe(".loadPeers", func() {
		BeforeEach(func() {
			addr, _ := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/9000/ipfs/12D3KooWM4yJB31d4hF2F9Vdwuj9WFo1qonoySyw4bVAQ9a9d21o")
			addr2, _ := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/9000/ipfs/12D3KooWM4yJB31d4hF2F9Vdwuj9WFo1qonoySyw4bVAQ9a9d21d")
			p2 := node.NewRemoteNode(addr, lp)
			p2.Timestamp = time.Now().Add(21 * time.Minute)
			p3 := node.NewRemoteNode(addr2, lp)
			p3.Timestamp = time.Now().Add(21 * time.Minute)
			mgr.KnownPeers()[p2.StringID()] = p2
			mgr.KnownPeers()[p3.StringID()] = p3
			err := mgr.SavePeers()
			Expect(err).To(BeNil())
			Expect(mgr.KnownPeers()).To(HaveLen(2))
		})

		It("should fetch", func() {
			mgr.SetKnownPeers(map[string]types.Engine{})
			err := mgr.LoadPeers()
			Expect(err).To(BeNil())
			Expect(mgr.KnownPeers()).To(HaveLen(2))
		})
	})

	Describe(".GetKnownPeer", func() {

		It("should return nil when peer is not in known peer list", func() {
			p2, err := node.NewNode(cfg, "127.0.0.1:40002", crypto.NewKeyFromIntSeed(0), log)
			Expect(err).To(BeNil())
			Expect(mgr.GetKnownPeer(p2.StringID())).To(BeNil())
			p2.GetHost().Close()
		})

		It("should return peer when peer is in known peer list", func() {
			p2, err := node.NewNode(cfg, "127.0.0.1:40003", crypto.NewKeyFromIntSeed(3), log)
			mgr.AddOrUpdatePeer(p2)
			Expect(err).To(BeNil())
			actual := mgr.GetKnownPeer(p2.StringID())
			Expect(actual).NotTo(BeNil())
			Expect(actual).To(Equal(p2))
			p2.GetHost().Close()
		})
	})

	Describe(".PeerExist", func() {

		It("peer does not exist, must return false", func() {
			p2, err := node.NewNode(cfg, "127.0.0.1:40002", crypto.NewKeyFromIntSeed(0), log)
			defer closeNode(p2)
			Expect(err).To(BeNil())
			Expect(mgr.PeerExist(lp.StringID())).To(BeFalse())
			p2.GetHost().Close()
		})

		It("peer exists, must return true", func() {
			p2, err := node.NewNode(cfg, "127.0.0.1:40002", crypto.NewKeyFromIntSeed(0), log)
			defer closeNode(p2)
			mgr.AddOrUpdatePeer(p2)
			Expect(err).To(BeNil())
			Expect(mgr.PeerExist(p2.StringID())).To(BeTrue())
			p2.GetHost().Close()
		})
	})

	Describe(".onPeerDisconnet", func() {

		It("should set the disconnected peer's timestamp to at least 1 hour ago", func() {
			p2, err := node.NewNode(cfg, "127.0.0.1:40002", crypto.NewKeyFromIntSeed(0), log)
			Expect(err).To(BeNil())

			mgr.AddOrUpdatePeer(p2)
			currentTimestamp := p2.Timestamp.Unix()

			addr, _ := ma.NewMultiaddr(p2.GetMultiAddr())
			mgr.OnPeerDisconnect(addr)
			newTimestamp := p2.Timestamp.Unix()
			Expect(newTimestamp).ToNot(Equal(currentTimestamp))
			Expect(newTimestamp >= currentTimestamp).ToNot(BeTrue())
			Expect(currentTimestamp - newTimestamp).To(Equal(int64(3600)))
		})
	})

	Describe(".GetKnownPeers", func() {
		It("should return peer1 as the only known peer", func() {
			p2, err := node.NewNode(cfg, "127.0.0.1:40002", crypto.NewKeyFromIntSeed(0), log)
			Expect(err).To(BeNil())
			mgr.AddOrUpdatePeer(p2)
			actual := mgr.GetKnownPeers()
			Expect(actual).To(HaveLen(1))
			Expect(actual).To(ContainElement(p2))
		})

	})

	Describe(".CreatePeerFromAddress", func() {
		address := "/ip4/127.0.0.1/tcp/40004/ipfs/12D3KooWHHzSeKaY8xuZVzkLbKFfvNgPPeKhFBGrMbNzbm5akpqd"

		It("peer return err='not a valid multiaddr' when address is /ip4/127.0.0.1/tcp/4000", func() {
			err := mgr.CreatePeerFromAddress("/ip4/127.0.0.1/tcp/4000")
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(Equal("not a valid multiaddr"))
		})

		It("peer return err='not a valid multiaddr' when address is /ip4/127.0.0.1/tcp/4000/ipfs", func() {
			err := mgr.CreatePeerFromAddress("/ip4/127.0.0.1/tcp/4000/ipfs")
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(Equal("not a valid multiaddr"))
		})

		It("peer with address '/ip4/127.0.0.1/tcp/40004/ipfs/12D3KooWHHzSeKaY8xuZVzkLbKFfvNgPPeKhFBGrMbNzbm5akpqd' must be added", func() {
			err := mgr.CreatePeerFromAddress(address)
			Expect(err).To(BeNil())
			p := mgr.KnownPeers()["12D3KooWHHzSeKaY8xuZVzkLbKFfvNgPPeKhFBGrMbNzbm5akpqd"]
			Expect(p).NotTo(BeNil())
		})

		It("duplicate peer should be not be recreated", func() {
			Expect(len(mgr.KnownPeers())).To(Equal(0))
			mgr.CreatePeerFromAddress(address)
			Expect(len(mgr.KnownPeers())).To(Equal(1))
			mgr.CreatePeerFromAddress(address)
			Expect(len(mgr.KnownPeers())).To(Equal(1))
		})
	})

	Describe(".IsActive", func() {

		It("should return false when Timestamp is zero", func() {
			peer := emptyNode(nil)
			Expect(mgr.IsActive(peer)).To(BeFalse())
		})

		It("should return false when Timestamp is 3 hours 10 seconds ago", func() {
			peer := emptyNode(&node.Node{
				Timestamp: time.Now().Add((-3 * (60 * 60) * time.Second) + 10),
			})
			Expect(mgr.IsActive(peer)).To(BeFalse())
		})

		It("should return false when Timestamp is 3 hours ago", func() {
			peer := emptyNode(&node.Node{
				Timestamp: time.Now().Add(-3 * (60 * 60) * time.Second),
			})
			Expect(mgr.IsActive(peer)).To(BeFalse())
		})

		It("should return true when Timestamp is 2 hours, 59 minute ago", func() {
			peer := emptyNode(&node.Node{
				Timestamp: time.Now().Add((-2 * (60 * 60) * time.Second) + 59*time.Minute),
			})
			Expect(mgr.IsActive(peer)).To(BeTrue())
		})

	})

	Describe(".GetActivePeers", func() {

		var peer1, peer2, peer3 *node.Node
		var mgr *node.Manager

		BeforeEach(func() {
			mgr = lp.PM()
			mgr.SetLocalNode(lp)
			peer1 = emptyNode(&node.Node{Timestamp: time.Now().Add(-1 * (60 * 60) * time.Second)})
			peer2 = emptyNode(&node.Node{Timestamp: time.Now().Add(-2 * (60 * 60) * time.Second)})
			peer3 = emptyNode(&node.Node{Timestamp: time.Now().Add(-3 * (60 * 60) * time.Second)})
			mgr.SetKnownPeers(map[string]types.Engine{
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
			peer1 := emptyNode(&node.Node{Timestamp: time.Now().Add(-1 * (60 * 60) * time.Second)})
			mgr.SetKnownPeers(map[string]types.Engine{
				"peer1": peer1,
			})
		})

		It("should return a different slice from the original knownPeer slice", func() {
			actual := mgr.CopyActivePeers(1)
			Expect(actual).To(HaveLen(1))
			Expect(actual).NotTo(Equal(mgr.KnownPeers))
		})
	})

	Describe(".GetRandomActivePeers", func() {

		It("should shuffle the slice of peers if the number of known/active peers is equal to the limit requested", func() {
			peer1 := emptyNode(&node.Node{Timestamp: time.Now().Add(-1 * (60 * 60) * time.Second)})
			peer2 := emptyNode(&node.Node{Timestamp: time.Now().Add(-2 * (60 * 60) * time.Second)})
			peer3 := emptyNode(&node.Node{Timestamp: time.Now().Add(-2 * (60 * 60) * time.Second)})
			mgr.SetKnownPeers(map[string]types.Engine{
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
			peer1 := emptyNode(&node.Node{Timestamp: time.Now().Add(-1 * (60 * 60) * time.Second)})
			peer2 := emptyNode(&node.Node{Timestamp: time.Now().Add(-2 * (60 * 60) * time.Second)})
			peer3 := emptyNode(&node.Node{Timestamp: time.Now().Add(-2 * (60 * 60) * time.Second)})
			mgr.SetKnownPeers(map[string]types.Engine{
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

	Describe(".NeedMorePeers", func() {

		It("should return true when peer manager does not have upto 1000 peers and has not reached max connection", func() {
			cfg.Node.MaxConnections = 120
			peer1 := emptyNode(&node.Node{Timestamp: time.Now().Add(-1 * (60 * 60) * time.Second)})
			mgr.SetKnownPeers(map[string]types.Engine{
				"peer1": peer1,
			})
			Expect(mgr.NeedMorePeers()).To(BeTrue())
		})

		It("should return false when peer manager does not have upto 1000 peers and but has reached max connection", func() {
			cfg.Node.MaxConnections = 10
			mgr.SetNumActiveConnections(10)
			peer1 := emptyNode(&node.Node{Timestamp: time.Now().Add(-1 * (60 * 60) * time.Second)})
			mgr.SetKnownPeers(map[string]types.Engine{
				"peer1": peer1,
			})
			Expect(mgr.NeedMorePeers()).To(BeFalse())
		})
	})

	Describe(".TimestampPunishment", func() {

		It("return err.Error('nil passed') when nil is passed as peer", func() {
			err := mgr.OnFailedConnection(nil)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("nil passed"))
		})

		It("reduce timestamp by an 3600 seconds", func() {
			peer1 := emptyNode(&node.Node{Timestamp: time.Now()})
			mgr.SetKnownPeers(map[string]types.Engine{
				"peer1": peer1,
			})
			currentTime := peer1.Timestamp.Unix()
			err := mgr.OnFailedConnection(peer1)
			Expect(err).To(BeNil())
			actual := peer1.Timestamp.Unix()
			Expect(currentTime > actual).To(BeTrue())
			Expect(currentTime - actual).To(Equal(int64(3600)))
		})

	})

	Describe(".IsLocalPeer", func() {

		It("should return false if nil is passed", func() {
			Expect(mgr.IsLocalNode(nil)).To(BeFalse())
		})

		It("should return false if local peer is nil", func() {
			mgr.SetLocalNode(nil)
			peer1 := emptyNode(&node.Node{Timestamp: time.Now()})
			Expect(mgr.IsLocalNode(peer1)).To(BeFalse())
		})

		It("should return false if not local peer", func() {
			peer1, err := node.NewNode(cfg, "127.0.0.1:40002", crypto.NewKeyFromIntSeed(0), log)
			Expect(err).To(BeNil())
			Expect(mgr.IsLocalNode(peer1)).To(BeFalse())
			peer1.GetHost().Close()
		})

		It("should return true if peer is the local peer", func() {
			peer1, err := node.NewNode(cfg, "127.0.0.1:40002", crypto.NewKeyFromIntSeed(0), log)
			Expect(err).To(BeNil())
			defer peer1.GetHost().Close()
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

	Describe(".GetUnconnectedPeers", func() {

		var host, host2 host.Host
		var p, p2 *node.Node
		var err error

		BeforeEach(func() {
			p, err = node.NewNode(cfg, "127.0.0.1:40106", crypto.NewKeyFromIntSeed(6), log)
			Expect(err).To(BeNil())
			p2, err = node.NewNode(cfg, "127.0.0.1:40107", crypto.NewKeyFromIntSeed(7), log)
			Expect(err).To(BeNil())
			p2.SetLocalNode(p)
			host = p.Host()
			Expect(err).To(BeNil())
			host2 = p2.Host()
			Expect(err).To(BeNil())

			host.SetStreamHandler("/protocol/0.0.1", testutil.NoOpStreamHandler)
			host.Peerstore().AddAddr(host2.ID(), host2.Addrs()[0], pstore.PermanentAddrTTL)
			host.Connect(context.Background(), host.Peerstore().PeerInfo(host2.ID()))
		})

		It("should return empty slice when all peers are connected", func() {
			peers := p.PM().GetUnconnectedPeers()
			Expect(peers).To(HaveLen(0))
		})

		It("should return p3 in slice when only 1 peer is not connected", func() {
			p3, err := node.NewNode(cfg, "127.0.0.1:40108", crypto.NewKeyFromIntSeed(8), log)
			Expect(err).To(BeNil())
			defer p3.Host().Close()
			p.PM().AddOrUpdatePeer(p3)
			peers := p.PM().GetUnconnectedPeers()
			Expect(peers).To(HaveLen(1))
			Expect(peers[0]).To(Equal(p3))
		})

		AfterEach(func() {
			p.Stop()
			closeNode(p2)
		})
	})
})
