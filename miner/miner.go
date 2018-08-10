package miner

import (
	"github.com/ellcrys/elld/blockchain/common"
	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/util/logger"
)

// Miner provides mining, block header modification and
// validation capabilities with respect to PoW. The miner
// leverages Ethash to performing PoW computation.
type Miner struct {

	// cfg is the miner configuration
	cfg *config.MinerConfig

	// log is the logger for the miner
	log logger.Logger

	// blockMaker provides functions for creating a block
	blockMaker common.BlockMaker
}

// New creates and returns a new Miner instance
func New(blockMaker common.Blockchain, cfg *config.MinerConfig, log logger.Logger) *Miner {
	return &Miner{
		cfg:        cfg,
		log:        log,
		blockMaker: blockMaker,
	}
}
