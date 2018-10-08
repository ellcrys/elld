package util

import (
	"fmt"
	"net"

	host "github.com/libp2p/go-libp2p-host"
	peer "github.com/libp2p/go-libp2p-peer"

	inet "github.com/libp2p/go-libp2p-net"
	ma "github.com/multiformats/go-multiaddr"
)

// IsValidHostPortAddress checks if an address
// is a valid address matching
// the format `host:port`
func IsValidHostPortAddress(address string) bool {
	_, _, err := net.SplitHostPort(address)
	return err == nil
}

// IsValidAddr checks if an address is a valid
// multi address with ip4, tcp, and ipfs protocols
func IsValidAddr(addr string) bool {
	mAddr, err := ma.NewMultiaddr(addr)
	if err != nil {
		return false
	}

	protocols := mAddr.Protocols()
	if len(protocols) != 3 || (protocols[0].Name != "ip4" &&
		protocols[0].Name != "ip6") || protocols[1].Name != "tcp" ||
		protocols[2].Name != "ipfs" {
		return false
	}

	return true
}

// IsRoutableAddr checks if an addr
// is valid and routable
func IsRoutableAddr(addr string) bool {
	maddr, err := ma.NewMultiaddr(addr)
	if err != nil {
		return false
	}
	ip, _ := maddr.ValueForProtocol(ma.P_IP6)
	if ip == "" {
		ip, _ = maddr.ValueForProtocol(ma.P_IP4)
	}
	return IsRoutable(net.ParseIP(ip))
}

// RemoteAddrFromStream gets the remote
// address from the given stream
func RemoteAddrFromStream(s inet.Stream) NodeAddr {
	ipfsAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", s.Conn().RemotePeer().Pretty()))
	fullAddr := s.Conn().RemoteMultiaddr().Encapsulate(ipfsAddr)
	return NodeAddr(fullAddr.String())
}

// RemoteAddrFromConn gets the remote address
// from the given connection
func RemoteAddrFromConn(c inet.Conn) NodeAddr {
	ipfsAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", c.RemotePeer().Pretty()))
	fullAddr := c.RemoteMultiaddr().Encapsulate(ipfsAddr)
	return NodeAddr(fullAddr.String())
}

// IDFromAddr extracts and returns the peer ID
func IDFromAddr(addr ma.Multiaddr) peer.ID {
	pid, _ := addr.ValueForProtocol(ma.P_IPFS)
	id, _ := peer.IDB58Decode(pid)
	return id
}

// IDFromAddrString is like IDFromAddr but accepts a string.
// Returns empty string if addr is not a valid multiaddr.
// Expects the caller to have validated addr before calling the function.
func IDFromAddrString(addr string) peer.ID {
	mAddr, err := ma.NewMultiaddr(addr)
	if err != nil {
		return ""
	}
	pid, _ := mAddr.ValueForProtocol(ma.P_IPFS)
	id, _ := peer.IDB58Decode(pid)
	return id
}

// ParseAddr returns the protocol and value present in a multiaddr.
// Expects the caller to have validate the address before calling the function.
func ParseAddr(addr string) map[string]string {
	mAddr, err := ma.NewMultiaddr(addr)
	if err != nil {
		return nil
	}

	tcp, _ := mAddr.ValueForProtocol(ma.P_TCP)
	ip4, _ := mAddr.ValueForProtocol(ma.P_IP4)
	ip6, _ := mAddr.ValueForProtocol(ma.P_IP6)
	ipfs, _ := mAddr.ValueForProtocol(ma.P_IPFS)

	return map[string]string{
		"tcp":  tcp,
		"ip4":  ip4,
		"ip6":  ip6,
		"ipfs": ipfs,
	}
}

// GetIPFromAddr get the IP4/6 ip of the address.
// Expects the caller to have validate the addr
func GetIPFromAddr(addr string) net.IP {
	addrParsed := ParseAddr(addr)
	if addrParsed == nil {
		return nil
	}

	ip := addrParsed["ip6"]
	if ip == "" {
		ip = addrParsed["ip4"]
	}

	return net.ParseIP(ip)
}

// ShortID returns the short version an ID
func ShortID(id peer.ID) string {
	address := id.Pretty()

	if address == "" {
		return ""
	}

	return address[0:12] + ".." + address[40:52]
}

// NodeAddr represents address that points
// to a node on a network. The address are
// represented as Multiaddr.
type NodeAddr string

// AddressFromHost gets address of an host
func AddressFromHost(host host.Host) NodeAddr {
	if host == nil {
		return ""
	}
	if len(host.Addrs()) == 0 {
		return ""
	}
	ipfsAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", host.ID().Pretty()))
	fullAddr := host.Addrs()[0].Encapsulate(ipfsAddr)
	return NodeAddr(fullAddr.String())
}

// IsValid checks whether the address
// is valid Multiaddr
func (a NodeAddr) IsValid() bool {
	return IsValidAddr(string(a))
}

// GetMultiaddr gets the address. It will panic
// if the address is not a valid Multiaddr
func (a NodeAddr) GetMultiaddr() ma.Multiaddr {
	if !a.IsValid() {
		panic(fmt.Errorf("invalid multiaddress"))
	}
	mAddr, _ := ma.NewMultiaddr(string(a))
	return mAddr
}

func (a NodeAddr) String() string {
	return string(a)
}

// Equal checks whether a and b match
func (a NodeAddr) Equal(b NodeAddr) bool {
	return a.GetMultiaddr().String() ==
		b.GetMultiaddr().String()
}

// ID gets the address unique ID
func (a NodeAddr) ID() peer.ID {
	return IDFromAddr(a.GetMultiaddr())
}

// StringID the address unique ID as a string
func (a NodeAddr) StringID() string {
	return IDFromAddr(a.GetMultiaddr()).Pretty()
}

// IsRoutable checks whether the
// address is routable
func (a NodeAddr) IsRoutable() bool {
	return IsRoutableAddr(string(a))
}

// IP gets the IP from the address
func (a NodeAddr) IP() net.IP {
	return GetIPFromAddr(string(a))
}

// DecapIPFS gets the address without the
// IPFS part
func (a NodeAddr) DecapIPFS() ma.Multiaddr {
	ipfsPart := fmt.Sprintf("/ipfs/%s", a.ID().Pretty())
	ipfsAddr, _ := ma.NewMultiaddr(ipfsPart)
	return a.GetMultiaddr().Decapsulate(ipfsAddr)
}
