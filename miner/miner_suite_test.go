package miner

import (
	"github.com/ellcrys/elld/blockchain"
	"github.com/ellcrys/elld/blockchain/common"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/util/logger"
)

var log logger.Logger

func init() {
	log = logger.NewLogrusNoOp()
	log.SetToDebug()
}

func MakeTestBlock(bc core.BlockMaker, chain *blockchain.Chain, gp *core.GenerateBlockParams) core.Block {
	blk, err := bc.Generate(gp, &common.ChainerOp{Chain: chain})
	if err != nil {
		panic(err)
	}
	return blk
}
