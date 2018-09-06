package miner

import (
	"fmt"
	"testing"

	"github.com/olebedev/emitter"

	"github.com/ellcrys/elld/blockchain"
	"github.com/ellcrys/elld/blockchain/common"
	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/elldb"
	"github.com/ellcrys/elld/testutil"
	"github.com/ellcrys/elld/txpool"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/types/core/objects"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/util/logger"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var log logger.Logger
var cfg *config.EngineConfig
var err error
var testStore core.ChainStorer
var db elldb.DB
var bc *blockchain.Blockchain
var chainID = util.String("chain1")
var genesisChain *blockchain.Chain
var genesisBlock *objects.Block
var txPool *txpool.TxPool
var sender, receiver *crypto.Key
var event *emitter.Emitter

func TestBlockchain(t *testing.T) {
	log = logger.NewLogrus()
	log.SetToDebug()
	RegisterFailHandler(Fail)
	RunSpecs(t, "Blockchain Suite")
}

func MakeTestBlock(bc core.BlockMaker, chain *blockchain.Chain, gp *core.GenerateBlockParams) core.Block {
	blk, err := bc.Generate(gp, &common.ChainerOp{Chain: chain})
	if err != nil {
		panic(err)
	}
	return blk
}

var _ = Describe("Blockchain", func() {

	BeforeEach(func() {
		var err error
		cfg, err = testutil.SetTestCfg()
		Expect(err).To(BeNil())
	})

	// Create the database and store instances
	BeforeEach(func() {
		db = elldb.NewDB(cfg.ConfigDir())
		err = db.Open("")
		Expect(err).To(BeNil())
	})

	// Initialize the default test transaction pool
	// and create the blockchain. Also set the store
	// on the blockchain.
	BeforeEach(func() {
		txPool = txpool.New(100)
		event = &emitter.Emitter{}
		bc = blockchain.New(txPool, cfg, log)
		bc.SetDB(db)
		bc.SetEventEmitter(event)
	})

	// Create default test block
	// and test account keys
	BeforeEach(func() {
		sender = crypto.NewKeyFromIntSeed(1)
		receiver = crypto.NewKeyFromIntSeed(2)
	})

	BeforeEach(func() {
		err = bc.Up()
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		db.Close()
	})

	AfterEach(func() {
		Expect(testutil.RemoveTestCfgDir()).To(BeNil())
	})

	var tests = []func() bool{
		MinerTest,
	}

	for i, t := range tests {
		Describe(fmt.Sprintf("Test %d", i), func() {
			t()
		})
	}
})
