package logic

import (
	"testing"

	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/util/logger"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var log = logger.NewLogrusNoOp()
var cfg *config.EngineConfig

func TestLogic(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Logic Suite")
}
