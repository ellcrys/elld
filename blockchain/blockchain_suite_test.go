package blockchain

import (
	"fmt"
	"math/big"
	"testing"
	"time"

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
var bc *Blockchain
var chainID = util.String("chain1")
var genesisChain *Chain
var genesisBlock core.Block
var txPool *txpool.TxPool
var sender, receiver *crypto.Key

func TestBlockchain(t *testing.T) {
	log = logger.NewLogrusNoOp()
	RegisterFailHandler(Fail)
	RunSpecs(t, "Blockchain Suite")
}

func MakeTestBlock(bc core.Blockchain, chain *Chain, gp *core.GenerateBlockParams) core.Block {
	blk, err := bc.Generate(gp, ChainOp{Chain: chain})
	if err != nil {
		panic(err)
	}
	return blk
}

var makeBlock = func(ch *Chain) core.Block {
	return MakeTestBlock(bc, ch, &core.GenerateBlockParams{
		Transactions: []core.Transaction{
			objects.NewTx(objects.TxTypeAlloc, 123, util.String(sender.Addr()), sender, "1", "0.1", time.Now().UnixNano()),
		},
		Creator:    sender,
		Nonce:      core.EncodeNonce(1),
		Difficulty: new(big.Int).SetInt64(131072),
	})
}

var makeBlockWithBalanceTx = func(ch *Chain) core.Block {
	return MakeTestBlock(bc, ch, &core.GenerateBlockParams{
		Transactions: []core.Transaction{
			objects.NewTx(objects.TxTypeBalance, 123, util.String(receiver.Addr()), sender, "1", "0.1", time.Now().UnixNano()),
		},
		Creator:    sender,
		Nonce:      core.EncodeNonce(1),
		Difficulty: new(big.Int).SetInt64(131072),
	})
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
		txPool = txpool.NewTxPool(100)
		bc = New(txPool, cfg, log)
		bc.SetDB(db)
		bc.SetGenesisBlock(GenesisBlock)
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
		genesisChain = NewChain(chainID, db, cfg, log)
		bc.addChain(genesisChain)
		bc.bestChain = genesisChain
		err = bc.saveChain(genesisChain, "", 0)
		Expect(err).To(BeNil())
	})

	// create test accounts here
	BeforeEach(func() {
		Expect(bc.putAccount(1, genesisChain, &objects.Account{
			Type:    objects.AccountTypeBalance,
			Address: util.String(sender.Addr()),
			Balance: "1000",
		})).To(BeNil())
	})

	BeforeEach(func() {
		genesisBlock = MakeTestBlock(bc, genesisChain, &core.GenerateBlockParams{
			Transactions: []core.Transaction{
				objects.NewTx(objects.TxTypeBalance, 123, util.String(receiver.Addr()), sender, "1", "0.1", 1532730722),
			},
			Creator:           sender,
			Nonce:             core.EncodeNonce(1),
			Difficulty:        new(big.Int).SetInt64(131072),
			OverrideTimestamp: time.Now().Add(-2 * time.Second).Unix(),
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
		WorldReaderTest,
		BlockchainTest,
		ReOrgTest,
		ChainTest,
		ProcessTest,
		BlockTest,
		CacheTest,
		TransactionValidatorTest,
		BlockValidatorTest,
		ChainTransverserTest,
	}

	Describe(fmt.Sprintf("Tests"), func() {
		for _, t := range tests {
			t()
		}
	})
})
