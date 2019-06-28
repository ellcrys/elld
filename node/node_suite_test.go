package node

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/ellcrys/mother/blockchain"
	"github.com/ellcrys/mother/blockchain/txpool"
	"github.com/ellcrys/mother/crypto"
	"github.com/ellcrys/mother/elldb"
	"github.com/ellcrys/mother/params"
	"github.com/ellcrys/mother/testutil"
	"github.com/ellcrys/mother/util"
	"github.com/ellcrys/mother/util/logger"
	"github.com/olebedev/emitter"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/phayes/freeport"
	"github.com/shopspring/decimal"
	funk "github.com/thoas/go-funk"
)

var log = logger.NewLogrusNoOp()

func TestNode(t *testing.T) {
	params.FeePerByte = decimal.NewFromFloat(0.01)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Node Suite")
}

func getPort() int {
	port, err := freeport.GetFreePort()
	if err != nil {
		panic(err)
	}
	return port
}

// makeTestNode creates a node with
// a blockchain attached to it.
func makeTestNode(port int) *Node {
	return makeTestNodeWith(port, -1)
}

// makeTestNode creates a node with
// a blockchain attached to it.
func makeTestNodeWith(port int, seed int) *Node {

	cfg, err := testutil.SetTestCfg()
	if err != nil {
		panic(err)
	}

	db := elldb.NewDB(cfg.NetDataDir())
	err = db.Open(util.RandString(5))
	if err != nil {
		panic(err)
	}

	evtEmitter := &emitter.Emitter{}
	txp := txpool.New(100)

	bc := blockchain.New(txp, cfg, log)
	bc.SetEventEmitter(evtEmitter)
	bc.SetDB(db)
	genesisBlock, err := blockchain.LoadBlockFromFile("genesis-test.json")
	Expect(err).To(BeNil())
	bc.SetGenesisBlock(genesisBlock)

	if seed < 0 {
		seed = funk.RandomInt(1, 5000000)
	}

	sk := crypto.NewKeyFromIntSeed(seed)
	n, err := NewNodeWithDB(db, cfg, fmt.Sprintf("127.0.0.1:%d", port), sk, log)
	if err != nil {
		panic(err)
	}
	n.SetLastSeen(time.Now())

	n.SetTxsPool(txp)
	n.SetBlockchain(bc)

	return n
}

func closeNode(n *Node) {
	go n.GetHost().Close()
	err := os.RemoveAll(n.GetCfg().DataDir())
	Expect(err).To(BeNil())
}
