package blockchain

import (
	"testing"

	"github.com/ellcrys/elld/configdir"
	"github.com/ellcrys/elld/util/logger"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var cfg *configdir.Config
var log logger.Logger

func TestBlockchain(t *testing.T) {
	log = logger.NewLogrusNoOp()
	RegisterFailHandler(Fail)
	RunSpecs(t, "Blockchain Suite")
}
