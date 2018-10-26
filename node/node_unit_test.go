package node_test

import (
	"context"
	"testing"

	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/node"
	"github.com/ellcrys/elld/testutil"
	"github.com/ellcrys/elld/util"
	host "github.com/libp2p/go-libp2p-host"
	pstore "github.com/libp2p/go-libp2p-peerstore"

	. "github.com/ncodes/goblin"
	. "github.com/onsi/gomega"
)

func TestNodeUnit(t *testing.T) {
	g := Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("Node Unit Test", func() {

		var n *node.Node
		var cfg *config.EngineConfig
		var lpPort int

		g.BeforeEach(func() {
			lpPort = getPort()
			n = makeTestNodeWith(lpPort, 0)
			cfg = n.GetCfg()
		})

		g.AfterEach(func() {
			closeNode(n)
		})

		g.Describe(".NewNode", func() {
			g.Context("address", func() {
				g.It("return err.Error('failed to parse address. Expects 'ip:port' format') when only port is provided", func() {
					_, err := node.NewNode(cfg, "40100", crypto.NewKeyFromIntSeed(1), log)
					Expect(err).NotTo(BeNil())
					Expect(err.Error()).To(Equal("failed to parse address. Expects 'ip:port' format"))
				})

				g.It("return err.Error('failed to parse address. Expects 'ip:port' format') when only ip is provided", func() {
					_, err := node.NewNode(cfg, "127.0.0.1", crypto.NewKeyFromIntSeed(1), log)
					Expect(err).NotTo(BeNil())
					Expect(err.Error()).To(Equal("failed to parse address. Expects 'ip:port' format"))
				})

				g.It("return err.Error('failed to create host > failed to parse ip4: 127.0.0 failed to parse ip4 addr: 127.0.0') when address is invalid and port is valid", func() {
					_, err := node.NewNode(cfg, "127.0.0:40000", crypto.NewKeyFromIntSeed(1), log)
					Expect(err).NotTo(BeNil())
					Expect(err.Error()).To(Equal("failed to create host > failed to parse ip4: 127.0.0 failed to parse ip4 addr: 127.0.0"))
				})

				g.It("return nil if address is ':40000'", func() {
					n, err := node.NewNode(cfg, ":40000", crypto.NewKeyFromIntSeed(1), log)
					Expect(err).To(BeNil())
					closeNode(n)
				})

				g.It("return nil if address is '127.0.0.1:40000'", func() {
					n, err := node.NewNode(cfg, "127.0.0.1:40000", crypto.NewKeyFromIntSeed(1), log)
					Expect(err).To(BeNil())
					closeNode(n)
				})
			})
		})

		g.Describe(".ID", func() {
			g.It("should return empty string when node has no address", func() {
				p := node.Node{}
				Expect(p.ID().Pretty()).To(Equal(""))
			})

			g.It("should return '12D3KooWL3XJ9EMCyZvmmGXL2LMiVBtrVa2BuESsJiXkSj7333Jw'", func() {
				Expect(n.ID().Pretty()).To(Equal("12D3KooWL3XJ9EMCyZvmmGXL2LMiVBtrVa2BuESsJiXkSj7333Jw"))
				closeNode(n)
			})
		})

		g.Describe(".IDPretty", func() {
			g.It("should return empty string when node has no address", func() {
				p := node.Node{}
				Expect(p.StringID()).To(Equal(""))
			})

			g.It("should return '12D3KooWL3XJ9EMCyZvmmGXL2LMiVBtrVa2BuESsJiXkSj7333Jw'", func() {
				Expect(n.StringID()).To(Equal("12D3KooWL3XJ9EMCyZvmmGXL2LMiVBtrVa2BuESsJiXkSj7333Jw"))
				closeNode(n)
			})
		})

		g.Describe(".IDShort", func() {
			g.It("should return empty string when node has no address", func() {
				p := node.Node{}
				Expect(p.ShortID()).To(Equal(""))
			})

			g.It("should return '12D3KooWL3XJ..JiXkSj7333Jw'", func() {
				Expect(n.ShortID()).To(Equal("12D3KooWL3XJ..JiXkSj7333Jw"))
				closeNode(n)
			})
		})

		g.Describe(".PrivKey", func() {
			g.It("should return private key", func() {
				Expect(n.PrivKey()).NotTo(BeNil())
				closeNode(n)
			})
		})

		g.Describe(".GetMultiAddr", func() {

			g.It("should return empty util.NodeAddr when node has no host", func() {
				n := node.Node{}
				Expect(n.GetAddress()).To(Equal(util.NodeAddr("")))
			})

			g.It("should return address", func() {
				Expect(n.GetAddress()).ToNot(BeEmpty())
			})
		})

		g.Describe(".AddBootstrapNodes", func() {
			g.Context("with empty address", func() {
				g.It("peer manager's bootstrap list should be empty", func() {
					n.AddAddresses(nil, false)
					Expect(n.PM().GetPeers()).To(HaveLen(0))
				})
			})

			g.Context("with invalid address", func() {
				g.It("peer manager's bootstrap list should not contain invalid address", func() {
					n.AddAddresses([]string{"/ip4/127.0.0.1/tcp/40000"}, false)
					Expect(n.PM().GetPeers()).To(HaveLen(0))
				})

				g.It("peer manager's bootstrap list contain only one valid address", func() {
					addresses := []string{
						"/ip4/127.0.0.1/tcp/40000",
						"127.0.0.0:1",
						"ellcrys://12D3KooWL3XJ9EMCyZvmmGXL2LMiVBtrVa2BuESsJiXkSj7333Jw@127.0.0.1:9000",
					}
					n.AddAddresses(addresses, false)
					Expect(n.PM().GetPeers()).To(HaveLen(1))
					Expect(n.PM().GetPeer("12D3KooWL3XJ9EMCyZvmmGXL2LMiVBtrVa2BuESsJiXkSj7333Jw")).ToNot(BeNil())
				})
			})
		})

		g.Describe(".PeersPublicAddrs", func() {

			var host, host2, host3 host.Host
			var n *node.Node
			var err error

			g.BeforeEach(func() {
				n, err = node.NewNode(cfg, "127.0.0.1:40105", crypto.NewKeyFromIntSeed(5), log)
				Expect(err).To(BeNil())
				host = n.GetHost()
				Expect(err).To(BeNil())
				host2, err = testutil.RandomHost(6, 40106)
				Expect(err).To(BeNil())
				host3, err = testutil.RandomHost(7, 40107)
				Expect(err).To(BeNil())

				host.SetStreamHandler("/protocol/0.0.1", testutil.NoOpStreamHandler)
				host.Peerstore().AddAddr(host2.ID(), host2.Addrs()[0], pstore.PermanentAddrTTL)
				host.Peerstore().AddAddr(host3.ID(), host3.Addrs()[0], pstore.PermanentAddrTTL)

				host.Connect(context.Background(), host.Peerstore().PeerInfo(host2.ID()))
				host.Connect(context.Background(), host.Peerstore().PeerInfo(host3.ID()))
			})

			g.AfterEach(func() {
				host.Close()
				host2.Close()
				host3.Close()
			})
		})

		g.Describe(".ip", func() {
			g.It("should return ip as 127.0.0.1", func() {
				ip := n.IP
				Expect(ip).ToNot(BeNil())
				Expect(ip.String()).To(Equal("127.0.0.1"))
			})
		})
	})
}
