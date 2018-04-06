package main

import (
	"go.uber.org/zap"

	"github.com/ellcrys/gcoin/cmd"
)

var log *zap.SugaredLogger

func init() {
	log = NewLogger("/main")
}

func main() {
	log.Infof("gcoin node started")
	cmd.Execute()
}
