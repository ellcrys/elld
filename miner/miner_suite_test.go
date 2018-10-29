package miner

import (
	"github.com/ellcrys/elld/util/logger"
)

var log logger.Logger

func init() {
	log = logger.NewLogrusNoOp()
	log.SetToDebug()
}
