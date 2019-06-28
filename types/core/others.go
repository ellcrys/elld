package core

import (
	"github.com/ellcrys/elld/util"
)

// ChainInfo describes a chain
type ChainInfo struct {
	ID                util.String `json:"id" msgpack:"json"`
	ParentChainID     util.String `json:"parentChainID" msgpack:"parentChainID"`
	ParentBlockNumber uint64      `json:"parentBlockNumber" msgpack:"parentBlockNumber"`
	Timestamp         int64       `json:"timestamp" msgpack:"timestamp"`
}

// GetID returns the ID
func (c *ChainInfo) GetID() util.String {
	return c.ID
}

// GetParentChainID returns the ID
func (c *ChainInfo) GetParentChainID() util.String {
	return c.ParentChainID
}

// GetParentBlockNumber returns the parent block number
func (c *ChainInfo) GetParentBlockNumber() uint64 {
	return c.ParentBlockNumber
}

// GetTimestamp returns the timestamp
func (c *ChainInfo) GetTimestamp() int64 {
	return c.Timestamp
}

// ArgGetMinedBlock represents arguments for fetching mined blocks
type ArgGetMinedBlock struct {
	Limit         int    `mapstructure:"limit"`
	LastHash      string `mapstructure:"lastHash"`
	CreatorPubKey string `mapstructure:"creatorPubKey"`
}
