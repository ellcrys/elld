package validators

import (
	"fmt"

	"github.com/ellcrys/elld/blockchain"
	"github.com/ellcrys/elld/blockchain/common"
	"github.com/ellcrys/elld/blockchain/leveldb"
	"github.com/ellcrys/elld/blockchain/testdata"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/database"
	"github.com/ellcrys/elld/testutil"

	"github.com/ellcrys/elld/wire"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Transaction", func() {

	var err error
	var store common.Store
	var bchain *blockchain.Blockchain
	var db database.DB
	var genesisBlock *wire.Block

	BeforeEach(func() {
		var err error
		cfg, err = testutil.SetTestCfg()
		Expect(err).To(BeNil())
	})

	BeforeEach(func() {
		db = database.NewLevelDB(cfg.ConfigDir())
		err = db.Open("")
		Expect(err).To(BeNil())
	})

	BeforeEach(func() {
		store, err = leveldb.New(db)
		Expect(err).To(BeNil())
	})

	BeforeEach(func() {
		blockchain.GenesisBlock = testdata.TestGenesisBlock
		genesisBlock, _ = wire.BlockFromString(blockchain.GenesisBlock)
		err = blockchain.SeedTestGenesisState(store, genesisBlock)
		Expect(err).To(BeNil())
	})

	BeforeEach(func() {
		bchain = blockchain.New(cfg, log)
		bchain.SetStore(store)
		err = bchain.Up()
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		db.Close()
	})

	AfterEach(func() {
		Expect(testutil.RemoveTestCfgDir()).To(BeNil())
	})

	Describe(".statelessChecks", func() {
		It("should check for validation errors", func() {
			var cases = map[*wire.Block]interface{}{
				nil:           fmt.Errorf("nil block"),
				&wire.Block{}: fmt.Errorf("field:header, error:header is required"),
				&wire.Block{}: fmt.Errorf("field:hash, error:hash is required"),
				&wire.Block{Hash: "invalid", Header: &wire.Header{}}: fmt.Errorf("field:hash, error:hash is not correct"),
				&wire.Block{}:                                                                                        fmt.Errorf("field:sig, error:signature is required"),
				&wire.Block{Header: &wire.Header{}}:                                                                  fmt.Errorf("field:header.parentHash, error:parent hash is required"),
				&wire.Block{Header: &wire.Header{}}:                                                                  fmt.Errorf("field:header.number, error:number must be greater or equal to 1"),
				&wire.Block{Header: &wire.Header{}}:                                                                  fmt.Errorf("field:header.number, error:number must be greater or equal to 1"),
				&wire.Block{Header: &wire.Header{}}:                                                                  fmt.Errorf("field:header.creatorPubKey, error:creator's public key is required"),
				&wire.Block{Header: &wire.Header{}}:                                                                  fmt.Errorf("field:header.transactionsRoot, error:transaction root is required"),
				&wire.Block{Header: &wire.Header{}}:                                                                  fmt.Errorf("field:header.stateRoot, error:state root is required"),
				&wire.Block{Header: &wire.Header{}}:                                                                  fmt.Errorf("field:header.mixHash, error:mix hash is required"),
				&wire.Block{Header: &wire.Header{}}:                                                                  fmt.Errorf("field:header.difficulty, error:difficulty must be non-zero and non-negative"),
				&wire.Block{Header: &wire.Header{}}:                                                                  fmt.Errorf("field:header.timestamp, error:timestamp must not be greater or equal to 1"),
				&wire.Block{Header: &wire.Header{}}:                                                                  fmt.Errorf("field:transactions, error:at least one transaction is required"),
				&wire.Block{Header: &wire.Header{}, Transactions: []*wire.Transaction{&wire.Transaction{Type: 109}}}: fmt.Errorf("tx:0, field:type, error:unsupported transaction type"),
			}
			for b, err := range cases {
				validator := NewBlockValidator(b, nil, nil, false)
				errs := validator.statelessChecks()
				Expect(errs).To(ContainElement(err))
			}
		})
	})

	Describe(".statelessChecks", func() {
		It("should check for validation errors", func() {
			key := crypto.NewKeyFromIntSeed(1)
			var cases = map[*wire.Block]interface{}{
				&wire.Block{Header: &wire.Header{}}:                                     fmt.Errorf("field:creatorPubKey, error:empty pub key"),
				&wire.Block{Header: &wire.Header{CreatorPubKey: "invalid"}}:             fmt.Errorf("field:creatorPubKey, error:invalid format: version and/or checksum bytes missing"),
				&wire.Block{Header: &wire.Header{CreatorPubKey: key.PubKey().Base58()}}: fmt.Errorf("field:sig, error:signature is not valid"),
			}
			for b, err := range cases {
				validator := NewBlockValidator(b, nil, nil, false)
				errs := validator.checkSignature()
				Expect(errs).To(ContainElement(err))
			}
		})
	})

	Describe(".Validate", func() {
		It("should return if block and a transaction in the block exist", func() {
			validator := NewBlockValidator(genesisBlock, nil, bchain, true)
			errs := validator.Validate()
			Expect(errs).To(ContainElement(fmt.Errorf("error:block is already known")))
			Expect(errs).To(ContainElement(fmt.Errorf("tx:0, error:transaction already exist in main chain")))
		})
	})
})
