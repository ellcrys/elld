package node

import (
	"testing"

	"github.com/ellcrys/elld/configdir"

	"github.com/ellcrys/elld/util/logger"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var log = logger.NewLogrusNoOp()
var cfg *configdir.Config

func TestPeer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Peer Suite")
}
