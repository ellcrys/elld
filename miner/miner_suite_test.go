package miner

import (
	"github.com/ellcrys/elld/blockchain"
	"github.com/ellcrys/elld/blockchain/common"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/util/logger"
)

var log logger.Logger

func init() {
	log = logger.NewLogrusNoOp()
	log.SetToDebug()
}

func MakeTestBlock(bc types.BlockMaker, chain *blockchain.Chain, gp *types.GenerateBlockParams) types.Block {
	blk, err := bc.Generate(gp, &common.ChainerOp{Chain: chain})
	if err != nil {
		panic(err)
	}
	return blk
}
