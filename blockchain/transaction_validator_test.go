package blockchain

import (
	"fmt"
	"time"

	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/txpool"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/wire"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var TransactionValidatorTest = func() bool {
	return Describe("TransactionValidator", func() {
		Describe(".checks", func() {

			var validator *TxsValidator

			It("should return error if tx = nil", func() {
				errs := validator.check(nil)
				Expect(errs).To(HaveLen(1))
				Expect(errs).To(ContainElement(fmt.Errorf("nil tx")))
			})

			It("should test all validation rules", func() {
				var cases = map[core.Transaction]interface{}{
					&wire.Transaction{Type: 0}:                                                                                                                 fmt.Errorf("index:0, field:type, error:unsupported transaction type"),
					&wire.Transaction{Type: wire.TxTypeBalance, Nonce: -1}:                                                                                     fmt.Errorf("index:0, field:nonce, error:nonce must be non-negative"),
					&wire.Transaction{Type: wire.TxTypeBalance, Nonce: 0}:                                                                                      fmt.Errorf("index:0, field:to, error:recipient address is required"),
					&wire.Transaction{To: "invalid", Type: wire.TxTypeBalance, Nonce: 0}:                                                                       fmt.Errorf("index:0, field:to, error:recipient address is not valid"),
					&wire.Transaction{From: "invalid", Type: wire.TxTypeBalance, Nonce: 0}:                                                                     fmt.Errorf("index:0, field:from, error:sender address is not valid"),
					&wire.Transaction{To: util.String(sender.Addr()), Type: wire.TxTypeBalance, Nonce: 0}:                                                      fmt.Errorf("index:0, field:senderPubKey, error:sender public key is required"),
					&wire.Transaction{SenderPubKey: "invalid", To: util.String(sender.Addr()), Type: wire.TxTypeBalance, Nonce: 0}:                             fmt.Errorf("index:0, field:senderPubKey, error:sender public key is not valid"),
					&wire.Transaction{SenderPubKey: util.String(sender.PubKey().Base58()), To: util.String(sender.Addr()), Type: wire.TxTypeBalance, Nonce: 0}: fmt.Errorf("index:0, field:value, error:value is required"),
					&wire.Transaction{Type: wire.TxTypeBalance, Value: "1oo"}:                                                                                  fmt.Errorf("index:0, field:value, error:could not convert to decimal"),
					&wire.Transaction{Type: wire.TxTypeBalance, Value: "-10"}:                                                                                  fmt.Errorf("index:0, field:value, error:value must be greater than zero"),
					&wire.Transaction{}:                                                                                                                        fmt.Errorf("index:0, field:timestamp, error:timestamp is required"),
					&wire.Transaction{Type: wire.TxTypeBalance}:                                                                                                fmt.Errorf("index:0, field:fee, error:fee is required"),
					&wire.Transaction{Type: wire.TxTypeBalance, Fee: "1oo"}:                                                                                    fmt.Errorf("index:0, field:fee, error:could not convert to decimal"),
					&wire.Transaction{Type: wire.TxTypeBalance, Fee: "0.0001"}:                                                                                 fmt.Errorf("index:0, field:fee, error:fee cannot be below the minimum balance transaction fee {0.0100000000000000}"),
					&wire.Transaction{}:                                                                                                                        fmt.Errorf("index:0, field:hash, error:hash is required"),
					&wire.Transaction{Hash: util.StrToHash("incorrect")}:                                                                                       fmt.Errorf("index:0, field:hash, error:hash is not correct"),
					&wire.Transaction{}:                                                                                                                        fmt.Errorf("index:0, field:sig, error:signature is required"),
					&wire.Transaction{Type: wire.TxTypeAlloc, From: util.String(sender.Addr()), To: util.String(receiver.Addr())}: fmt.Errorf("index:0, field:from, error:sender and recipient must be same address"),
				}
				for tx, err := range cases {
					validator = NewTxsValidator([]core.Transaction{tx}, nil, bc, false)
					errs := validator.check(tx)
					Expect(errs).To(ContainElement(err))
				}
			})

			It("should check if transaction exists in the txpool supplied", func() {
				sender := crypto.NewKeyFromIntSeed(1)
				receiver := crypto.NewKeyFromIntSeed(1)
				tx := &wire.Transaction{
					Type:         wire.TxTypeBalance,
					Nonce:        1,
					To:           util.String(sender.Addr()),
					From:         util.String(receiver.Addr()),
					Value:        "10",
					SenderPubKey: util.String(sender.PubKey().Base58()),
					Fee:          "0.1",
					Timestamp:    time.Now().Unix(),
					Hash:         util.StrToHash("some_hash"),
					Sig:          []byte("invalid"),
				}
				tx.Hash = tx.ComputeHash()
				sig, err := wire.TxSign(tx, sender.PrivKey().Base58())
				Expect(err).To(BeNil())
				tx.Sig = sig

				txp := txpool.NewTxPool(1)
				err = txp.Put(tx)
				Expect(err).To(BeNil())

				validator = NewTxsValidator([]core.Transaction{tx}, txp, bc, true)
				errs := validator.Validate()
				Expect(errs).To(HaveLen(1))
				Expect(errs).To(ContainElement(fmt.Errorf("index:0, error:transaction already exist in tx pool")))
			})

			Context("transaction in main chain", func() {

				var tx core.Transaction

				BeforeEach(func() {
					tx = genesisBlock.GetTransactions()[0]
					err = genesisChain.PutTransactions([]core.Transaction{tx})
					Expect(err).To(BeNil())
				})

				It("should return error if transaction already exists in the main chain", func() {
					validator = NewTxsValidator([]core.Transaction{tx}, nil, bc, true)
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
				validator := NewTxsValidator(nil, nil, bc, false)
				errs := validator.checkSignature(tx)
				Expect(errs).To(HaveLen(1))
				Expect(errs).To(ContainElement(fmt.Errorf("index:0, field:senderPubKey, error:checksum error")))
			})

			It("should return err if signature is not correct", func() {
				tx := &wire.Transaction{
					Type:         wire.TxTypeBalance,
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
				tx := &wire.Transaction{
					Type:         wire.TxTypeBalance,
					Nonce:        1,
					To:           util.String(receiver.Addr()),
					From:         util.String(sender.Addr()),
					SenderPubKey: util.String(sender.PubKey().Base58()),
					Value:        "10",
					Timestamp:    time.Now().Unix(),
					Fee:          "0.1",
				}
				tx.Hash = tx.ComputeHash()
				sig, err := wire.TxSign(tx, sender.PrivKey().Base58())
				Expect(err).To(BeNil())
				tx.Sig = sig

				validator := NewTxsValidator(nil, nil, bc, false)
				errs := validator.checkSignature(tx)
				Expect(errs).To(HaveLen(0))
			})
		})
	})
}
