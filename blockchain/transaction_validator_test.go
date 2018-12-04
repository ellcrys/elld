package blockchain

import (
	"fmt"
	"math/big"
	"os"
	"time"

	. "github.com/ellcrys/elld/blockchain/testutil"
	"github.com/ellcrys/elld/blockchain/txpool"
	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/elldb"
	"github.com/ellcrys/elld/testutil"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/types/core"

	"github.com/ellcrys/elld/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TransactionValidator", func() {

	var err error
	var bc *Blockchain
	var cfg *config.EngineConfig
	var db elldb.DB
	var genesisBlock types.Block
	var genesisChain *Chain
	var sender, receiver *crypto.Key

	BeforeEach(func() {
		cfg, err = testutil.SetTestCfg()
		Expect(err).To(BeNil())

		db = elldb.NewDB(cfg.DataDir())
		err = db.Open(util.RandString(5))
		Expect(err).To(BeNil())

		sender = crypto.NewKeyFromIntSeed(1)
		receiver = crypto.NewKeyFromIntSeed(2)

		bc = New(txpool.New(100), cfg, log)
		bc.SetDB(db)
	})

	BeforeEach(func() {
		genesisBlock, err = LoadBlockFromFile("genesis-test.json")
		Expect(err).To(BeNil())
		bc.SetGenesisBlock(genesisBlock)
		err = bc.Up()
		Expect(err).To(BeNil())
		genesisChain = bc.bestChain
	})

	AfterEach(func() {
		db.Close()
		err = os.RemoveAll(cfg.DataDir())
		Expect(err).To(BeNil())
	})

	Describe(".checks", func() {

		var validator *TxsValidator

		It("should return error if tx = nil", func() {
			errs := validator.CheckFields(nil)
			Expect(errs).To(HaveLen(1))
			Expect(errs).To(ContainElement(fmt.Errorf("nil tx")))
		})

		It("should test validation rules", func() {
			var cases = map[types.Transaction]interface{}{
				&core.Transaction{
					Type: 0,
				}: fmt.Errorf("index:0, field:type, error:unsupported transaction type"),

				&core.Transaction{
					Type:  core.TxTypeBalance,
					Nonce: 0,
				}: fmt.Errorf("index:0, field:to, error:recipient address is required"),

				&core.Transaction{
					To:    "invalid",
					Type:  core.TxTypeBalance,
					Nonce: 0,
				}: fmt.Errorf("index:0, field:to, error:recipient address is not valid"),

				&core.Transaction{
					From:  "invalid",
					Type:  core.TxTypeBalance,
					Nonce: 0,
				}: fmt.Errorf("index:0, field:from, error:sender address is not valid"),

				&core.Transaction{
					To:    util.String(sender.Addr()),
					Type:  core.TxTypeBalance,
					Nonce: 0,
				}: fmt.Errorf("index:0, field:senderPubKey, error:sender public key is required"),

				&core.Transaction{
					SenderPubKey: "invalid",
					To:           util.String(sender.Addr()),
					Type:         core.TxTypeBalance,
					Nonce:        0,
				}: fmt.Errorf("index:0, field:senderPubKey, error:sender public key is not valid"),

				&core.Transaction{
					SenderPubKey: util.String(sender.PubKey().Base58()),
					To:           util.String(sender.Addr()),
					Type:         core.TxTypeBalance,
					Nonce:        0,
				}: fmt.Errorf("index:0, field:value, error:value is required"),

				&core.Transaction{
					Type:  core.TxTypeBalance,
					Value: "1oo",
				}: fmt.Errorf("index:0, field:value, error:could not convert to decimal"),

				&core.Transaction{
					Type:  core.TxTypeBalance,
					Value: "-10",
				}: fmt.Errorf("index:0, field:value, error:negative value not allowed"),

				&core.Transaction{}: fmt.Errorf("index:0, field:timestamp, error:timestamp is required"),

				&core.Transaction{
					Type: core.TxTypeBalance,
				}: fmt.Errorf("index:0, field:fee, error:fee is required"),

				&core.Transaction{
					Type: core.TxTypeBalance,
					Fee:  "1oo",
				}: fmt.Errorf("index:0, field:fee, error:could not convert to decimal"),

				&core.Transaction{
					Type: core.TxTypeBalance,
					Fee:  "0.000001",
				}: fmt.Errorf("index:0, field:fee, error:fee is too low. Minimum fee expected: 0.440000000000000009159339953157541458494961261749267578125 (for 44 bytes)"),

				&core.Transaction{}: fmt.Errorf("index:0, field:hash, error:hash is required"),

				&core.Transaction{
					Hash: util.StrToHash("incorrect"),
				}: fmt.Errorf("index:0, field:hash, error:hash is not correct"),

				&core.Transaction{}: fmt.Errorf("index:0, field:sig, error:signature is required"),

				&core.Transaction{
					SenderPubKey: util.String(sender.PubKey().Base58()),
					From:         receiver.Addr(),
				}: fmt.Errorf("index:0, field:from, error:sender address is not derived from the sender public key"),
			}
			for tx, err := range cases {
				validator = NewTxsValidator([]types.Transaction{tx}, nil, bc)
				errs := validator.CheckFields(tx)
				Expect(errs).To(ContainElement(err))
			}
		})

		Describe(".Validate", func() {

			Context("when duplicate transactions exist", func() {

				var txs []types.Transaction

				BeforeEach(func() {
					txs = []types.Transaction{
						core.NewTx(core.TxTypeBalance, 1, util.String(sender.Addr()), sender, "1", "2.5", 1532730723),
						core.NewTx(core.TxTypeBalance, 1, util.String(sender.Addr()), sender, "1", "2.5", 1532730723),
					}
				})

				It("should return duplicate transaction error", func() {
					txp := txpool.New(1)
					validator = NewTxsValidator(txs, txp, bc)
					errs := validator.Validate()
					Expect(errs).To(ContainElement(fmt.Errorf("index:1, error:duplicate transaction")))
				})
			})

			Context("context checks", func() {

				var tx *core.Transaction

				BeforeEach(func() {
					tx = core.NewTx(core.TxTypeBalance, 1,
						util.String(sender.Addr()), sender, "1", "2.5", time.Now().Unix())
					bc.txPool.Put(tx)
				})

				Context("when ContextBlock is set", func() {
					It("should not check transaction existence in the transaction pool", func() {
						validator = NewTxsValidator([]types.Transaction{tx}, bc.txPool, bc)
						validator.addContext(types.ContextBlock)
						errs := validator.Validate()
						Expect(errs).To(HaveLen(0))
					})
				})

				Context("when ContextBranch is set", func() {

					var block2 types.Block

					BeforeEach(func() {
						block2 = MakeBlockWithSingleTx(bc, genesisChain, sender, sender, 1)
						_, err := bc.ProcessBlock(block2)
						Expect(err).To(BeNil())
					})

					It("should not check transaction existence in the main chain", func() {
						tx.Nonce = 2
						tx.Hash = tx.ComputeHash()
						sig, err := core.TxSign(tx, sender.PrivKey().Base58())
						Expect(err).To(BeNil())
						tx.Sig = sig

						txp := txpool.New(1)
						validator = NewTxsValidator([]types.Transaction{tx}, txp, bc)
						validator.addContext(types.ContextBranch)
						errs := validator.Validate()
						Expect(errs).To(HaveLen(0))
					})
				})
			})

			Context("when a transaction is already on the main chain", func() {

				var tx types.Transaction

				BeforeEach(func() {
					block := MakeBlockWithBalanceTx(bc, genesisChain, sender, sender)
					_, err := bc.ProcessBlock(block)
					Expect(err).To(BeNil())
					tx = block.GetTransactions()[0]
				})

				It("should return error if transaction already exists in the main chain", func() {
					txp := txpool.New(1)
					validator = NewTxsValidator([]types.Transaction{tx}, txp, bc)
					errs := validator.Validate()
					Expect(errs).To(HaveLen(1))
					Expect(errs).To(ContainElement(fmt.Errorf("index:0, error:transaction already exist in main chain")))
				})
			})
		})

	})

	Describe(".checkSignature", func() {

		var sender, receiver *crypto.Key

		BeforeEach(func() {
			sender, receiver = crypto.NewKeyFromIntSeed(1), crypto.NewKeyFromIntSeed(2)
		})

		It("should return err if sender pub key is invalid", func() {
			tx := &core.Transaction{SenderPubKey: "incorrect"}
			validator := NewTxsValidator(nil, nil, bc)
			errs := validator.checkSignature(tx)
			Expect(errs).To(HaveLen(1))
			Expect(errs).To(ContainElement(fmt.Errorf("index:0, field:senderPubKey, error:checksum error")))
		})

		It("should return err if signature is not correct", func() {
			tx := &core.Transaction{
				Type:         core.TxTypeBalance,
				Nonce:        1,
				To:           util.String(receiver.Addr()),
				From:         util.String(sender.Addr()),
				SenderPubKey: util.String(sender.PubKey().Base58()),
				Value:        "10",
				Timestamp:    time.Now().Unix(),
				Fee:          "0.1",
				Hash:         util.StrToHash("some_hash"),
				Sig:          []byte("invalid"),
			}
			validator := NewTxsValidator(nil, nil, bc)
			errs := validator.checkSignature(tx)
			Expect(errs).To(HaveLen(1))
			Expect(errs).To(ContainElement(fmt.Errorf("index:0, field:sig, error:signature is not valid")))
		})

		It("should return no error if signature is correct", func() {
			tx := &core.Transaction{
				Type:         core.TxTypeBalance,
				Nonce:        1,
				To:           util.String(receiver.Addr()),
				From:         util.String(sender.Addr()),
				SenderPubKey: util.String(sender.PubKey().Base58()),
				Value:        "10",
				Timestamp:    time.Now().Unix(),
				Fee:          "0.1",
			}
			tx.Hash = tx.ComputeHash()
			sig, err := core.TxSign(tx, sender.PrivKey().Base58())
			Expect(err).To(BeNil())
			tx.Sig = sig

			validator := NewTxsValidator(nil, nil, bc)
			errs := validator.checkSignature(tx)
			Expect(errs).To(HaveLen(0))
		})
	})

	Describe(".consistencyCheck", func() {

		var tx types.Transaction
		var sender, receiver *crypto.Key
		var txp *txpool.TxPool

		BeforeEach(func() {
			txp = txpool.New(1)
			sender, receiver = crypto.NewKeyFromIntSeed(1), crypto.NewKeyFromIntSeed(2)
			tx = &core.Transaction{
				Type:         core.TxTypeBalance,
				Nonce:        1,
				To:           util.String(receiver.Addr()),
				From:         util.String(sender.Addr()),
				SenderPubKey: util.String(sender.PubKey().Base58()),
				Value:        "10",
				Timestamp:    time.Now().Unix(),
				Fee:          "0.1",
			}
			tx.SetHash(tx.ComputeHash())
		})

		It("should return error when exact transaction exist in the pool", func() {
			txp := txpool.New(1)
			Expect(txp.Put(tx)).To(BeNil())
			validator := NewTxValidator(nil, txp, bc)
			errs := validator.consistencyCheck(tx)
			Expect(errs).ToNot(BeEmpty())
			Expect(errs).To(ContainElement(fmt.Errorf("index:0, error:transaction already exist in the transactions pool")))
		})

		Context("add a block with test transactions", func() {

			var block2 types.Block

			BeforeEach(func() {
				block2 = MakeTestBlock(bc, genesisChain, &types.GenerateBlockParams{
					Transactions: []types.Transaction{
						core.NewTx(core.TxTypeBalance, 1, util.String(receiver.Addr()), sender, "1", "2.5", 1532730723),
						core.NewTx(core.TxTypeAlloc, 1, util.String(sender.Addr()), sender, "2.5", "0", 1532730725),
					},
					Creator:    sender,
					Nonce:      util.EncodeNonce(1),
					Difficulty: new(big.Int).SetInt64(131072),
				})
				_, err = bc.ProcessBlock(block2)
				Expect(err).To(BeNil())
			})

			It("should return err='index:0, error:transaction already exist in main chain' when exact transaction exist in the main chain", func() {
				validator := NewTxValidator(nil, txp, bc)
				errs := validator.consistencyCheck(block2.GetTransactions()[0])
				Expect(errs).ToNot(BeEmpty())
				Expect(errs).To(ContainElement(fmt.Errorf("index:0, error:transaction already exist in main chain")))
			})
		})

		When("transaction originator's account is not found", func() {
			It("should return err='index:0, field:from, error:sender account not found'", func() {
				tx.SetFrom("unknown_address")
				validator := NewTxValidator(nil, txp, bc)
				errs := validator.consistencyCheck(tx)
				Expect(errs).ToNot(BeEmpty())
				Expect(errs).To(ContainElement(fmt.Errorf("index:0, field:from, error:sender account not found")))
			})
		})

		When("sender has insufficient balance", func() {
			It("should return err='index:0, error:insufficient account balance'", func() {
				tx.SetValue("10000000")
				validator := NewTxValidator(nil, txp, bc)
				errs := validator.consistencyCheck(tx)
				Expect(errs).ToNot(BeEmpty())
				Expect(errs).To(ContainElement(fmt.Errorf("index:0, error:insufficient account balance")))
			})
		})

		When("In ContextTxPool validation context and a transaction's nonce is lower than expected", func() {

			var tx2 types.Transaction

			BeforeEach(func() {
				tx2 = &core.Transaction{
					Type:         core.TxTypeBalance,
					Nonce:        0,
					To:           util.String(receiver.Addr()),
					From:         util.String(sender.Addr()),
					SenderPubKey: util.String(sender.PubKey().Base58()),
					Value:        "1",
					Timestamp:    1234567,
					Fee:          "0.1",
				}
				tx2.SetHash(tx2.ComputeHash())
			})

			It("should return err='index:0, error:invalid nonce: has 0, wants from 1", func() {
				txp := txpool.New(1)

				validator := NewTxValidator(nil, txp, bc)
				errs := validator.consistencyCheck(tx2)

				Expect(errs).ToNot(BeEmpty())
				Expect(errs).To(ContainElement(fmt.Errorf("index:0, error:invalid nonce: has 0, wants from 1")))
			})
		})

		When("In ContextTxPool validation context and a transaction's nonce 2 more than current nonce", func() {

			var tx2 types.Transaction

			BeforeEach(func() {
				tx2 = &core.Transaction{
					Type:         core.TxTypeBalance,
					Nonce:        2,
					To:           util.String(receiver.Addr()),
					From:         util.String(sender.Addr()),
					SenderPubKey: util.String(sender.PubKey().Base58()),
					Value:        "1",
					Timestamp:    1234567,
					Fee:          "0.1",
				}
				tx2.SetHash(tx2.ComputeHash())
			})

			It("should return nil", func() {
				txp := txpool.New(1)
				validator := NewTxValidator(nil, txp, bc)
				errs := validator.consistencyCheck(tx2)
				Expect(errs).To(BeNil())
			})
		})

		When("In types.ContextBlock validation context and a transaction's nonce is greater than expected", func() {

			var tx2 types.Transaction

			BeforeEach(func() {
				tx2 = &core.Transaction{
					Type:         core.TxTypeBalance,
					Nonce:        2,
					To:           util.String(receiver.Addr()),
					From:         util.String(sender.Addr()),
					SenderPubKey: util.String(sender.PubKey().Base58()),
					Value:        "1",
					Timestamp:    1234567,
					Fee:          "0.1",
				}
				tx2.SetHash(tx2.ComputeHash())
			})

			It("should return err='index:0, error:invalid nonce: has 2, wants 1'", func() {
				txp := txpool.New(1)
				validator := NewTxValidator(nil, txp, bc)
				validator.addContext(types.ContextBlock)
				errs := validator.consistencyCheck(tx2)
				Expect(errs).ToNot(BeEmpty())
				Expect(errs).To(ContainElement(fmt.Errorf("index:0, error:invalid nonce: has 2, wants 1")))
			})
		})
	})

})
