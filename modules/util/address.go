package util

import (
	"net"
)

// IsValidHostPortAddress checks if an address is a valid address matching
// the format `host:port`
func IsValidHostPortAddress(address string) bool {
	_, _, err := net.SplitHostPort(address)
	return err == nil
}
