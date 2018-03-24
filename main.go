package main

import (
	"go.uber.org/zap"

	"github.com/ellcrys/garagecoin/cmd"
	"github.com/ellcrys/garagecoin/modules"
)

var log *zap.SugaredLogger

func init() {
	log = modules.NewLogger("/main")
}

func main() {
	log.Infof("Garagecoin node started")
	cmd.Execute()
}
