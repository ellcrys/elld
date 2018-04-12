package util

import (
	"fmt"
	"net"

	"github.com/libp2p/go-libp2p-host"

	inet "github.com/libp2p/go-libp2p-net"
	ma "github.com/multiformats/go-multiaddr"
)

// IsValidHostPortAddress checks if an address is a valid address matching
// the format `host:port`
func IsValidHostPortAddress(address string) bool {
	_, _, err := net.SplitHostPort(address)
	return err == nil
}

// IsValidAddress checks if a value is a valid multi address with ip4 and ipfs protocols
func IsValidAddress(addr string) bool {
	mAddr, err := ma.NewMultiaddr(addr)
	if err != nil {
		return false
	}

	protocols := mAddr.Protocols()
	if len(protocols) != 3 || protocols[0].Name != "ip4" || protocols[1].Name != "tcp" || protocols[2].Name != "ipfs" {
		return false
	}

	return true
}

// FullRemoteAddressFromStream returns the full peer multi address containing ip4, tcp and ipfs protocols
func FullRemoteAddressFromStream(s inet.Stream) ma.Multiaddr {
	if s == nil {
		return nil
	}
	ipfsAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", s.Conn().RemotePeer().Pretty()))
	fullAddr := s.Conn().RemoteMultiaddr().Encapsulate(ipfsAddr)
	return fullAddr
}

// FullRemoteAddressFromConn returns the full peer multi address containing ip4, tcp and ipfs protocols
func FullRemoteAddressFromConn(c inet.Conn) ma.Multiaddr {
	if c == nil {
		return nil
	}
	ipfsAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", c.RemotePeer().Pretty()))
	fullAddr := c.RemoteMultiaddr().Encapsulate(ipfsAddr)
	return fullAddr
}

// FullAddressFromHost returns the full peer multi address containing ip4, tcp and ipfs protocols
func FullAddressFromHost(host host.Host) ma.Multiaddr {
	if host == nil {
		return nil
	}
	if len(host.Addrs()) == 0 {
		return nil
	}
	ipfsAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", host.ID().Pretty()))
	fullAddr := host.Addrs()[0].Encapsulate(ipfsAddr)
	return fullAddr
}
