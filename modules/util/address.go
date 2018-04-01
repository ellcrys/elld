package util

import (
	"net"

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
