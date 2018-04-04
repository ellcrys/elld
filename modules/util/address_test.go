package util_test

import (
	"context"

	net "github.com/libp2p/go-libp2p-net"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	protocol "github.com/libp2p/go-libp2p-protocol"

	host "github.com/libp2p/go-libp2p-host"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/ellcrys/gcoin/modules/util"
	"github.com/ellcrys/gcoin/modules/util/testutil"
)

// NoOpStreamHandler accepts a stream and does nothing with it
var NoOpStreamHandler = func(s net.Stream) {}

var _ = Describe("Address", func() {

	Describe(".IsValidHostPortAddress", func() {
		It("should return false if params is 'abc'", func() {
			Expect(IsValidHostPortAddress("abc")).To(BeFalse())
		})

		It("should return true if params is '1.1.1.1:1234'", func() {
			Expect(IsValidHostPortAddress("1.1.1.1:1234")).To(BeTrue())
		})
	})

	Describe(".IsValidAddress", func() {

		It("Should return false for all cases ", func() {
			falseCases := []string{"ip4/1.1.1.1", "/ip4/1.1.1.1", "/ip4/1.1.1.1/tcp/1234", "/ip4/1.1.1.1/tcp/1234/ipfs"}
			for _, c := range falseCases {
				Expect(IsValidAddress(c)).To(BeFalse())
			}
		})

		It("Should return true if address is /ip4/1.1.1.1/tcp/1234/ipfs/12D3KooWKRyzVWW6ChFjQjK4miCty85Niy49tpPV95XdKu1BcvMA", func() {
			Expect(IsValidAddress("/ip4/1.1.1.1/tcp/1234/ipfs/12D3KooWKRyzVWW6ChFjQjK4miCty85Niy49tpPV95XdKu1BcvMA")).To(BeTrue())
		})
	})

	Describe(".FullRemoteAddressFromStream", func() {
		var host host.Host

		BeforeEach(func() {
			var err error
			host, err = testutil.RandomHost(0, 40100)
			Expect(err).To(BeNil())
		})

		It("should return nil if nil is passed", func() {
			addr := FullRemoteAddressFromStream(nil)
			Expect(addr).To(BeNil())
		})

		It("should return /ip4/127.0.0.1/tcp/40101/ipfs/12D3KooWE3AwZFT9zEWDUxhya62hmvEbRxYBWaosn7Kiqw5wsu73", func() {
			remoteHost, err := testutil.RandomHost(1234, 40101)
			Expect(err).To(BeNil())
			remoteHost.SetStreamHandler("/protocol/0.0.1", NoOpStreamHandler)

			host.Peerstore().AddAddr(remoteHost.ID(), remoteHost.Addrs()[0], pstore.PermanentAddrTTL)
			s, err := host.NewStream(context.Background(), remoteHost.ID(), protocol.ID("/protocol/0.0.1"))
			Expect(err).To(BeNil())
			defer host.Close()
			defer remoteHost.Close()

			addr := FullRemoteAddressFromStream(s)
			Expect(addr.String()).To(Equal("/ip4/127.0.0.1/tcp/40101/ipfs/12D3KooWE3AwZFT9zEWDUxhya62hmvEbRxYBWaosn7Kiqw5wsu73"))
		})
	})

	Describe(".FullAddressFromHost", func() {
		var host host.Host

		It("should return nil when nil is passed", func() {
			addr := FullAddressFromHost(nil)
			Expect(addr).To(BeNil())
		})

		It("should return /ip4/127.0.0.1/tcp/40102/ipfs/12D3KooWG7YTN3ADjgCqkxXMFQ5tdHUFVDGGU9tXfDHWUV4hUs42", func() {
			var err error
			host, err = testutil.RandomHost(12345, 40102)
			Expect(err).To(BeNil())
			addr := FullAddressFromHost(host)
			Expect(addr.String()).To(Equal("/ip4/127.0.0.1/tcp/40102/ipfs/12D3KooWG7YTN3ADjgCqkxXMFQ5tdHUFVDGGU9tXfDHWUV4hUs42"))
		})
	})
})
