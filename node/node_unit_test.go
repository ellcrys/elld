package node_test

import (
	"context"
	"fmt"
	"time"

	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/node"
	"github.com/ellcrys/elld/testutil"
	host "github.com/libp2p/go-libp2p-host"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	ma "github.com/multiformats/go-multiaddr"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Node Unit Test", func() {

	var n *node.Node
	var cfg *config.EngineConfig
	var lpPort int

	BeforeEach(func() {
		lpPort = getPort()
		n = makeTestNodeWith(lpPort, 0)
		cfg = n.GetCfg()
	})

	AfterEach(func() {
		closeNode(n)
	})

	Describe(".NewNode", func() {
		Context("address", func() {
			It("return err.Error('failed to parse address. Expects 'ip:port' format') when only port is provided", func() {
				_, err := node.NewNode(cfg, "40100", crypto.NewKeyFromIntSeed(1), log)
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(Equal("failed to parse address. Expects 'ip:port' format"))
			})

			It("return err.Error('failed to parse address. Expects 'ip:port' format') when only ip is provided", func() {
				_, err := node.NewNode(cfg, "127.0.0.1", crypto.NewKeyFromIntSeed(1), log)
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(Equal("failed to parse address. Expects 'ip:port' format"))
			})

			It("return err.Error('failed to create host > failed to parse ip4: 127.0.0 failed to parse ip4 addr: 127.0.0') when address is invalid and port is valid", func() {
				_, err := node.NewNode(cfg, "127.0.0:40000", crypto.NewKeyFromIntSeed(1), log)
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(Equal("failed to create host > failed to parse ip4: 127.0.0 failed to parse ip4 addr: 127.0.0"))
			})

			It("return nil if address is ':40000'", func() {
				_, err := node.NewNode(cfg, ":40000", crypto.NewKeyFromIntSeed(1), log)
				Expect(err).To(BeNil())
			})

			It("return nil if address is '127.0.0.1:40000'", func() {
				_, err := node.NewNode(cfg, "127.0.0.1:40000", crypto.NewKeyFromIntSeed(1), log)
				Expect(err).To(BeNil())
			})
		})
	})

	Describe(".ID", func() {
		It("should return empty string when node has no address", func() {
			p := node.Node{}
			Expect(p.ID().Pretty()).To(Equal(""))
		})

		It("should return '12D3KooWL3XJ9EMCyZvmmGXL2LMiVBtrVa2BuESsJiXkSj7333Jw'", func() {
			Expect(n.ID().Pretty()).To(Equal("12D3KooWL3XJ9EMCyZvmmGXL2LMiVBtrVa2BuESsJiXkSj7333Jw"))
			closeNode(n)
		})
	})

	Describe(".IDPretty", func() {
		It("should return empty string when node has no address", func() {
			p := node.Node{}
			Expect(p.StringID()).To(Equal(""))
		})

		It("should return '12D3KooWL3XJ9EMCyZvmmGXL2LMiVBtrVa2BuESsJiXkSj7333Jw'", func() {
			Expect(n.StringID()).To(Equal("12D3KooWL3XJ9EMCyZvmmGXL2LMiVBtrVa2BuESsJiXkSj7333Jw"))
			closeNode(n)
		})
	})

	Describe(".IDShort", func() {
		It("should return empty string when node has no address", func() {
			p := node.Node{}
			Expect(p.ShortID()).To(Equal(""))
		})

		It("should return '12D3KooWL3XJ..JiXkSj7333Jw'", func() {
			Expect(n.ShortID()).To(Equal("12D3KooWL3XJ..JiXkSj7333Jw"))
			closeNode(n)
		})
	})

	Describe(".PrivKey", func() {
		It("should return private key", func() {
			Expect(n.PrivKey()).NotTo(BeNil())
			closeNode(n)
		})
	})

	Describe(".GetMultiAddr", func() {
		It("should return empty string when node has no host", func() {
			n := node.Node{}
			Expect(n.GetMultiAddr()).To(Equal(""))
		})

		// It("should return '/ip4/127.0.0.1/tcp/40000/ipfs/12D3KooWL3XJ9EMCyZvmmGXL2LMiVBtrVa2BuESsJiXkSj7333Jw'", func() {
		// 	Expect(n.GetMultiAddr()).To(Equal("/ip4/127.0.0.1/tcp/40000/ipfs/12D3KooWL3XJ9EMCyZvmmGXL2LMiVBtrVa2BuESsJiXkSj7333Jw"))
		// 	closeNode(n)
		// })
	})

	Describe(".GetAddr", func() {
		It("should return expected address'", func() {
			Expect(n.GetAddr()).To(Equal(fmt.Sprintf("127.0.0.1:%d", lpPort)))
			closeNode(n)
		})
	})

	Describe(".NodeFromAddr", func() {
		It("should return error if address is not valid", func() {
			_, err := n.NodeFromAddr("/invalid", false)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("invalid address provided"))
		})
	})

	Describe(".IsBadTimestamp", func() {
		It("should return false when time is zero", func() {
			n.SetTimestamp(time.Time{})
			Expect(n.IsBadTimestamp()).To(BeTrue())
		})

		It("should return false when time 10 minutes, 1 second in the future", func() {
			n.SetTimestamp(time.Now().Add(10*time.Minute + 1*time.Second))
			Expect(n.IsBadTimestamp()).To(BeTrue())
		})

		It("should return false when time 3 hours, 1 second in the past", func() {
			n.SetTimestamp(time.Now().Add(-3 * time.Hour))
			Expect(n.IsBadTimestamp()).To(BeTrue())
		})
	})

	Describe(".GetIP4TCPAddr", func() {
		It("should return expected address", func() {
			Expect(n.GetIP4TCPAddr().String()).To(Equal(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", lpPort)))
		})
	})

	Describe(".AddBootstrapNodes", func() {
		Context("with empty address", func() {
			It("peer manager's bootstrap list should be empty", func() {
				n.AddBootstrapNodes(nil, false)
				Expect(n.PM().GetBootstrapNodes()).To(HaveLen(0))
			})
		})

		Context("with invalid address", func() {
			It("peer manager's bootstrap list should not contain invalid address", func() {
				n.AddBootstrapNodes([]string{"/ip4/127.0.0.1/tcp/40000"}, false)
				Expect(n.PM().GetBootstrapNodes()).To(HaveLen(0))
			})

			It("peer manager's bootstrap list contain only one valid address", func() {
				n.AddBootstrapNodes([]string{
					"/ip4/127.0.0.1/tcp/40000",
					"/ip4/127.0.0.1/tcp/40000/ipfs/12D3KooWL3XJ9EMCyZvmmGXL2LMiVBtrVa2BuESsJiXkSj7333Jw",
				}, false)
				Expect(n.PM().GetBootstrapNodes()).To(HaveLen(1))
				Expect(n.PM().GetBootstrapPeer("12D3KooWL3XJ9EMCyZvmmGXL2LMiVBtrVa2BuESsJiXkSj7333Jw")).To(BeAssignableToTypeOf(&node.Node{}))
			})
		})
	})

	Describe(".PeersPublicAddrs", func() {

		var host, host2, host3 host.Host
		var n *node.Node
		var err error

		BeforeEach(func() {
			n, err = node.NewNode(cfg, "127.0.0.1:40105", crypto.NewKeyFromIntSeed(5), log)
			Expect(err).To(BeNil())
			host = n.Host()
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

		It("the peers (/ip4/127.0.0.1/tcp/40106, /ip4/127.0.0.1/tcp/40107) must be returned", func() {
			peers := n.PeersPublicAddr(nil)
			Expect(peers).To(HaveLen(2))
			expected, _ := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/40106")
			expected2, _ := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/40107")
			Expect(peers).To(ContainElement(expected))
			Expect(peers).To(ContainElement(expected2))
		})

		It("only /ip4/127.0.0.1/tcp/40107 must be returned", func() {
			peers := n.PeersPublicAddr([]string{host2.ID().Pretty()})
			Expect(peers).To(HaveLen(1))
			expected, _ := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/40107")
			Expect(peers).To(ContainElement(expected))
		})

		AfterEach(func() {
			host.Close()
			host2.Close()
			host3.Close()
		})
	})

	Describe(".Connected", func() {

		var n, n2 *node.Node
		var err error

		BeforeEach(func() {
			n, err = node.NewNode(cfg, "127.0.0.1:40106", crypto.NewKeyFromIntSeed(6), log)
			Expect(err).To(BeNil())
			n2, err = node.NewNode(cfg, "127.0.0.1:40107", crypto.NewKeyFromIntSeed(7), log)
			Expect(err).To(BeNil())
			n2.SetLocalNode(n)
		})

		It("should return false when localPeer is nil", func() {
			n.SetLocalNode(nil)
			Expect(n.Connected()).To(BeFalse())
		})

		It("should return true when peer is connected", func() {
			n.GetHost().Peerstore().AddAddr(n2.GetHost().ID(), n2.GetHost().Addrs()[0], pstore.PermanentAddrTTL)
			n.GetHost().Connect(context.Background(), n.GetHost().Peerstore().PeerInfo(n2.GetHost().ID()))
			Expect(n2.Connected()).To(BeTrue())
		})

		AfterEach(func() {
			closeNode(n)
			closeNode(n2)
		})
	})

	Describe(".ip", func() {
		It("should return ip as 127.0.0.1", func() {
			ip := n.IP
			Expect(ip).ToNot(BeNil())
			Expect(ip.String()).To(Equal("127.0.0.1"))
		})
	})
})
