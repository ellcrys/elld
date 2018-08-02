package node

import (
	"fmt"
	"testing"

	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/testutil"

	"github.com/ellcrys/elld/util/logger"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var log = logger.NewLogrusNoOp()
var cfg *config.EngineConfig
var err error

func TestPeer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Peer Suite")
}

var _ = Describe("Engine", func() {

	BeforeEach(func() {
		var err error
		cfg, err = testutil.SetTestCfg()
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		Expect(testutil.RemoveTestCfgDir()).To(BeNil())
	})

	var tests = []func() bool{
		TransactionTest,
		AddrTest,
		GetAddrTest,
		TransactionSessionTest,
		SelfAdvTest,
		PingTest,
		PeerManagerTest,
		NodeTest,
		HandshakeTest,
	}

	for i, t := range tests {
		Describe(fmt.Sprintf("Test %d", i), func() {
			t()
		})
	}
})
