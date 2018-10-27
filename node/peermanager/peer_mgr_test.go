package peermanager_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/imdario/mergo"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	"github.com/shopspring/decimal"

	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/node"
	"github.com/ellcrys/elld/node/peermanager"
	"github.com/ellcrys/elld/params"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/util"
	ma "github.com/multiformats/go-multiaddr"

	. "github.com/ncodes/goblin"
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

func NewMgr(cfg *config.EngineConfig, localNode *node.Node) *peermanager.Manager {
	mgr := peermanager.NewManager(cfg, localNode, log)
	return mgr
}

func TestPeerManager(t *testing.T) {
	params.FeePerByte = decimal.NewFromFloat(0.01)
	g := Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("Peer Manager", func() {

		var lp, rp *node.Node
		var mgr *peermanager.Manager
		var cfg *config.EngineConfig

		g.BeforeEach(func() {
			lp = makeTestNodeWith(getPort(), 1)
			Expect(lp.GetBlockchain().Up()).To(BeNil())
			mgr = lp.PM()
			mgr.SetLocalNode(lp)
			cfg = lp.GetCfg()

			rp = makeTestNodeWith(getPort(), 2)
			Expect(rp.GetBlockchain().Up()).To(BeNil())
			mgr = rp.PM()
			mgr.SetLocalNode(rp)
		})

		g.AfterEach(func() {
			closeNode(lp)
			closeNode(rp)
		})

		g.Describe(".AddOrUpdatePeer", func() {
			var p2 *node.Node

			g.When("peer does not previously exist", func() {

				g.BeforeEach(func() {
					p2, _ = node.NewNode(cfg, "127.0.0.1:40002", crypto.NewKeyFromIntSeed(0), log)
					defer closeNode(p2)
					mgr.AddOrUpdateNode(p2)
				})

				g.It("should add peer with time set to an hour ago", func() {
					Expect(mgr.Peers()).To(HaveLen(1))
					Expect(p2.GetLastSeen().IsZero()).To(BeFalse())
					Expect(p2.GetLastSeen().Unix()).To(Equal(time.Now().Add(-1 * time.Hour).Unix()))
				})
			})

			g.When("peer previously existed and not currently connected to the local node", func() {
				var existingPeer *node.Node
				var err error

				g.BeforeEach(func() {
					existingPeer, err = node.NewNode(cfg, "127.0.0.1:40002", crypto.NewKeyFromIntSeed(0), log)
					Expect(err).To(BeNil())
					existingPeer.SetLastSeen(time.Now())
					defer closeNode(existingPeer)
					mgr.AddPeer(existingPeer)
					mgr.AddOrUpdateNode(existingPeer)
				})

				g.It("should deduct one hour ago from its current time", func() {
					Expect(existingPeer.GetLastSeen().Unix()).To(Equal(time.Now().Add(-1 * time.Hour).Unix()))
				})
			})

			g.When("peer previously existed and currently connected to the local node", func() {
				var existingPeer *node.Node
				var err error

				g.BeforeEach(func() {
					existingPeer, err = node.NewNode(cfg, "127.0.0.1:40002", crypto.NewKeyFromIntSeed(0), log)
					Expect(err).To(BeNil())
					existingPeer.SetLocalNode(lp)

					lp.GetHost().Peerstore().AddAddr(existingPeer.ID(), existingPeer.GetAddress().DecapIPFS(), pstore.PermanentAddrTTL)
					err := lp.GetHost().Connect(context.TODO(), lp.GetHost().Peerstore().PeerInfo(existingPeer.ID()))
					Expect(err).To(BeNil())

					mgr.AddPeer(existingPeer)
				})

				g.AfterEach(func() {
					closeNode(existingPeer)
				})

				g.It("should update the last seen to current time", func() {
					now := time.Now().Unix()
					mgr.AddOrUpdateNode(existingPeer)
					Expect(existingPeer.GetLastSeen().Unix()).To(Equal(now))
				})
			})
		})

		g.Describe(".AddTimeBan", func() {
			g.When("peer has not been time banned before", func() {
				g.It("should add peer with non-zero time", func() {
					mgr.AddTimeBan(lp, 20*time.Second)
					entry := mgr.TimeBanIndex()[lp.GetAddress().IP().String()]
					Expect(entry.IsZero()).To(BeFalse())
				})
			})

			g.When("peer has been added with a later ban time", func() {
				g.BeforeEach(func() {
					mgr.AddTimeBan(lp, -20*time.Second)
					entry := mgr.TimeBanIndex()[lp.GetAddress().IP().String()]
					Expect(entry.Before(time.Now())).To(BeTrue())
				})

				g.It("should update the time to the current time before adding ban time", func() {
					mgr.AddTimeBan(lp, 20*time.Minute)
					entry := mgr.TimeBanIndex()[lp.GetAddress().IP().String()]
					Expect(entry.After(time.Now())).To(BeTrue())
				})
			})
		})

		g.Describe(".IsBanned", func() {
			g.When("peer has not been added to the time ban index", func() {
				g.It("should return false", func() {
					Expect(mgr.IsBanned(lp)).To(BeFalse())
				})
			})

			g.When("peer has been added to the time ban index", func() {
				g.When("peer ban end time is ahead of current time", func() {
					g.It("should return true", func() {
						mgr.AddTimeBan(lp, 20*time.Minute)
						Expect(mgr.IsBanned(lp)).To(BeTrue())
					})
				})

				g.When("peer ban end time is before the current time", func() {
					g.It("should return false", func() {
						mgr.AddTimeBan(lp, -20*time.Minute)
						Expect(mgr.IsBanned(lp)).To(BeFalse())
					})
				})
			})
		})

		g.Describe(".GetUnconnectedPeers", func() {

			var n *node.Node

			g.BeforeEach(func() {
				n = makeTestNode(getPort())
				lp.PM().AddPeer(n)
			})

			g.AfterEach(func() {
				closeNode(n)
			})

			g.It("should get correct unconnected peer", func() {
				ucp := lp.PM().GetUnconnectedPeers()
				Expect(ucp).To(HaveLen(1))
				Expect(ucp[0]).To(Equal(n))
			})
		})

		g.Describe(".canAcceptNode", func() {

			var n *node.Node

			g.BeforeEach(func() {
				n = makeTestNode(getPort())
				lp.PM().AddPeer(n)
			})

			g.AfterEach(func() {
				closeNode(n)
			})

			g.When("in test mode", func() {
				g.It("should return true", func() {
					cfg.Node.Mode = config.ModeTest
					accept, err := lp.PM().CanAcceptNode(n)
					Expect(err).To(BeNil())
					Expect(accept).To(BeTrue())
				})
			})

			g.When("node is not acquainted", func() {
				g.It("should return false and err", func() {
					cfg.Node.Mode = config.ModeProd
					accept, err := lp.PM().CanAcceptNode(n)
					Expect(err).ToNot(BeNil())
					Expect(err.Error()).To(Equal("unacquainted node"))
					Expect(accept).To(BeFalse())
				})
			})

			g.When("node is serving ban time", func() {
				g.When("ban time is over 3 hours ago", func() {
					g.It("should return false and err", func() {
						cfg.Node.Mode = config.ModeProd
						lp.PM().AddAcquainted(n)
						lp.PM().AddTimeBan(n, 4*time.Hour)

						accept, err := lp.PM().CanAcceptNode(n)
						Expect(err).ToNot(BeNil())
						Expect(err.Error()).To(Equal("currently serving ban time"))
						Expect(accept).To(BeFalse())
					})
				})

				g.When("ban time is less than 3 hours ago", func() {
					g.It("should return false and err", func() {
						cfg.Node.Mode = config.ModeProd
						lp.PM().AddAcquainted(n)
						lp.PM().AddTimeBan(n, 2*time.Hour)

						accept, err := lp.PM().CanAcceptNode(n)
						Expect(err).To(BeNil())
						Expect(accept).To(BeTrue())
					})
				})
			})

			g.When("node is acquainted and not serving ban time", func() {
				g.It("should return true", func() {
					cfg.Node.Mode = config.ModeProd
					lp.PM().AddAcquainted(n)

					accept, err := lp.PM().CanAcceptNode(n)
					Expect(err).To(BeNil())
					Expect(accept).To(BeTrue())
				})
			})
		})

		g.Describe("connection failure count", func() {

			var n *node.Node

			g.BeforeEach(func() {
				n = makeTestNode(getPort())
				lp.PM().AddPeer(n)
			})

			g.AfterEach(func() {
				closeNode(n)
			})

			g.Describe(".GetConnFailCount", func() {
				g.When("no failure record", func() {
					g.It("should return zero", func() {
						Expect(lp.PM().GetConnFailCount(n.GetAddress())).To(Equal(0))
					})
				})

				g.When("1 failure had been recorded", func() {
					g.It("should return 1", func() {
						lp.PM().IncrConnFailCount(n.GetAddress())
						Expect(lp.PM().GetConnFailCount(n.GetAddress())).To(Equal(1))
					})
				})

				g.When("1 failure had been recorded", func() {

					g.BeforeEach(func() {
						lp.PM().IncrConnFailCount(n.GetAddress())
						Expect(lp.PM().GetConnFailCount(n.GetAddress())).To(Equal(1))
					})

					g.When(".ClearConnFailCount is called", func() {
						g.It("should return 0", func() {
							lp.PM().ClearConnFailCount(n.GetAddress())
							Expect(lp.PM().GetConnFailCount(n.GetAddress())).To(Equal(0))
						})
					})
				})
			})
		})

		g.Describe(".CleanPeers", func() {

			g.It("should return 0 when number of connected peers is less than 3", func() {
				n := mgr.CleanPeers()
				Expect(n).To(BeZero())
			})

			g.When("all peers are active", func() {
				g.It("should remove 0 peers", func() {
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

			g.When("peer is not banned but has not been see in the last 3 hours", func() {

				var p3 *node.Node

				g.BeforeEach(func() {
					addr2, _ := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/9000/ipfs/12D3KooWM4yJB31d4hF2F9Vdwuj9WFo1qonoySyw4bVAQ9a9d21d")
					p3 = node.NewRemoteNodeFromMultiAddr(addr2, lp)
					p3.SetLastSeen(time.Now().Add(-4 * time.Hour))
					mgr.AddPeer(p3)
				})

				g.It("should remove 1 peer", func() {
					n := mgr.CleanPeers()
					Expect(n).To(Equal(1))
				})

				g.It("should remove the peer from acquainted cache", func() {
					n := mgr.CleanPeers()
					Expect(n).To(Equal(1))
					Expect(mgr.IsAcquainted(p3)).To(BeFalse())
				})
			})

			g.When("one peer is banned for over 3 hours", func() {

				var p3 *node.Node

				g.BeforeEach(func() {
					addr, _ := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/9000/ipfs/12D3KooWM4yJB31d4hF2F9Vdwuj9WFo1qonoySyw4bVAQ9a9d21o")
					p2 := node.NewRemoteNodeFromMultiAddr(addr, lp)
					p2.SetLastSeen(time.Now())
					mgr.Peers()[p2.StringID()] = p2
					mgr.AddTimeBan(p2, 4*time.Hour)

					addr2, _ := ma.NewMultiaddr("/ip4/127.0.0.2/tcp/9000/ipfs/12D3KooWM4yJB31d4hF2F9Vdwuj9WFo1qonoySyw4bVAQ9a9d21d")
					p3 = node.NewRemoteNodeFromMultiAddr(addr2, lp)
					p3.SetLastSeen(time.Now())
					mgr.Peers()[p3.StringID()] = p3
				})

				g.It("should remove 1 peer", func() {
					n := mgr.CleanPeers()
					Expect(n).To(Equal(1))
					Expect(mgr.PeerExist(p3.StringID())).To(BeTrue())
				})
			})

			g.When("one peer is banned for less than 3 hours", func() {

				var p *node.Node

				g.BeforeEach(func() {
					addr, _ := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/9000/ipfs/12D3KooWM4yJB31d4hF2F9Vdwuj9WFo1qonoySyw4bVAQ9a9d21o")
					p = node.NewRemoteNodeFromMultiAddr(addr, lp)
					p.SetLastSeen(time.Now())
					mgr.Peers()[p.StringID()] = p
					mgr.AddTimeBan(p, 2*time.Hour)
				})

				g.It("should not remove peer", func() {
					n := mgr.CleanPeers()
					Expect(n).To(Equal(0))
					Expect(mgr.PeerExist(p.StringID())).To(BeTrue())
				})
			})
		})

		g.Describe(".SavePeers", func() {

			g.When("peer was created less than 20 minutes ago", func() {
				g.It("should not store peer", func() {
					addr, _ := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/9000/ipfs/12D3KooWM4yJB31d4hF2F9Vdwuj9WFo1qonoySyw4bVAQ9a9d21o")
					rp := node.NewRemoteNodeFromMultiAddr(addr, lp)
					rp.SetCreatedAt(time.Now())

					mgr = lp.PM()
					mgr.Peers()[rp.StringID()] = rp

					err := mgr.SavePeers()
					Expect(err).To(BeNil())

					result := mgr.LocalPeer().DB().GetByPrefix([]byte("address"))
					Expect(result).To(BeEmpty())
				})
			})

			g.When("peers were created more than 20 minutes ago", func() {
				g.It("should successfully store peers", func() {

					addr, _ := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/9000/ipfs/12D3KooWM4yJB31d4hF2F9Vdwuj9WFo1qonoySyw4bVAQ9a9d21o")
					addr2, _ := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/9000/ipfs/12D3KooWM4yJB31d4hF2F9Vdwuj9WFo1qonoySyw4bVAQ9a9d21d")

					p2 := node.NewRemoteNodeFromMultiAddr(addr, lp)
					p2.SetLastSeen(time.Now())
					p2.SetCreatedAt(time.Now().Add(-21 * time.Minute))
					mgr.AddPeer(p2)

					p3 := node.NewRemoteNodeFromMultiAddr(addr2, lp)
					p3.SetLastSeen(time.Now())
					p3.SetCreatedAt(time.Now().Add(-21 * time.Minute))
					mgr.AddPeer(p3)

					err := mgr.SavePeers()
					Expect(err).To(BeNil())

					result := mgr.LocalPeer().DB().GetByPrefix([]byte("address"))
					Expect(result).To(HaveLen(2))
				})
			})

			g.When("peer has a ban time", func() {

				var p2 *node.Node

				g.BeforeEach(func() {
					addr, _ := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/9000/ipfs/12D3KooWM4yJB31d4hF2F9Vdwuj9WFo1qonoySyw4bVAQ9a9d21o")
					p2 = node.NewRemoteNodeFromMultiAddr(addr, lp)
					p2.SetLastSeen(time.Now())
					p2.SetCreatedAt(time.Now().Add(-21 * time.Minute))
					mgr.AddTimeBan(p2, 20*time.Minute)
					mgr.Peers()[p2.StringID()] = p2
				})

				g.It("should be saved along with other data", func() {

					err := mgr.SavePeers()
					Expect(err).To(BeNil())

					result := mgr.LocalPeer().DB().GetByPrefix([]byte("address"))
					Expect(result).To(HaveLen(1))

					var m map[string]interface{}
					result[0].Scan(&m)
					Expect(int64(m["banTime"].(uint32))).To(Equal(mgr.GetBanTime(p2).Unix()))
				})
			})

		})

		g.Describe(".loadPeers", func() {

			var lastSeen time.Time
			var createdAt time.Time
			var peer *node.Node

			g.Context("load peer with no ban time", func() {
				g.BeforeEach(func() {
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

				g.It("should fetch 1 address", func() {
					mgr.SetPeers(map[string]core.Engine{})
					err := mgr.LoadPeers()
					Expect(err).To(BeNil())
					Expect(mgr.Peers()).To(HaveKey(peer.StringID()))
					peer := mgr.Peers()[peer.StringID()]
					Expect(peer.GetLastSeen().Unix()).To(Equal(lastSeen.Unix()))
					Expect(peer.CreatedAt().Unix()).To(Equal(createdAt.Unix()))
				})
			})

			g.Context("load peer with ban time", func() {

				var banTime int64

				g.BeforeEach(func() {
					createdAt = time.Now().Add(-21 * time.Minute)
					lastSeen = time.Now()
					addr, _ := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/9000/ipfs/12D3KooWM4yJB31d4hF2F9Vdwuj9WFo1qonoySyw4bVAQ9a9d21o")

					peer = node.NewRemoteNodeFromMultiAddr(addr, lp)
					peer.SetCreatedAt(createdAt)
					peer.SetLastSeen(lastSeen)
					mgr.AddPeer(peer)
					Expect(mgr.Peers()).To(HaveLen(1))

					mgr.AddTimeBan(peer, 20*time.Minute)
					banTime = mgr.GetBanTime(peer).Unix()

					err := mgr.SavePeers()
					Expect(err).To(BeNil())
				})

				g.It("should fetch 1 address including its ban time", func() {
					mgr.SetPeers(map[string]core.Engine{})
					err := mgr.LoadPeers()
					Expect(err).To(BeNil())
					Expect(mgr.Peers()).To(HaveKey(peer.StringID()))
					peer := mgr.Peers()[peer.StringID()]
					Expect(peer.GetLastSeen().Unix()).To(Equal(lastSeen.Unix()))
					Expect(peer.CreatedAt().Unix()).To(Equal(createdAt.Unix()))
					Expect(mgr.GetBanTime(peer).Unix()).To(Equal(banTime))
				})
			})
		})

		g.Describe(".GetKnownPeer", func() {

			g.It("should return nil when peer is not in known peer list", func() {
				p2, err := node.NewNode(cfg, "127.0.0.1:40002", crypto.NewKeyFromIntSeed(0), log)
				Expect(err).To(BeNil())
				defer closeNode(p2)
				Expect(mgr.GetPeer(p2.StringID())).To(BeNil())
			})

			g.It("should return peer when peer is in known peer list", func() {
				p2, err := node.NewNode(cfg, "127.0.0.1:40003", crypto.NewKeyFromIntSeed(3), log)
				defer closeNode(p2)
				mgr.AddOrUpdateNode(p2)
				Expect(err).To(BeNil())
				actual := mgr.GetPeer(p2.StringID())
				Expect(actual).NotTo(BeNil())
				Expect(actual).To(Equal(p2))
			})
		})

		g.Describe(".PeerExist", func() {

			g.It("peer does not exist, must return false", func() {
				p2, err := node.NewNode(cfg, "127.0.0.1:40002", crypto.NewKeyFromIntSeed(0), log)
				defer closeNode(p2)
				Expect(err).To(BeNil())
				Expect(mgr.PeerExist(lp.StringID())).To(BeFalse())
				closeNode(p2)
			})

			g.It("peer exists, must return true", func() {
				p2, err := node.NewNode(cfg, "127.0.0.1:40002", crypto.NewKeyFromIntSeed(0), log)
				defer closeNode(p2)
				mgr.AddOrUpdateNode(p2)
				Expect(err).To(BeNil())
				Expect(mgr.PeerExist(p2.StringID())).To(BeTrue())
				closeNode(p2)
			})
		})

		g.Describe(".GetPeers", func() {
			g.It("should return peer1 as the only known peer", func() {
				p2, err := node.NewNode(cfg, "127.0.0.1:40002", crypto.NewKeyFromIntSeed(0), log)
				Expect(err).To(BeNil())
				defer closeNode(p2)
				mgr.AddOrUpdateNode(p2)
				actual := mgr.GetPeers()
				Expect(actual).To(HaveLen(1))
				Expect(actual).To(ContainElement(p2))
			})

		})

		g.Describe(".IsActive", func() {

			g.It("should return false when Timestamp is zero", func() {
				peer := emptyNode(nil)
				Expect(mgr.IsActive(peer)).To(BeFalse())
			})

			g.It("should return false when Timestamp is 3 hours 10 seconds ago", func() {
				peer := emptyNode(nil)
				peer.SetLastSeen(time.Now().Add((-3 * (60 * 60) * time.Second) + 10))
				Expect(mgr.IsActive(peer)).To(BeFalse())
			})

			g.It("should return false when Timestamp is 3 hours ago", func() {
				peer := emptyNode(nil)
				peer.SetLastSeen(time.Now().Add(-3 * (60 * 60) * time.Second))
				Expect(mgr.IsActive(peer)).To(BeFalse())
			})

			g.It("should return true when Timestamp is 2 hours, 59 minute ago", func() {
				peer := emptyNode(nil)
				peer.SetLastSeen(time.Now().Add((-2 * (60 * 60) * time.Second) + 59*time.Minute))
				Expect(mgr.IsActive(peer)).To(BeTrue())
			})

		})

		g.Describe(".RandomPeers", func() {

			var peer1, peer2, peer3 *node.Node
			var mgr *peermanager.Manager

			g.BeforeEach(func() {
				mgr = lp.PM()
				mgr.SetLocalNode(lp)
				peer1 = emptyNodeWithLastSeenTime(time.Now())
				peer2 = emptyNodeWithLastSeenTime(time.Now())
				peer3 = emptyNodeWithLastSeenTime(time.Now())
				mgr.SetPeers(map[string]core.Engine{
					"peer1": peer1,
					"peer2": peer2,
					"peer3": peer3,
				})
			})

			g.It("", func() {

			})
		})

		g.Describe(".GetActivePeers", func() {

			var peer1, peer2, peer3 *node.Node
			var mgr *peermanager.Manager

			g.BeforeEach(func() {
				mgr = lp.PM()
				mgr.SetLocalNode(lp)
				peer1 = emptyNodeWithLastSeenTime(time.Now().Add(-1 * (60 * 60) * time.Second))
				peer2 = emptyNodeWithLastSeenTime(time.Now().Add(-2 * (60 * 60) * time.Second))
				peer3 = emptyNodeWithLastSeenTime(time.Now().Add(-3 * (60 * 60) * time.Second))
				mgr.SetPeers(map[string]core.Engine{
					"peer1": peer1,
					"peer2": peer2,
					"peer3": peer3,
				})
			})

			g.It("should return a map with 2 elements with id peer1 and peer2 when limit is set to 0", func() {
				actual := mgr.GetActivePeers(0)
				Expect(actual).To(HaveLen(2))
				Expect(actual).To(ContainElement(peer1))
				Expect(actual).To(ContainElement(peer2))
			})

			g.It("should return a map with 2 elements with id peer1 and peer2 when limit is set to a negative number", func() {
				actual := mgr.GetActivePeers(-1)
				Expect(actual).To(HaveLen(2))
				Expect(actual).To(ContainElement(peer1))
				Expect(actual).To(ContainElement(peer2))
			})

			g.It("should return a map with 1 elements with either peer1 or peer2 when limit is set to 1", func() {
				actual := mgr.GetActivePeers(1)
				Expect(actual).To(HaveLen(1))
				Expect([]*node.Node{peer1, peer2}).To(ContainElement(actual[0]))
			})

		})

		g.Describe(".CopyActivePeers", func() {

			var mgr *peermanager.Manager

			g.BeforeEach(func() {
				mgr = NewMgr(cfg, lp)
				peer1 := emptyNodeWithLastSeenTime(time.Now().Add(-1 * (60 * 60) * time.Second))
				mgr.SetPeers(map[string]core.Engine{
					"peer1": peer1,
				})
			})

			g.It("should return a different slice from the original knownPeer slice", func() {
				actual := mgr.CopyActivePeers(1)
				Expect(actual).To(HaveLen(1))
				Expect(actual).NotTo(Equal(mgr.Peers))
			})
		})

		g.Describe(".GetRandomActivePeers", func() {

			g.It("should shuffle the slice of peers if the number of known/active peers is equal to the limit requested", func() {
				peer1 := emptyNodeWithLastSeenTime(time.Now().Add(-1 * (60 * 60) * time.Second))
				peer2 := emptyNodeWithLastSeenTime(time.Now().Add(-2 * (60 * 60) * time.Second))
				peer3 := emptyNodeWithLastSeenTime(time.Now().Add(-2 * (60 * 60) * time.Second))
				mgr.SetPeers(map[string]core.Engine{
					"peer1": peer1,
					"peer2": peer2,
					"peer3": peer3,
				})

				var result []core.Engine
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

			g.It("Should return the limit requested if known active peers are more than limit", func() {
				peer1 := emptyNodeWithLastSeenTime(time.Now().Add(-1 * (60 * 60) * time.Second))
				peer2 := emptyNodeWithLastSeenTime(time.Now().Add(-2 * (60 * 60) * time.Second))
				peer3 := emptyNodeWithLastSeenTime(time.Now().Add(-2 * (60 * 60) * time.Second))
				mgr.SetPeers(map[string]core.Engine{
					"peer1": peer1,
					"peer2": peer2,
					"peer3": peer3,
				})

				var result []core.Engine
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

		g.Describe(".RequirePeers", func() {

			g.Context("when the node does not have up to 1000 active addresses", func() {
				g.Context("when max outbound connection has not been reached", func() {
					g.It("should return true", func() {
						cfg.Node.MaxOutboundConnections = 120
						peer1 := emptyNodeWithLastSeenTime(time.Now().Add(-1 * (60 * 60) * time.Second))
						mgr.SetPeers(map[string]core.Engine{"peer1": peer1})
						Expect(mgr.RequirePeers()).To(BeTrue())
					})
				})

				g.Context("when max outbound connection has been reached", func() {
					g.It("should return false", func() {
						cfg.Node.MaxOutboundConnections = 10
						mgr.ConnMgr().SetConnsInfo(peermanager.NewConnsInfo(0, 10))
						peer1 := emptyNodeWithLastSeenTime(time.Now().Add(-1 * (60 * 60) * time.Second))
						mgr.SetPeers(map[string]core.Engine{"peer1": peer1})
						Expect(mgr.RequirePeers()).To(BeFalse())
					})
				})

			})
		})

		g.Describe(".HasDisconnected", func() {

			g.Context("when address does not contain a known peer id", func() {
				g.It("should return err='unknown peer'", func() {
					addr := "/ip4/127.0.0.1/tcp/9000/ipfs/12D3KooWM4yJB31d4hF2F9Vdwuj9WFo1qonoySyw4bVAQ9a9d212"
					err := mgr.HasDisconnected(util.NodeAddr(addr))
					Expect(err).ToNot(BeNil())
					Expect(err.Error()).To(Equal("unknown peer"))
				})
			})

			g.Context("when address contain a known peer id", func() {
				g.It("should set 'last seen' time to 1 hour ago", func() {
					lp.SetLastSeen(time.Now())
					mgr.SetPeers(map[string]core.Engine{lp.GetAddress().StringID(): lp})
					currentTime := lp.GetLastSeen().Unix()
					err := mgr.HasDisconnected(lp.GetAddress())
					Expect(err).To(BeNil())
					actual := lp.GetLastSeen().Unix()
					Expect(currentTime > actual).To(BeTrue())
					Expect(currentTime - actual).To(Equal(int64(3600)))
				})
			})

		})

		g.Describe(".IsLocalPeer", func() {

			g.It("should return false if nil is passed", func() {
				Expect(mgr.IsLocalNode(nil)).To(BeFalse())
			})

			g.It("should return false if local peer is nil", func() {
				mgr.SetLocalNode(nil)
				peer1 := emptyNodeWithLastSeenTime(time.Now())
				Expect(mgr.IsLocalNode(peer1)).To(BeFalse())
			})

			g.It("should return false if not local peer", func() {
				peer1, err := node.NewNode(cfg, "127.0.0.1:40002", crypto.NewKeyFromIntSeed(0), log)
				Expect(err).To(BeNil())
				defer closeNode(peer1)
				Expect(mgr.IsLocalNode(peer1)).To(BeFalse())
			})

			g.It("should return true if peer is the local peer", func() {
				peer1, err := node.NewNode(cfg, "127.0.0.1:40002", crypto.NewKeyFromIntSeed(0), log)
				Expect(err).To(BeNil())
				defer closeNode(peer1)
				mgr.SetLocalNode(peer1)
				Expect(mgr.IsLocalNode(peer1)).To(BeTrue())
			})
		})

		g.Describe(".ConnectToPeer", func() {
			g.It("should return nil when peer does not exist", func() {
				var mgr = NewMgr(cfg, lp)
				err := mgr.ConnectToPeer("invalid")
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("peer not found"))
			})

			g.When("connection is successful", func() {

				g.BeforeEach(func() {
					lp.PM().AddPeer(rp)
					rp.SetLocalNode(lp)
				})

				g.It("should be connected", func() {
					err := lp.PM().ConnectToPeer(rp.StringID())
					Expect(err).To(BeNil())
					Expect(rp.Connected()).To(BeTrue())
				})
			})
		})

		g.Describe(".ConnectToNode", func() {
			g.When("connection is successful", func() {

				g.BeforeEach(func() {
					lp.PM().AddPeer(rp)
					rp.SetLocalNode(lp)
				})

				g.It("should be connected", func() {
					err := lp.PM().ConnectToNode(rp)
					Expect(err).To(BeNil())
					Expect(rp.Connected()).To(BeTrue())
				})
			})
		})

		g.Describe(".GetConnectedPeers", func() {
			g.When("connection is successful", func() {

				g.BeforeEach(func() {
					lp.PM().AddPeer(rp)
					rp.SetLocalNode(lp)
					err := lp.PM().ConnectToNode(rp)
					Expect(err).To(BeNil())
					Expect(rp.Connected()).To(BeTrue())
				})

				g.It("should return one peer", func() {
					peers := lp.PM().GetConnectedPeers()
					Expect(peers).To(HaveLen(1))
					Expect(peers[0]).To(Equal(rp))
				})
			})
		})
	})
}
