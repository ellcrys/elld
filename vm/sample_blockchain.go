package vm

import "github.com/ellcrys/druid/blockcode"

// SampleBlockchain defines a structure of a blockchain.
// This is for test only. Blockchains are more complex than this.
type SampleBlockchain struct {
	blockcodes map[string]*blockcode.Blockcode
}

// NewSampleBlockchain creates a SampleBlockchain
func NewSampleBlockchain() *SampleBlockchain {
	b := new(SampleBlockchain)
	bc, err := blockcode.FromDir("./testdata/blockcode_example")
	if err != nil {
		panic(err)
	}
	b.blockcodes = map[string]*blockcode.Blockcode{
		"some_address": bc,
	}
	return b
}

// GetBlockCode returns a blockcode
func (b *SampleBlockchain) GetBlockCode(address string) *blockcode.Blockcode {
	return b.blockcodes[address]
}
