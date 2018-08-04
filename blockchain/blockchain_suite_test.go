package blockchain

import (
	"fmt"
	"testing"

	"github.com/ellcrys/elld/blockchain/common"
	"github.com/ellcrys/elld/blockchain/leveldb"
	"github.com/ellcrys/elld/blockchain/testdata"
	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/database"
	"github.com/ellcrys/elld/testutil"
	"github.com/ellcrys/elld/txpool"
	"github.com/ellcrys/elld/util/logger"
	"github.com/ellcrys/elld/wire"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var log logger.Logger
var cfg *config.EngineConfig
var err error
var store common.Store
var db database.DB
var bc *Blockchain
var chainID = "chain1"
var chain *Chain
var block *wire.Block
var txPool *txpool.TxPool
var sender, receiver *crypto.Key

func TestBlockchain(t *testing.T) {
	log = logger.NewLogrusNoOp()
	RegisterFailHandler(Fail)
	RunSpecs(t, "Blockchain Suite")
}

var _ = Describe("Blockchain", func() {

	BeforeEach(func() {
		var err error
		cfg, err = testutil.SetTestCfg()
		Expect(err).To(BeNil())
	})

	// Create the database and store instances
	BeforeEach(func() {
		db = database.NewLevelDB(cfg.ConfigDir())
		err = db.Open("")
		Expect(err).To(BeNil())
		store, err = leveldb.New(db)
		Expect(err).To(BeNil())
	})

	// Initialize the default test transaction pool
	// and create the blockchain. Also set the store
	// on the blockchain.
	BeforeEach(func() {
		txPool = txpool.NewTxPool(100)
		bc = New(txPool, cfg, log)
		bc.SetStore(store)
	})

	// Create default test block
	// and test account keys
	BeforeEach(func() {
		block = testdata.GenesisBlock
		sender = crypto.NewKeyFromIntSeed(1)
		receiver = crypto.NewKeyFromIntSeed(2)
	})

	// create default test chain and add it to
	// the blockchain. Also append the test block
	// to the chain
	BeforeEach(func() {
		chain = NewChain(chainID, store, cfg, log)
		Expect(err).To(BeNil())
		bc.addChain(chain)
		err = chain.append(block)
		Expect(err).To(BeNil())
		bc.bestChain = chain
	})

	// create test accounts here
	BeforeEach(func() {
		err = bc.putAccount(1, chain, &wire.Account{
			Type:    wire.AccountTypeBalance,
			Address: sender.Addr(),
			Balance: "1000",
		})
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		db.Close()
	})

	AfterEach(func() {
		Expect(testutil.RemoveTestCfgDir()).To(BeNil())
	})

	var tests = []func() bool{
		BlockchainTest,
		AccountTest,
		CacheTest,
		ChainTest,
		MetadataTest,
		ProcessTest,
		BlockTest,
		TransactionValidatorTest,
		BlockValidatorTest,
	}

	for i, t := range tests {
		Describe(fmt.Sprintf("Test %d", i), func() {
			t()
		})
	}
})
