package blockchain

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/ellcrys/elld/blockchain/common"
	"github.com/ellcrys/elld/blockchain/store"
	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/elldb"
	"github.com/ellcrys/elld/testutil"
	"github.com/ellcrys/elld/txpool"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/util/logger"
	"github.com/ellcrys/elld/wire"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var log logger.Logger
var cfg *config.EngineConfig
var err error
var testStore store.Storer
var db elldb.DB
var bc *Blockchain
var chainID = "chain1"
var genesisChain *Chain
var genesisBlock *wire.Block
var txPool *txpool.TxPool
var sender, receiver *crypto.Key

func TestBlockchain(t *testing.T) {
	log = logger.NewLogrusNoOp()
	RegisterFailHandler(Fail)
	RunSpecs(t, "Blockchain Suite")
}

func makeTestBlock(bc common.Blockchain, chain *Chain, gp *common.GenerateBlockParams) *wire.Block {
	blk, err := bc.GenerateBlock(gp, ChainOp{Chain: chain})
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
		testStore, err = store.New(db)
		Expect(err).To(BeNil())
	})

	// Initialize the default test transaction pool
	// and create the blockchain. Also set the store
	// on the blockchain.
	BeforeEach(func() {
		txPool = txpool.NewTxPool(100)
		bc = New(txPool, cfg, log)
		bc.SetStore(testStore)
	})

	// Create default test block
	// and test account keys
	BeforeEach(func() {
		sender = crypto.NewKeyFromIntSeed(1)
		receiver = crypto.NewKeyFromIntSeed(2)
	})

	// create default test chain and add it to
	// the blockchain. Also append the test block
	// to the chain
	BeforeEach(func() {
		genesisChain = NewChain(chainID, testStore, cfg, log)
		bc.addChain(genesisChain)
		bc.bestChain = genesisChain
	})

	// create test accounts here
	BeforeEach(func() {
		Expect(bc.putAccount(1, genesisChain, &wire.Account{
			Type:    wire.AccountTypeBalance,
			Address: sender.Addr(),
			Balance: "1000",
		})).To(BeNil())
	})

	BeforeEach(func() {
		genesisBlock = makeTestBlock(bc, genesisChain, &common.GenerateBlockParams{
			Transactions: []*wire.Transaction{
				wire.NewTx(wire.TxTypeBalance, 123, receiver.Addr(), sender, "1", "0.1", 1532730722),
			},
			Creator:    sender,
			Nonce:      wire.EncodeNonce(1),
			MixHash:    util.BytesToHash([]byte("mix hash")),
			Difficulty: new(big.Int).SetInt64(500),
		})
		err = genesisChain.append(genesisBlock)
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
		ChainTest,
		ProcessTest,
		BlockTest,
		AccountTest,
		CacheTest,
		MetadataTest,
		TransactionValidatorTest,
		BlockValidatorTest,
	}

	for i, t := range tests {
		Describe(fmt.Sprintf("Test %d", i), func() {
			t()
		})
	}
})
