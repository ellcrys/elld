package node_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/phayes/freeport"
	"github.com/shopspring/decimal"
	"github.com/thoas/go-funk"

	. "github.com/onsi/gomega"

	. "github.com/onsi/ginkgo"

	"github.com/ellcrys/elld/blockchain"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/elldb"
	"github.com/ellcrys/elld/node"
	"github.com/ellcrys/elld/params"
	"github.com/ellcrys/elld/testutil"
	"github.com/ellcrys/elld/txpool"
	"github.com/ellcrys/elld/util"

	"github.com/ellcrys/elld/util/logger"
)

var log = logger.NewLogrusNoOp()

// var makeBlock = func(bchain core.Blockchain, sender, receiver *crypto.Key, timestamp int64) core.Block {
// 	block, err := bchain.Generate(&core.GenerateBlockParams{
// 		Transactions: []core.Transaction{
// 			objects.NewTx(objects.TxTypeBalance, 1, util.String(sender.Addr()), sender, "0", "2.5", time.Now().UnixNano()),
// 		},
// 		Creator:           sender,
// 		Nonce:             core.EncodeNonce(1),
// 		Difficulty:        new(big.Int).SetInt64(131072),
// 		OverrideTimestamp: timestamp,
// 		AddFeeAlloc:       true,
// 	})
// 	if err != nil {
// 		panic(err)
// 	}
// 	return block
// }

// var makeBlockWithSingleTx = func(bchain core.Blockchain, sender, receiver *crypto.Key, timestamp int64, senderNonce uint64) core.Block {
// 	block, err := bchain.Generate(&core.GenerateBlockParams{
// 		Transactions: []core.Transaction{
// 			objects.NewTx(objects.TxTypeBalance, senderNonce, util.String(sender.Addr()), sender, "0", "2.5", time.Now().UnixNano()),
// 		},
// 		Creator:           sender,
// 		Nonce:             core.EncodeNonce(1),
// 		Difficulty:        new(big.Int).SetInt64(131072),
// 		OverrideTimestamp: timestamp,
// 		AddFeeAlloc:       true,
// 	})
// 	if err != nil {
// 		panic(err)
// 	}
// }

func getPort() int {
	port, err := freeport.GetFreePort()
	if err != nil {
		panic(err)
	}
	return port
}

// makeTestNode creates a node with
// a blockchain attached to it.
func makeTestNode(port int) *node.Node {
	return makeTestNodeWith(port, -1)
}

// makeTestNode creates a node with
// a blockchain attached to it.
func makeTestNodeWith(port int, seed int) *node.Node {
	cfg, err := testutil.SetTestCfg()
	if err != nil {
		panic(err)
	}

	db := elldb.NewDB(cfg.ConfigDir())
	err = db.Open(util.RandString(5))
	if err != nil {
		panic(err)
	}

	txPool := txpool.New(100)
	bc := blockchain.New(txPool, cfg, log)
	bc.SetDB(db)
	genesisBlock, err := blockchain.LoadBlockFromFile("genesis-test.json")
	Expect(err).To(BeNil())
	bc.SetGenesisBlock(genesisBlock)

	if seed < 0 {
		seed = funk.RandomInt(1, 5000000)
	}
	sk := crypto.NewKeyFromIntSeed(seed)
	n, err := node.NewNodeWithDB(db, cfg, fmt.Sprintf("127.0.0.1:%d", port), sk, log)
	if err != nil {
		panic(err)
	}

	gossip := node.NewGossip(n, log)
	n.SetGossipProtocol(gossip)
	n.SetBlockchain(bc)

	return n
}

func TestNodeSuite(t *testing.T) {
	params.FeePerByte = decimal.NewFromFloat(0.01)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Node Suite")
}

func closeNode(n *node.Node) {
	n.Host().Close()
	err := os.RemoveAll(n.GetCfg().ConfigDir())
	Expect(err).To(BeNil())
}
