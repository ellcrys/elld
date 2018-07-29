package node

import (
	"testing"

	"github.com/ellcrys/elld/config"

	"github.com/ellcrys/elld/util/logger"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var log = logger.NewLogrusNoOp()
var cfg *config.EngineConfig

func TestPeer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Peer Suite")
}
