package peermanager_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/olebedev/emitter"

	"github.com/ellcrys/elld/blockchain"
	"github.com/ellcrys/elld/blockchain/txpool"
	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/elldb"
	"github.com/ellcrys/elld/node"
	"github.com/ellcrys/elld/node/peermanager"
	"github.com/ellcrys/elld/params"
	"github.com/ellcrys/elld/testutil"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/util/logger"

	"github.com/imdario/mergo"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/phayes/freeport"
	"github.com/shopspring/decimal"
	funk "github.com/thoas/go-funk"
)

var log = logger.NewLogrusNoOp()

func init() {
	params.FeePerByte = decimal.NewFromFloat(0.01)
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
	bc.SetNodeKey(crypto.NewKeyFromIntSeed(1234))
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
	n.SetLastSeen(time.Now())

	n.SetTxsPool(txp)
	n.SetBlockchain(bc)

	return n
}

func closeNode(n *node.Node) {
	go n.GetHost().Close()
	err := os.RemoveAll(n.GetCfg().DataDir())
	Expect(err).To(BeNil())
}

func emptyNode(n *node.Node) *node.Node {
	n2 := node.NewAlmostEmptyNode()
	if n != nil {
		mergo.MergeWithOverwrite(n2, n)
	}
	return n2
}

func emptyNodeWithLastSeenTime(t time.Time) *node.Node {
	n := node.NewAlmostEmptyNode()
	n.SetLastSeen(t)
	return n
}

func NewMgr(cfg *config.EngineConfig, localNode *node.Node) *peermanager.Manager {
	mgr := peermanager.NewManager(cfg, localNode, log)
	return mgr
}

func TestPeermanager(t *testing.T) {
	params.FeePerByte = decimal.NewFromFloat(0.01)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Peermanager Suite")
}
