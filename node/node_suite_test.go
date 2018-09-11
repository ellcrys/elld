package node

import (
	"context"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/ellcrys/elld/blockchain"
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

var log = logger.NewLogrusNoOp()
var cfg *config.EngineConfig
var err error

func closeNode(n *Node) {
	n.Host().ConnManager().TrimOpenConns(context.Background())
}

var makeBlock = func(bchain core.Blockchain) core.Block {
	block, err := bchain.Generate(&core.GenerateBlockParams{
		Transactions: []core.Transaction{
			objects.NewTx(objects.TxTypeAlloc, 123, util.String(sender.Addr()), sender, "0", "0", time.Now().UnixNano()),
		},
		Creator:    sender,
		Nonce:      core.EncodeNonce(1),
		Difficulty: new(big.Int).SetInt64(131072),
	})
	if err != nil {
		panic(err)
	}
	return block
}

func TestPeer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Peer Suite")
}

var testStore core.ChainStorer
var db, db2 elldb.DB
var lpBc, rpBc core.Blockchain
var chainID = util.String("chain1")
var txPool, txPool2 *txpool.TxPool
var sender = crypto.NewKeyFromIntSeed(1)
var receiver = crypto.NewKeyFromIntSeed(2)

var _ = Describe("Engine", func() {

	BeforeEach(func() {
		var err error
		cfg, err = testutil.SetTestCfg()
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		err = os.RemoveAll(cfg.ConfigDir())
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
		txPool = txpool.New(100)
		lpBc = blockchain.New(txPool, cfg, log)
		lpBc.SetDB(db)
		lpBc.SetGenesisBlock(blockchain.GenesisBlock)

		txPool2 = txpool.New(100)
		rpBc = blockchain.New(txPool2, cfg, log)
		rpBc.SetDB(db2)
		rpBc.SetGenesisBlock(blockchain.GenesisBlock)
	})

	BeforeEach(func() {
		err = lpBc.Up()
		Expect(err).To(BeNil())
		// Expect(lpBc.CreateAccount(1, lpBc.GetBestChain(), &objects.Account{
		// 	Type:    objects.AccountTypeBalance,
		// 	Address: util.String(sender.Addr()),
		// 	Balance: "1000",
		// })).To(BeNil())

		err = rpBc.Up()
		Expect(err).To(BeNil())
		// Expect(rpBc.CreateAccount(1, rpBc.GetBestChain(), &objects.Account{
		// 	Type:    objects.AccountTypeBalance,
		// 	Address: util.String(sender.Addr()),
		// 	Balance: "1000",
		// })).To(BeNil())
	})

	var tests = []func() bool{
		HandshakeTest,
		TransactionTest,
		AddrTest,
		// GetAddrTest,
		// SelfAdvTest,
		// PingTest,
		// PeerManagerTest,
		// NodeTest,
		// BlockTest,
	}

	for _, t := range tests {
		t()
	}
})
