package util_test

import (
	"context"
	"net"

	peer "github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	protocol "github.com/libp2p/go-libp2p-protocol"
	ma "github.com/multiformats/go-multiaddr"

	host "github.com/libp2p/go-libp2p-host"
	. "github.com/onsi/gomega"

	"github.com/ellcrys/elld/testutil"
	. "github.com/ellcrys/elld/util"
	. "github.com/onsi/ginkgo"
)

var _ = Describe("Address", func() {

	Describe(".IsValidHostPortAddress", func() {
		It("should return false if params is 'abc'", func() {
			Expect(IsValidHostPortAddress("abc")).To(BeFalse())
		})

		It("should return true if params is '1.1.1.1:1234'", func() {
			Expect(IsValidHostPortAddress("1.1.1.1:1234")).To(BeTrue())
		})
	})

	Describe(".IsValidFullMultiaddr", func() {

		It("Should return false for all cases ", func() {
			falseCases := []string{"ip4/1.1.1.1", "/ip4/1.1.1.1",
				"/ip4/1.1.1.1/tcp/1234", "/ip4/1.1.1.1/tcp/1234/ipfs"}
			for _, c := range falseCases {
				Expect(IsValidAddr(c)).To(BeFalse())
			}
		})

		It("Should return true if address is valid", func() {
			addr := "/ip4/1.1.1.1/tcp/1234/ipfs/12D3KooWKRyzVWW6ChFjQjK4miCty85Niy49tpPV95XdKu1BcvMA"
			Expect(IsValidAddr(addr)).To(BeTrue())
		})
	})

	Describe(".RemoteAddrFromStream", func() {
		var host host.Host

		BeforeEach(func() {
			var err error
			host, err = testutil.RandomHost(0, 40100)
			Expect(err).To(BeNil())
		})

		It("should return get address from stream", func() {
			remoteHost, err := testutil.RandomHost(1234, 40101)
			Expect(err).To(BeNil())
			defer remoteHost.Close()
			remoteHost.SetStreamHandler("/protocol/0.0.1", testutil.NoOpStreamHandler)

			host.Peerstore().AddAddr(remoteHost.ID(), remoteHost.Addrs()[0], pstore.PermanentAddrTTL)
			s, err := host.NewStream(context.Background(), remoteHost.ID(), protocol.ID("/protocol/0.0.1"))
			Expect(err).To(BeNil())
			defer host.Close()

			addr := RemoteAddrFromStream(s)
			expected := "/ip4/127.0.0.1/tcp/40101/ipfs/12D3KooWE3AwZFT9zEWDUxhya62hmvEbRxYBWaosn7Kiqw5wsu73"
			Expect(addr.String()).To(Equal(expected))
		})
	})

	Describe(".ValidateAndResolveConnString", func() {

		It("should return error if connection string is not valid", func() {
			_, err := ValidateAndResolveConnString("ellcrys12D3KooWSpFL@google.com>9000")
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("connection string is not valid"))
		})

		It("should resolve domain to IP", func() {
			cs, err := ValidateAndResolveConnString("ellcrys://12D3KooWSpgHpR83HeYYqYkRKc7mVuT3vcG919fsUsebLwVqghFL@google.com:9000")
			Expect(err).To(BeNil())
			Expect(IsValidConnectionString(cs)).To(BeTrue())
			connStrParts := ParseConnString(cs)
			Expect(net.ParseIP(connStrParts.Address).To4()).ToNot(BeNil())
		})
	})

	Describe(".RemoteAddrFromConn", func() {
		var host host.Host

		BeforeEach(func() {
			var err error
			host, err = testutil.RandomHost(0, 40102)
			Expect(err).To(BeNil())
		})

		It("should return /ip4/127.0.0.1/tcp/40103/ipfs/12D3KooWE3AwZFT9zEWDUxhya62hmvEbRxYBWaosn7Kiqw5wsu73", func() {
			remoteHost, err := testutil.RandomHost(1234, 40103)
			Expect(err).To(BeNil())
			remoteHost.SetStreamHandler("/protocol/0.0.1", testutil.NoOpStreamHandler)
			defer remoteHost.Close()

			host.Peerstore().AddAddr(remoteHost.ID(), remoteHost.Addrs()[0], pstore.PermanentAddrTTL)
			s, err := host.NewStream(context.Background(), remoteHost.ID(), protocol.ID("/protocol/0.0.1"))
			Expect(err).To(BeNil())
			defer host.Close()

			addr := RemoteAddrFromConn(s.Conn())
			Expect(addr.String()).To(Equal("/ip4/127.0.0.1/tcp/40103/ipfs/12D3KooWE3AwZFT9zEWDUxhya62hmvEbRxYBWaosn7Kiqw5wsu73"))
		})
	})

	Describe(".AddressFromHost", func() {
		var host host.Host

		It("should return empty address when nil is passed", func() {
			addr := AddressFromHost(nil)
			Expect(addr).To(Equal(NodeAddr("")))
		})

		It("should return /ip4/127.0.0.1/tcp/40104/ipfs/12D3KooWG7YTN3ADjgCqkxXMFQ5tdHUFVDGGU9tXfDHWUV4hUs42", func() {
			var err error
			host, err = testutil.RandomHost(12345, 40104)
			Expect(err).To(BeNil())
			addr := AddressFromHost(host)
			Expect(addr.String()).To(Equal("/ip4/127.0.0.1/tcp/40104/ipfs/12D3KooWG7YTN3ADjgCqkxXMFQ5tdHUFVDGGU9tXfDHWUV4hUs42"))
			host.Close()
		})
	})

	Describe(".IDFromAddr", func() {
		It("should return 12D3KooWG7YTN3ADjgCqkxXMFQ5tdHUFVDGGU9tXfDHWUV4hUs42", func() {
			addr, _ := ma.NewMultiaddr("/ip4/127.0.0.1/tcp/40104/ipfs/12D3KooWG7YTN3ADjgCqkxXMFQ5tdHUFVDGGU9tXfDHWUV4hUs42")
			Expect(IDFromAddr(addr).Pretty()).To(Equal("12D3KooWG7YTN3ADjgCqkxXMFQ5tdHUFVDGGU9tXfDHWUV4hUs42"))
		})
	})

	Describe(".IDFromAddrString", func() {
		It("should return empty string if address is invalid", func() {
			addr := IDFromAddrString("/ip427.0.0.1/tcp40104/ipfs/12D3KooWG7YTN3ADjgCqkxXMFQ5tdHUFVDGGU9tXfDHWUV4hUs42")
			Expect(addr.Pretty()).To(Equal(""))
		})

		It("should return 12D3KooWG7YTN3ADjgCqkxXMFQ5tdHUFVDGGU9tXfDHWUV4hUs42", func() {
			addr := IDFromAddrString("/ip4/127.0.0.1/tcp/40104/ipfs/12D3KooWG7YTN3ADjgCqkxXMFQ5tdHUFVDGGU9tXfDHWUV4hUs42")
			Expect(addr.Pretty()).To(Equal("12D3KooWG7YTN3ADjgCqkxXMFQ5tdHUFVDGGU9tXfDHWUV4hUs42"))
		})
	})

	Describe(".ParseAddr", func() {
		It("should return tcp = 40104, ip4 = 127.0.0.1, ip6= and ipfs=12D3KooWG7YTN3ADjgCqkxXMFQ5tdHUFVDGGU9tXfDHWUV4hUs42", func() {
			result := ParseAddr("/ip4/127.0.0.1/tcp/40104/ipfs/12D3KooWG7YTN3ADjgCqkxXMFQ5tdHUFVDGGU9tXfDHWUV4hUs42")
			Expect(result).To(Equal(map[string]string{
				"tcp":  "40104",
				"ip4":  "127.0.0.1",
				"ip6":  "",
				"ipfs": "12D3KooWG7YTN3ADjgCqkxXMFQ5tdHUFVDGGU9tXfDHWUV4hUs42",
			}))
		})

		It("should return nil when addr is invalid", func() {
			result := ParseAddr("/ip427.0.0.1/tcp04/ipfs2D3KooWG7YTN3ADjgCqkxXMFQ5tdHUFVDGGU9tXfDHWUV4hUs42")
			Expect(result).To(BeNil())
		})
	})

	Describe(".GetIPFromAddr", func() {
		It("should return ip4 ip", func() {
			ip := GetIPFromAddr("/ip4/127.0.0.1/tcp/40104/ipfs/12D3KooWG7YTN3ADjgCqkxXMFQ5tdHUFVDGGU9tXfDHWUV4hUs42")
			Expect(ip).ToNot(BeNil())
			Expect(ip.String()).To(Equal("127.0.0.1"))
		})

		It("should return ip6 ip", func() {
			ip := GetIPFromAddr("/ip6/::1/tcp/40104/ipfs/12D3KooWG7YTN3ADjgCqkxXMFQ5tdHUFVDGGU9tXfDHWUV4hUs42")
			Expect(ip).ToNot(BeNil())
			Expect(ip.String()).To(Equal("::1"))
		})

		It("should return nil when addr is invalid", func() {
			ip := GetIPFromAddr("/ip6::1tcp/40104/ipfs/12D3KooWG7YTN3ADjgCqkxXMFQ5tdHUFVDGGU9tXfDHWUV4hUs42")
			Expect(ip).To(BeNil())
		})
	})

	Describe(".ShortID", func() {
		It("should return empty string", func() {
			Expect(ShortID(peer.ID(""))).To(Equal(""))
		})

		It("should return 'CovLVG4fQcqR..oMt32Q6LgZDK'", func() {
			Expect(ShortID(peer.ID("12D3KooWG7YTN3ADjgCqkxXMFQ5tdHUFVDGGU9tXfDHWUV4hUs42"))).To(Equal("CovLVG4fQcqR..oMt32Q6LgZDK"))
		})
	})

	Describe(".IsRoutableAddr", func() {
		It("should return false when addr is not a valid multiaddr", func() {
			valid := IsRoutableAddr("invalid_addr")
			Expect(valid).To(BeFalse())
		})

		It("should return false when addr is '/ip4/0.0.0.0/tcp/40104/ipfs/12D3KooWG7YTN3ADjgCqkxXMFQ5tdHUFVDGGU9tXfDHWUV4hUs42'", func() {
			addr := "/ip4/0.0.0.0/tcp/40104/ipfs/12D3KooWG7YTN3ADjgCqkxXMFQ5tdHUFVDGGU9tXfDHWUV4hUs42"
			valid := IsRoutableAddr(addr)
			Expect(valid).To(BeFalse())
		})

		It("should return false when addr is '/ip4/127.0.0.1/tcp/40104/ipfs/12D3KooWG7YTN3ADjgCqkxXMFQ5tdHUFVDGGU9tXfDHWUV4hUs42'", func() {
			addr := "/ip4/127.0.0.1/tcp/40104/ipfs/12D3KooWG7YTN3ADjgCqkxXMFQ5tdHUFVDGGU9tXfDHWUV4hUs42"
			valid := IsRoutableAddr(addr)
			Expect(valid).To(BeFalse())
		})

		It("should return false when addr is '/ip6/::ffff:abcd:ef12:1/tcp/40104/ipfs/12D3KooWG7YTN3ADjgCqkxXMFQ5tdHUFVDGGU9tXfDHWUV4hUs42'", func() {
			addr := "/ip6/::1/tcp/40104/ipfs/12D3KooWG7YTN3ADjgCqkxXMFQ5tdHUFVDGGU9tXfDHWUV4hUs42"
			valid := IsRoutableAddr(addr)
			Expect(valid).To(BeFalse())
		})
	})

	Describe(".IsValidConnectionString", func() {

		When("scheme is 'ell'", func() {
			It("should return true for a valid connection string", func() {
				str := "ell://12D3KooWQJKY8C35U2JCZLaobGueSJnXGcx6Z9FStDXxQKtnhtwy@127.0.0.1:9000"
				Expect(IsValidConnectionString(str)).To(BeTrue())
			})
		})

		When("scheme is 'ellcrys'", func() {
			It("should return true for a valid connection string", func() {
				str := "ellcrys://12D3KooWQJKY8C35U2JCZLaobGueSJnXGcx6Z9FStDXxQKtnhtwy@127.0.0.1:9000"
				Expect(IsValidConnectionString(str)).To(BeTrue())
			})
		})

		It("should return false for a valid connection string", func() {
			str := "mysql://12D3KooWQJKY8C35U2JCZLaobGueSJnXGcx6Z9FStDXxQKtnhtwy@127.0.0.1:9000"
			Expect(IsValidConnectionString(str)).To(BeFalse())
		})

		It("should return false for a valid connection string", func() {
			str := "mysql://12D3KooWQJKY8C35U2JCZLaobGueSJnXGcx6Z9FStDXxQKtnhtwy@127.0.0.1:900a"
			Expect(IsValidConnectionString(str)).To(BeFalse())
		})
	})

	Describe(".ParseConnectionString", func() {
		It("should return expected connection string data", func() {
			str := "ell://12D3KooWQJKY8C35U2JCZLaobGueSJnXGcx6Z9FStDXxQKtnhtwy@127.0.0.1:9000"
			expected := &ConnStringData{
				ID:      "12D3KooWQJKY8C35U2JCZLaobGueSJnXGcx6Z9FStDXxQKtnhtwy",
				Address: "127.0.0.1",
				Port:    "9000",
			}
			Expect(ParseConnString(str)).To(Equal(expected))
		})

		It("should return nil if connection string is invalid", func() {
			str := "stuff://"
			Expect(ParseConnString(str)).To(BeNil())
		})
	})

	Describe(".AddressFromConnectionString", func() {
		It("should return empty NodeAddr if connection string is invalid", func() {
			str := "stuff://"
			Expect(AddressFromConnString(str)).To(Equal(NodeAddr("")))
		})

		Context("using an ip4 address", func() {
			It("should return a valid NodeAddr", func() {
				str := "ell://12D3KooWQJKY8C35U2JCZLaobGueSJnXGcx6Z9FStDXxQKtnhtwy@127.0.0.1:9000"
				expected := NodeAddr("/ip4/127.0.0.1/tcp/9000/ipfs/12D3KooWQJKY8C35U2JCZLaobGueSJnXGcx6Z9FStDXxQKtnhtwy")
				addr := AddressFromConnString(str)
				Expect(addr).To(Equal(expected))
				Expect(addr.IsValid()).To(BeTrue())
			})
		})

		Context("using an ip6 address", func() {
			It("should return a valid NodeAddr", func() {
				str := "ell://12D3KooWQJKY8C35U2JCZLaobGueSJnXGcx6Z9FStDXxQKtnhtwy@[2607:f0d0:1002:0051:0000:0000:0000:0004]:80"
				expected := NodeAddr("/ip6/2607:f0d0:1002:51::4/tcp/80/ipfs/12D3KooWQJKY8C35U2JCZLaobGueSJnXGcx6Z9FStDXxQKtnhtwy")
				addr := AddressFromConnString(str)
				Expect(addr).To(Equal(expected))
				Expect(addr.IsValid()).To(BeTrue())
			})
		})
	})

})
