package node

import (
	"fmt"
	"testing"

	"github.com/ellcrys/elld/blockchain"
	"github.com/ellcrys/elld/blockchain/store"
	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/elldb"
	"github.com/ellcrys/elld/testutil"
	"github.com/ellcrys/elld/txpool"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/util"

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

var testStore store.ChainStorer
var db, db2 elldb.DB
var lpBc, rpBc core.Blockchain
var chainID = util.String("chain1")
var txPool, txPool2 *txpool.TxPool
var sender, receiver *crypto.Key

var _ = Describe("Engine", func() {

	BeforeEach(func() {
		var err error
		cfg, err = testutil.SetTestCfg()
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		Expect(testutil.RemoveTestCfgDir()).To(BeNil())
	})

	BeforeEach(func() {
		var err error
		cfg, err = testutil.SetTestCfg()
		Expect(err).To(BeNil())
	})

	// Create the databases
	BeforeEach(func() {
		db = elldb.NewDB(cfg.ConfigDir())
		err = db.Open(util.RandString(5))
		Expect(err).To(BeNil())
		db2 = elldb.NewDB(cfg.ConfigDir())
		err = db2.Open(util.RandString(5))
		Expect(err).To(BeNil())
	})

	// Initialize the default test transaction pools
	// and create the blockchain instances and set their db
	BeforeEach(func() {
		txPool = txpool.NewTxPool(100)
		txPool2 = txpool.NewTxPool(100)
		lpBc = blockchain.New(txPool, cfg, log)
		rpBc = blockchain.New(txPool2, cfg, log)
		lpBc.SetDB(db)
		rpBc.SetDB(db2)
	})

	// Create test account keys
	BeforeEach(func() {
		sender = crypto.NewKeyFromIntSeed(1)
		receiver = crypto.NewKeyFromIntSeed(2)
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
		BlockTest,
	}

	for i, t := range tests {
		Describe(fmt.Sprintf("Test %d", i), func() {
			t()
		})
	}
})
