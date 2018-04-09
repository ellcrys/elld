package main

import (
	"go.uber.org/zap"

	"github.com/ellcrys/gcoin/cmd"
	"github.com/ellcrys/gcoin/util"
)

var log *zap.SugaredLogger

func init() {
	log = util.NewLogger("/main")
}

func main() {
	log.Infof("gcoin node started")
	cmd.Execute()
}
