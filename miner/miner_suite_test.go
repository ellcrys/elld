package miner

import (
	"testing"

	"github.com/ellcrys/elld/blockchain"
	"github.com/ellcrys/elld/blockchain/common"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/util/logger"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var log logger.Logger

func TestBlockchain(t *testing.T) {
	log = logger.NewLogrusNoOp()
	log.SetToDebug()
	RegisterFailHandler(Fail)
	RunSpecs(t, "Miner Suite")
}

func MakeTestBlock(bc core.BlockMaker, chain *blockchain.Chain, gp *core.GenerateBlockParams) core.Block {
	blk, err := bc.Generate(gp, &common.ChainerOp{Chain: chain})
	if err != nil {
		panic(err)
	}
	return blk
}
