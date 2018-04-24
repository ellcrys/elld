package peer

import (
	"testing"

	"github.com/ellcrys/druid/util/logger"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var log = logger.NewLogrusNoOp()

func TestPeer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Peer Suite")
}
