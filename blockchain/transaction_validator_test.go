package blockchain

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/txpool"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/types/core/objects"
	"github.com/ellcrys/elld/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var TransactionValidatorTest = func() bool {
	return Describe("TransactionValidator", func() {
		Describe(".checks", func() {

			var validator *TxsValidator

			It("should return error if tx = nil", func() {
				errs := validator.fieldsCheck(nil)
				Expect(errs).To(HaveLen(1))
				Expect(errs).To(ContainElement(fmt.Errorf("nil tx")))
			})

			It("should test all validation rules", func() {
				var cases = map[core.Transaction]interface{}{
					&objects.Transaction{Type: 0}:                                                                                                                    fmt.Errorf("index:0, field:type, error:unsupported transaction type"),
					&objects.Transaction{Type: objects.TxTypeBalance, Nonce: 0}:                                                                                      fmt.Errorf("index:0, field:to, error:recipient address is required"),
					&objects.Transaction{To: "invalid", Type: objects.TxTypeBalance, Nonce: 0}:                                                                       fmt.Errorf("index:0, field:to, error:recipient address is not valid"),
					&objects.Transaction{From: "invalid", Type: objects.TxTypeBalance, Nonce: 0}:                                                                     fmt.Errorf("index:0, field:from, error:sender address is not valid"),
					&objects.Transaction{To: util.String(sender.Addr()), Type: objects.TxTypeBalance, Nonce: 0}:                                                      fmt.Errorf("index:0, field:senderPubKey, error:sender public key is required"),
					&objects.Transaction{SenderPubKey: "invalid", To: util.String(sender.Addr()), Type: objects.TxTypeBalance, Nonce: 0}:                             fmt.Errorf("index:0, field:senderPubKey, error:sender public key is not valid"),
					&objects.Transaction{SenderPubKey: util.String(sender.PubKey().Base58()), To: util.String(sender.Addr()), Type: objects.TxTypeBalance, Nonce: 0}: fmt.Errorf("index:0, field:value, error:value is required"),
					&objects.Transaction{Type: objects.TxTypeBalance, Value: "1oo"}:                                                                                  fmt.Errorf("index:0, field:value, error:could not convert to decimal"),
					&objects.Transaction{Type: objects.TxTypeBalance, Value: "-10"}:                                                                                  fmt.Errorf("index:0, field:value, error:value must be greater than zero"),
					&objects.Transaction{}:                                                                                                                           fmt.Errorf("index:0, field:timestamp, error:timestamp is required"),
					&objects.Transaction{Type: objects.TxTypeBalance}:                                                                                                fmt.Errorf("index:0, field:fee, error:fee is required"),
					&objects.Transaction{Type: objects.TxTypeBalance, Fee: "1oo"}:                                                                                    fmt.Errorf("index:0, field:fee, error:could not convert to decimal"),
					&objects.Transaction{Type: objects.TxTypeBalance, Fee: "0.000001"}:                                                                               fmt.Errorf("index:0, field:fee, error:fee is too low. Minimum fee expected: 0.44 (for 44 bytes)"),
					&objects.Transaction{}:                                                                                                                           fmt.Errorf("index:0, field:hash, error:hash is required"),
					&objects.Transaction{Hash: util.StrToHash("incorrect")}:                                                                                          fmt.Errorf("index:0, field:hash, error:hash is not correct"),
					&objects.Transaction{}:                                                                                                                           fmt.Errorf("index:0, field:sig, error:signature is required"),
				}
				for tx, err := range cases {
					validator = NewTxsValidator([]core.Transaction{tx}, nil, bc, false)
					errs := validator.fieldsCheck(tx)
					Expect(errs).To(ContainElement(err))
				}
			})

			It("should check if transaction exists in the txpool supplied", func() {
				sender := crypto.NewKeyFromIntSeed(1)
				receiver := crypto.NewKeyFromIntSeed(1)
				tx := &objects.Transaction{
					Type:         objects.TxTypeBalance,
					Nonce:        1,
					To:           util.String(sender.Addr()),
					From:         util.String(receiver.Addr()),
					Value:        "10",
					SenderPubKey: util.String(sender.PubKey().Base58()),
					Fee:          "2.37",
					Timestamp:    time.Now().Unix(),
					Hash:         util.StrToHash("some_hash"),
					Sig:          []byte("invalid"),
				}
				tx.Hash = tx.ComputeHash()
				sig, err := objects.TxSign(tx, sender.PrivKey().Base58())
				Expect(err).To(BeNil())
				tx.Sig = sig

				txp := txpool.New(1)
				err = txp.Put(tx)
				Expect(err).To(BeNil())

				validator = NewTxsValidator([]core.Transaction{tx}, txp, bc, true)
				errs := validator.Validate()
				Expect(errs).To(HaveLen(1))
				Expect(errs).To(ContainElement(fmt.Errorf("index:0, error:transaction already exist in the transactions pool")))
			})

			Context("transaction in main chain", func() {

				var tx core.Transaction

				BeforeEach(func() {
					tx = genesisBlock.GetTransactions()[0]
					err = genesisChain.PutTransactions([]core.Transaction{tx}, genesisBlock.GetNumber())
					Expect(err).To(BeNil())
				})

				It("should return error if transaction already exists in the main chain", func() {
					txp := txpool.New(1)
					validator = NewTxsValidator([]core.Transaction{tx}, txp, bc, true)
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
				tx := &objects.Transaction{SenderPubKey: "incorrect"}
				validator := NewTxsValidator(nil, nil, bc, false)
				errs := validator.checkSignature(tx)
				Expect(errs).To(HaveLen(1))
				Expect(errs).To(ContainElement(fmt.Errorf("index:0, field:senderPubKey, error:checksum error")))
			})

			It("should return err if signature is not correct", func() {
				tx := &objects.Transaction{
					Type:         objects.TxTypeBalance,
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
				validator := NewTxsValidator(nil, nil, bc, false)
				errs := validator.checkSignature(tx)
				Expect(errs).To(HaveLen(1))
				Expect(errs).To(ContainElement(fmt.Errorf("index:0, field:sig, error:signature is not valid")))
			})

			It("should return no error if signature is correct", func() {
				tx := &objects.Transaction{
					Type:         objects.TxTypeBalance,
					Nonce:        1,
					To:           util.String(receiver.Addr()),
					From:         util.String(sender.Addr()),
					SenderPubKey: util.String(sender.PubKey().Base58()),
					Value:        "10",
					Timestamp:    time.Now().Unix(),
					Fee:          "0.1",
				}
				tx.Hash = tx.ComputeHash()
				sig, err := objects.TxSign(tx, sender.PrivKey().Base58())
				Expect(err).To(BeNil())
				tx.Sig = sig

				validator := NewTxsValidator(nil, nil, bc, false)
				errs := validator.checkSignature(tx)
				Expect(errs).To(HaveLen(0))
			})
		})

		Describe(".consistencyCheck", func() {

			var tx core.Transaction
			var sender, receiver *crypto.Key
			var txp *txpool.TxPool

			BeforeEach(func() {
				txp = txpool.New(1)
				sender, receiver = crypto.NewKeyFromIntSeed(1), crypto.NewKeyFromIntSeed(2)
				tx = &objects.Transaction{
					Type:         objects.TxTypeBalance,
					Nonce:        1,
					To:           util.String(receiver.Addr()),
					From:         util.String(sender.Addr()),
					SenderPubKey: util.String(sender.PubKey().Base58()),
					Value:        "10",
					Timestamp:    1234567,
					Fee:          "0.1",
				}
				tx.SetHash(tx.ComputeHash())
			})

			It("should return err='index:0, error:transaction already exist in the transactions pool' when exact transaction exist in pool", func() {
				txp := txpool.New(1)
				Expect(txp.Put(tx)).To(BeNil())
				validator := NewTxValidator(nil, txp, bc)
				errs := validator.consistencyCheck(tx)
				Expect(errs).ToNot(BeEmpty())
				Expect(errs).To(ContainElement(fmt.Errorf("index:0, error:transaction already exist in the transactions pool")))
			})

			Context("add a block with test transactions", func() {

				var block2 core.Block

				BeforeEach(func() {
					block2 = MakeTestBlock(bc, genesisChain, &core.GenerateBlockParams{
						Transactions: []core.Transaction{
							objects.NewTx(objects.TxTypeBalance, 1, util.String(receiver.Addr()), sender, "1", "2.36", 1532730723),
							objects.NewTx(objects.TxTypeAlloc, 1, util.String(sender.Addr()), sender, "2.36", "0", 1532730725),
						},
						Creator:    sender,
						Nonce:      core.EncodeNonce(1),
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

				var tx2 core.Transaction

				BeforeEach(func() {
					tx2 = &objects.Transaction{
						Type:         objects.TxTypeBalance,
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

				var tx2 core.Transaction

				BeforeEach(func() {
					tx2 = &objects.Transaction{
						Type:         objects.TxTypeBalance,
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

			When("In ContextBlock validation context and a transaction's nonce is greater than expected", func() {

				var tx2 core.Transaction

				BeforeEach(func() {
					tx2 = &objects.Transaction{
						Type:         objects.TxTypeBalance,
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
					validator.SetContext(ContextBlock)
					errs := validator.consistencyCheck(tx2)
					Expect(errs).ToNot(BeEmpty())
					Expect(errs).To(ContainElement(fmt.Errorf("index:0, error:invalid nonce: has 2, wants 1")))
				})
			})
		})
	})
}
