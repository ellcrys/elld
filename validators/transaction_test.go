package validators

import (
	"fmt"
	"time"

	"github.com/ellcrys/elld/blockchain"
	"github.com/ellcrys/elld/blockchain/common"
	"github.com/ellcrys/elld/blockchain/leveldb"
	"github.com/ellcrys/elld/blockchain/testdata"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/database"
	"github.com/ellcrys/elld/testutil"
	"github.com/ellcrys/elld/txpool"
	"github.com/ellcrys/elld/util"

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

	Describe(".checkFormatAndValue", func() {

		var validator *TxsValidator

		It("should return error if tx = nil", func() {
			errs := validator.checkFormatAndValue(nil)
			Expect(errs).To(HaveLen(1))
			Expect(errs).To(ContainElement(fmt.Errorf("nil tx")))
		})

		It("should test all validation rules", func() {
			var cases = map[*wire.Transaction]interface{}{
				&wire.Transaction{Type: 0}:                                       fmt.Errorf("index:0, field:type, error:unsupported transaction type"),
				&wire.Transaction{Nonce: -1}:                                     fmt.Errorf("index:0, field:nonce, error:nonce must be non-negative"),
				&wire.Transaction{}:                                              fmt.Errorf("index:0, field:to, error:recipient address is required"),
				&wire.Transaction{To: "invalid"}:                                 fmt.Errorf("index:0, field:to, error:recipient address is not valid"),
				&wire.Transaction{To: "invalid"}:                                 fmt.Errorf("index:0, field:to, error:recipient address is not valid"),
				&wire.Transaction{From: "invalid"}:                               fmt.Errorf("index:0, field:from, error:sender address is not valid"),
				&wire.Transaction{From: "invalid"}:                               fmt.Errorf("index:0, field:from, error:sender address is not valid"),
				&wire.Transaction{}:                                              fmt.Errorf("index:0, field:senderPubKey, error:sender public key is required"),
				&wire.Transaction{SenderPubKey: "invalid"}:                       fmt.Errorf("index:0, field:senderPubKey, error:sender public key is not valid"),
				&wire.Transaction{Type: wire.TxTypeBalance}:                      fmt.Errorf("index:0, field:value, error:value is required"),
				&wire.Transaction{Type: wire.TxTypeBalance, Value: "1oo"}:        fmt.Errorf("index:0, field:value, error:could not convert to decimal"),
				&wire.Transaction{Type: wire.TxTypeBalance, Value: "-10"}:        fmt.Errorf("index:0, field:value, error:value must be greater than zero"),
				&wire.Transaction{}:                                              fmt.Errorf("index:0, field:timestamp, error:timestamp is required"),
				&wire.Transaction{Timestamp: time.Now().Add(time.Minute).Unix()}: fmt.Errorf("index:0, field:timestamp, error:timestamp cannot be a future time"),
				&wire.Transaction{}:                                              fmt.Errorf("index:0, field:fee, error:fee is required"),
				&wire.Transaction{Fee: "1oo"}:                                    fmt.Errorf("index:0, field:fee, error:could not convert to decimal"),
				&wire.Transaction{Type: wire.TxTypeBalance, Fee: "0.0001"}:       fmt.Errorf("index:0, field:fee, error:fee cannot be below the minimum balance transaction fee {0.01000000}"),
				&wire.Transaction{}:                                              fmt.Errorf("index:0, field:hash, error:hash is required"),
				&wire.Transaction{Hash: "incorrect"}:                             fmt.Errorf("index:0, field:hash, error:hash is not correct"),
				&wire.Transaction{}:                                              fmt.Errorf("index:0, field:sig, error:signature is required"),
			}
			for tx, err := range cases {
				validator = NewTxsValidator([]*wire.Transaction{tx}, nil, bchain, false)
				errs := validator.checkFormatAndValue(tx)
				Expect(errs).To(ContainElement(err))
			}
		})

		It("should check if transaction exists in the txpool supplied", func() {
			sender := crypto.NewKeyFromIntSeed(1)
			receiver := crypto.NewKeyFromIntSeed(1)
			tx := &wire.Transaction{
				Type:         wire.TxTypeBalance,
				Nonce:        1,
				To:           sender.Addr(),
				From:         receiver.Addr(),
				Value:        "10",
				SenderPubKey: sender.PubKey().Base58(),
				Fee:          "0.1",
				Timestamp:    time.Now().Unix(),
				Hash:         "some_hash",
				Sig:          "invalid",
			}
			tx.Hash = tx.ComputeHash2()
			sig, err := wire.TxSign(tx, sender.PrivKey().Base58())
			Expect(err).To(BeNil())
			tx.Sig = util.ToHex(sig)

			txp := txpool.NewTxPool(1)
			err = txp.Put(tx)
			Expect(err).To(BeNil())

			validator = NewTxsValidator([]*wire.Transaction{tx}, txp, bchain, true)
			errs := validator.Validate()
			Expect(errs).To(HaveLen(1))
			Expect(errs).To(ContainElement(fmt.Errorf("index:0, error:transaction already exist in tx pool")))
		})

		Context("transaction in main chain", func() {
			It("should return error if transaction already", func() {
				var result []*database.KVObject
				store.Get(nil, &result)
				tx := genesisBlock.Transactions[0]
				validator = NewTxsValidator([]*wire.Transaction{tx}, nil, bchain, true)
				errs := validator.Validate()
				Expect(errs).To(HaveLen(1))
				Expect(errs).To(ContainElement(fmt.Errorf("index:0, error:transaction already exist in main chain")))
			})
		})
	})

	Describe(".checkSignature", func() {

		var sender, receiver *crypto.Key

		BeforeEach(func() {
			sender, receiver = crypto.NewKeyFromIntSeed(1), crypto.NewKeyFromIntSeed(2)
		})

		It("should return err if sender pub key is invalid", func() {
			tx := &wire.Transaction{SenderPubKey: "incorrect"}
			validator := NewTxsValidator(nil, nil, bchain, false)
			errs := validator.checkSignature(tx)
			Expect(errs).To(HaveLen(1))
			Expect(errs).To(ContainElement(fmt.Errorf("index:0, field:senderPubKey, error:checksum error")))
		})

		It("should return err if signature is not correct", func() {
			tx := &wire.Transaction{
				Type:         wire.TxTypeBalance,
				Nonce:        1,
				To:           receiver.Addr(),
				From:         sender.Addr(),
				SenderPubKey: sender.PubKey().Base58(),
				Value:        "10",
				Timestamp:    time.Now().Unix(),
				Fee:          "0.1",
				Hash:         "some_hash",
				Sig:          "invalid",
			}
			validator := NewTxsValidator(nil, nil, bchain, false)
			errs := validator.checkSignature(tx)
			Expect(errs).To(HaveLen(1))
			Expect(errs).To(ContainElement(fmt.Errorf("index:0, field:sig, error:signature is not valid")))
		})

		It("should return no error if signature is correct", func() {
			tx := &wire.Transaction{
				Type:         wire.TxTypeBalance,
				Nonce:        1,
				To:           receiver.Addr(),
				From:         sender.Addr(),
				SenderPubKey: sender.PubKey().Base58(),
				Value:        "10",
				Timestamp:    time.Now().Unix(),
				Fee:          "0.1",
			}
			tx.Hash = tx.ComputeHash2()
			sig, err := wire.TxSign(tx, sender.PrivKey().Base58())
			Expect(err).To(BeNil())
			tx.Sig = util.ToHex(sig)

			validator := NewTxsValidator(nil, nil, bchain, false)
			errs := validator.checkSignature(tx)
			Expect(errs).To(HaveLen(0))
		})
	})
})
