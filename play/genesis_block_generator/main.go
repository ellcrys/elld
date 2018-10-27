package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"time"

	"github.com/ellcrys/elld/blockchain"
	"github.com/ellcrys/elld/blockchain/common"
	"github.com/ellcrys/elld/blockchain/txpool"
	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/elldb"
	"github.com/ellcrys/elld/params"
	"github.com/ellcrys/elld/testutil"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/types/core"

	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/util/logger"
)

var cfg *config.EngineConfig
var err error
var log logger.Logger

func init() {
	cfg, err = testutil.SetTestCfg()
	if err != nil {
		panic(err)
	}
	log = logger.NewLogrusNoOp()
}

func main() {
	defer os.RemoveAll(cfg.DataDir())

	maxTxs := flag.Int64("numTxs", 1000, "number of transaction")
	difficulty := flag.Int64("diff", params.GenesisDifficulty.Int64(), "Block difficulty")
	flag.Parse()

	// create temporary database
	db := elldb.NewDB(cfg.DataDir())
	err = db.Open(util.RandString(5))
	if err != nil {
		panic(err)
	}

	// create blockchain
	txPool := txpool.New(100)
	bc := blockchain.New(txPool, cfg, log)
	bc.SetDB(db)

	// create random creator key
	creator, _ := crypto.NewKey(nil)

	// generate some allocation transactions
	var txs = []types.Transaction{}
	var addrsPrivateKey = make(map[string]string)

	for i := int64(1); i < *maxTxs+1; i++ {
		recipient := crypto.NewKeyFromIntSeed(int(i))
		allocTx := core.NewTx(core.TxTypeAlloc, 0, util.String(recipient.Addr()), creator, "100", "0", time.Now().Unix())
		txs = append(txs, allocTx)
		addrsPrivateKey[recipient.Addr()] = recipient.PrivKey().Base58()
	}

	params := &types.GenerateBlockParams{
		Transactions:            txs,
		Creator:                 creator,
		Nonce:                   util.EncodeNonce(1),
		Difficulty:              new(big.Int).SetInt64(*difficulty),
		OverrideTotalDifficulty: new(big.Int).SetInt64(*difficulty),
		OverrideTimestamp:       time.Now().Unix(),
	}

	// create an empty chain
	genesisChain := blockchain.NewChain("genesis", db, cfg, log)

	// generate block
	block, err := bc.Generate(params, &common.ChainerOp{Chain: genesisChain})
	if err != nil {
		panic(err)
	}

	// validate
	// bv := blockchain.NewBlockValidator(block, txPool, bc, cfg, log)

	// write to file
	bs, _ := json.Marshal(block)
	if err = ioutil.WriteFile("genesis.json", bs, 0644); err != nil {
		panic(err)
	}

	// write address keys
	bs, _ = json.Marshal(addrsPrivateKey)
	if err = ioutil.WriteFile("keys.json", bs, 0644); err != nil {
		panic(err)
	}

	fmt.Println("Generated!")
	fmt.Println("Block is located in `genesis.json`")
	fmt.Println("Allocated accounts keys located in `keys.json`")
}
