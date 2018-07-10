package types

import (
	"github.com/ellcrys/elld/util"
)

// Hex represents a hex encoded value
type Hex string

// Bytes returns the decoded hex value
func (h *Hex) Bytes() []byte {
	v, _ := util.FromHex(string(*h))
	return v
}

// FromByteToHex create a Hex instance from a byte slice
func FromByteToHex(b []byte) *Hex {
	hexStr := util.ToHex(b)
	hx := Hex(hexStr)
	return &hx
}
