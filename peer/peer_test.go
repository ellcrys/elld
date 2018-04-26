package peer

import (
	"context"
	"time"

	"github.com/ellcrys/druid/testutil"
	"github.com/ellcrys/druid/util"
	host "github.com/libp2p/go-libp2p-host"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	ma "github.com/multiformats/go-multiaddr"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Peer", func() {

	BeforeEach(func() {
		Expect(setTestCfg()).To(BeNil())
	})

	AfterEach(func() {
		Expect(removeTestCfgDir()).To(BeNil())
	})

	Describe(".NewPeer", func() {
		Context("address", func() {
			It("return err.Error('failed to parse address. Expects 'ip:port' format') when only port is provided", func() {
				_, err := NewPeer(cfg, "40100", 1, log)
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(Equal("failed to parse address. Expects 'ip:port' format"))
			})

			It("return err.Error('failed to parse address. Expects 'ip:port' format') when only ip is provided", func() {
				_, err := NewPeer(cfg, "127.0.0.1", 1, log)
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(Equal("failed to parse address. Expects 'ip:port' format"))
			})

			It("return err.Error('failed to create host > failed to parse ip4: 127.0.0 failed to parse ip4 addr: 127.0.0') when address is invalid and port is valid", func() {
				_, err := NewPeer(cfg, "127.0.0:40000", 1, log)
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(Equal("failed to create host > failed to parse ip4: 127.0.0 failed to parse ip4 addr: 127.0.0"))
			})

			It("return nil if address is ':40000'", func() {
				_, err := NewPeer(cfg, ":40000", 1, log)
				Expect(err).To(BeNil())
			})

			It("return nil if address is '127.0.0.1:40000'", func() {
				_, err := NewPeer(cfg, "127.0.0.1:40000", 1, log)
				Expect(err).To(BeNil())
			})
		})
	})

	Describe(".ID", func() {
		It("should return empty string when peer has no address", func() {
			p := Peer{}
			Expect(p.ID().Pretty()).To(Equal(""))
		})

		It("should return '12D3KooWL3XJ9EMCyZvmmGXL2LMiVBtrVa2BuESsJiXkSj7333Jw'", func() {
			p, err := NewPeer(cfg, "127.0.0.1:40000", 0, log)
			Expect(err).To(BeNil())
			Expect(p.ID().Pretty()).To(Equal("12D3KooWL3XJ9EMCyZvmmGXL2LMiVBtrVa2BuESsJiXkSj7333Jw"))
			p.Host().Close()
		})
	})

	Describe(".IDPretty", func() {
		It("should return empty string when peer has no address", func() {
			p := Peer{}
			Expect(p.StringID()).To(Equal(""))
		})

		It("should return '12D3KooWL3XJ9EMCyZvmmGXL2LMiVBtrVa2BuESsJiXkSj7333Jw'", func() {
			p, err := NewPeer(cfg, "127.0.0.1:40000", 0, log)
			Expect(err).To(BeNil())
			Expect(p.StringID()).To(Equal("12D3KooWL3XJ9EMCyZvmmGXL2LMiVBtrVa2BuESsJiXkSj7333Jw"))
			p.Host().Close()
		})
	})

	Describe(".IDShort", func() {
		It("should return empty string when peer has no address", func() {
			p := Peer{}
			Expect(p.ShortID()).To(Equal(""))
		})

		It("should return '12D3KooWL3XJ..JiXkSj7333Jw'", func() {
			p, err := NewPeer(cfg, "127.0.0.1:40000", 0, log)
			Expect(err).To(BeNil())
			Expect(p.ShortID()).To(Equal("12D3KooWL3XJ..JiXkSj7333Jw"))
			p.Host().Close()
		})
	})

	Describe(".PrivKey", func() {
		It("should return private key", func() {
			p, err := NewPeer(cfg, "127.0.0.1:40000", 0, log)
			Expect(err).To(BeNil())
			Expect(p.PrivKey()).NotTo(BeNil())
			p.Host().Close()
		})
	})

	Describe(".GetMultiAddr", func() {
		It("should return empty string when peer has no host", func() {
			p := Peer{}
			Expect(p.GetMultiAddr()).To(Equal(""))
		})

		It("should return '/ip4/127.0.0.1/tcp/40000/ipfs/12D3KooWL3XJ9EMCyZvmmGXL2LMiVBtrVa2BuESsJiXkSj7333Jw'", func() {
			p, err := NewPeer(cfg, "127.0.0.1:40000", 0, log)
			Expect(err).To(BeNil())
			Expect(p.GetMultiAddr()).To(Equal("/ip4/127.0.0.1/tcp/40000/ipfs/12D3KooWL3XJ9EMCyZvmmGXL2LMiVBtrVa2BuESsJiXkSj7333Jw"))
			p.Host().Close()
		})
	})

	Describe(".GetAddr", func() {
		It("should return '127.0.0.1:40000'", func() {
			p, err := NewPeer(cfg, "127.0.0.1:40000", 0, log)
			Expect(err).To(BeNil())
			Expect(p.GetAddr()).To(Equal("127.0.0.1:40000"))
			p.Host().Close()
		})
	})

	Describe(".PeerFromAddr", func() {
		It("should return error if address is not valid", func() {
			p, err := NewPeer(cfg, "127.0.0.1:40000", 0, log)
			Expect(err).To(BeNil())
			_, err = p.PeerFromAddr("/invalid", false)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("addr is not valid"))
		})
	})

	Describe(".IsBadTimestamp", func() {
		It("should return false when time is zero", func() {
			p, err := NewPeer(cfg, "127.0.0.1:40000", 0, log)
			p.Timestamp = time.Time{}
			Expect(err).To(BeNil())
			Expect(p.IsBadTimestamp()).To(BeTrue())
		})

		It("should return false when time 10 minutes, 1 second in the future", func() {
			p, err := NewPeer(cfg, "127.0.0.1:40000", 0, log)
			p.Timestamp = time.Now().Add(10*time.Minute + 1*time.Second)
			Expect(err).To(BeNil())
			Expect(p.IsBadTimestamp()).To(BeTrue())
		})

		It("should return false when time 3 hours, 1 second in the past", func() {
			p, err := NewPeer(cfg, "127.0.0.1:40000", 0, log)
			p.Timestamp = time.Now().Add(-3 * time.Hour)
			Expect(err).To(BeNil())
			Expect(p.IsBadTimestamp()).To(BeTrue())
		})
	})

	Describe(".GetIP4TCPAddr", func() {
		It("should return '/ip4/127.0.0.1/tcp/40000'", func() {
			p, err := NewPeer(cfg, "127.0.0.1:40000", 0, log)
			Expect(err).To(BeNil())
			Expect(p.GetIP4TCPAddr().String()).To(Equal("/ip4/127.0.0.1/tcp/40000"))
			p.Host().Close()
		})
	})

	Describe(".AddBootstrapPeers", func() {
		Context("with empty address", func() {
			It("peer manager's bootstrap list should be empty", func() {
				p, err := NewPeer(cfg, "127.0.0.1:40000", 0, log)
				Expect(err).To(BeNil())
				p.AddBootstrapPeers(nil, false)
				Expect(p.PM().GetBootstrapPeers()).To(HaveLen(0))
				p.Host().Close()
			})
		})

		Context("with invalid address", func() {
			It("peer manager's bootstrap list should not contain invalid address", func() {
				p, err := NewPeer(cfg, "127.0.0.1:40000", 0, log)
				Expect(err).To(BeNil())
				p.AddBootstrapPeers([]string{"/ip4/127.0.0.1/tcp/40000"}, false)
				Expect(p.PM().GetBootstrapPeers()).To(HaveLen(0))
				p.Host().Close()
			})

			It("peer manager's bootstrap list contain only one valid address", func() {
				p, err := NewPeer(cfg, "127.0.0.1:40000", 0, log)
				Expect(err).To(BeNil())
				p.AddBootstrapPeers([]string{
					"/ip4/127.0.0.1/tcp/40000",
					"/ip4/127.0.0.1/tcp/40000/ipfs/12D3KooWL3XJ9EMCyZvmmGXL2LMiVBtrVa2BuESsJiXkSj7333Jw",
				}, false)
				Expect(p.PM().GetBootstrapPeers()).To(HaveLen(1))
				Expect(p.PM().GetBootstrapPeer("12D3KooWL3XJ9EMCyZvmmGXL2LMiVBtrVa2BuESsJiXkSj7333Jw")).To(BeAssignableToTypeOf(&Peer{}))
				p.Host().Close()
			})
		})
	})

	Describe(".GetPeersPublicAddrs", func() {

		var host, host2, host3 host.Host
		var p *Peer
		var err error

		BeforeEach(func() {
			p, err = NewPeer(cfg, "127.0.0.1:40105", 5, log)
			Expect(err).To(BeNil())
			host = p.Host()
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
			peers := p.GetPeersPublicAddrs(nil)
			Expect(peers).To(HaveLen(2))
			expected, _ := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/40106")
			expected2, _ := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/40107")
			Expect(peers).To(ContainElement(expected))
			Expect(peers).To(ContainElement(expected2))
		})

		It("only /ip4/127.0.0.1/tcp/40107 must be returned", func() {
			peers := p.GetPeersPublicAddrs([]string{host2.ID().Pretty()})
			Expect(peers).To(HaveLen(1))
			expected, _ := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/40107")
			Expect(peers).To(ContainElement(expected))
		})

		AfterEach(func() {
			host.Close()
			host2.Close()
			host3.Close()
		})

		Context("without ignore list", func() {

		})
	})

	Describe(".Connected", func() {

		var host, host2 host.Host
		var p, p2 *Peer
		var err error

		BeforeEach(func() {
			p, err = NewPeer(cfg, "127.0.0.1:40106", 6, log)
			Expect(err).To(BeNil())
			p2, err = NewPeer(cfg, "127.0.0.1:40107", 7, log)
			Expect(err).To(BeNil())
			p2.SetLocalPeer(p)
			host = p.Host()
			Expect(err).To(BeNil())
			host2 = p2.Host()
			Expect(err).To(BeNil())

			host.SetStreamHandler("/protocol/0.0.1", testutil.NoOpStreamHandler)
			host.Peerstore().AddAddr(host2.ID(), host2.Addrs()[0], pstore.PermanentAddrTTL)
			host.Connect(context.Background(), host.Peerstore().PeerInfo(host2.ID()))
		})

		It("should return false when localPeer is nil", func() {
			p.localPeer = nil
			Expect(p.Connected()).To(BeFalse())
		})

		It("should return true when peer is connected", func() {
			lp, err := NewPeer(cfg, "127.0.0.1:40108", 8, log)
			Expect(err).To(BeNil())
			defer lp.host.Close()
			lpProtoc := NewInception(lp, log)

			rp, err := NewPeer(cfg, "127.0.0.1:40109", 9, log)
			Expect(err).To(BeNil())
			defer rp.host.Close()
			rpProtoc := NewInception(rp, log)
			rp.SetProtocolHandler(util.PingVersion, rpProtoc.OnPing)

			rp.localPeer = lp
			err = lpProtoc.sendPing(rp)
			Expect(err).To(BeNil())

			Expect(lp.PM().GetKnownPeer(rp.StringID()).Connected()).To(BeTrue())
			Expect(rp.PM().GetKnownPeer(lp.StringID()).Connected()).To(BeTrue())
		})

		AfterEach(func() {
			host.Close()
			host2.Close()
		})
	})

	Describe(".ip", func() {
		It("should return ip as 127.0.0.1", func() {
			p, err := NewPeer(cfg, "127.0.0.1:40106", 6, log)
			Expect(err).To(BeNil())
			ip := p.ip()
			Expect(ip).ToNot(BeNil())
			Expect(ip.String()).To(Equal("127.0.0.1"))
		})
	})
})
