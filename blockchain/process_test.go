package blockchain

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ellcrys/elld/blockchain/common"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/types/core/objects"
	"github.com/ellcrys/elld/util"
	"github.com/jinzhu/copier"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var ProcessTest = func() bool {
	return Describe("Process", func() {

		Describe(".addOp", func() {

			var curOps = []common.Transition{
				&common.OpNewAccountBalance{
					OpBase: &common.OpBase{Addr: "addr1"},
					Account: &objects.Account{
						Balance: "10",
					},
				},
			}

			It("should add an additional op successfully", func() {
				op := &common.OpNewAccountBalance{
					OpBase: &common.OpBase{Addr: "addr2"},
					Account: &objects.Account{
						Balance: "10",
					},
				}
				newOps := addOp(curOps, op)
				Expect(newOps).To(HaveLen(2))
			})

			It("should replace any equal op found", func() {
				op := &common.OpNewAccountBalance{
					OpBase: &common.OpBase{Addr: "addr1"},
					Account: &objects.Account{
						Balance: "30",
					},
				}
				newOps := addOp(curOps, op)
				Expect(newOps).To(HaveLen(1))
				Expect(newOps[0]).To(Equal(op))
				Expect(newOps[0]).ToNot(Equal(curOps[0]))
			})
		})

		Describe(".processTransactions (only TxTypeBalance transactions)", func() {

			var account *objects.Account

			BeforeEach(func() {
				account = &objects.Account{Type: objects.AccountTypeBalance, Address: util.String(sender.Addr()), Balance: "10"}
				err = bc.CreateAccount(1, genesisChain, account)
				Expect(err).To(BeNil())
			})

			It("should return error if sender does not exist in the best chain", func() {
				var txs = []core.Transaction{
					&objects.Transaction{Type: 1, Nonce: 1, To: "e6i7rxApBYUt7w94gGDKTz45A5J567JfkS", From: "unknown", SenderPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
						Value: "1", Timestamp: 1532730724,
						Fee: "0.1", Sig: []uint8{},
						Hash: util.Hash{},
					},
				}

				_, err := bc.ProcessTransactions(txs, genesisChain)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("index{0}: failed to get sender's account: account not found"))
			})

			It("should return error if sender account has insufficient value", func() {
				var txs = []core.Transaction{
					&objects.Transaction{Type: 1, Nonce: 1, To: "e6i7rxApBYUt7w94gGDKTz45A5J567JfkS", From: "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad", SenderPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
						Value: "10000000", Timestamp: 1532730724,
						Fee: "0.1", Sig: []uint8{},
						Hash: util.Hash{},
					},
				}

				_, err := bc.ProcessTransactions(txs, genesisChain)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("index{0}: insufficient sender account balance"))
			})

			It("should panic if sender value is could not be converted to decimal", func() {
				var txs = []core.Transaction{
					&objects.Transaction{Type: 1, Nonce: 1, To: "e6i7rxApBYUt7w94gGDKTz45A5J567JfkS", From: "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad", SenderPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
						Value: "100_333", Timestamp: 1532730724,
						Fee: "0.1", Sig: []uint8{},
						Hash: util.Hash{},
					},
				}
				Expect(func() {
					bc.ProcessTransactions(txs, genesisChain)
				}).To(Panic())
			})

			Context("recipient does not have an account", func() {

				It(`should return operations and no error; expects 3 ops; 1 OpCreateAccount (for recipient) and 2 OpNewAccountBalance (sender and recipient);
					1st OpNewAccountBalance.Amount = 9; 2nd OpNewAccountBalance.Amount = 1`, func() {
					var txs = []core.Transaction{
						&objects.Transaction{Type: 1, Nonce: 1, To: "e6i7rxApBYUt7w94gGDKTz45A5J567JfkS", From: "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad", SenderPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
							Value: "1", Timestamp: 1532730724,
							Fee: "0.1", Sig: []uint8{},
							Hash: util.Hash{},
						},
					}

					ops, err := bc.ProcessTransactions(txs, genesisChain)
					Expect(err).To(BeNil())
					Expect(ops).To(HaveLen(3))

					Expect(ops[0]).To(BeAssignableToTypeOf(&common.OpCreateAccount{}))
					Expect(ops[0].Address()).To(Equal(txs[0].GetTo()))

					Expect(ops[1]).To(BeAssignableToTypeOf(&common.OpNewAccountBalance{}))
					Expect(ops[1].Address()).To(Equal(txs[0].GetFrom()))
					Expect(ops[1].(*common.OpNewAccountBalance).Account.GetBalance()).To(Equal(util.String("8.900000000000000000")))

					Expect(ops[2]).To(BeAssignableToTypeOf(&common.OpNewAccountBalance{}))
					Expect(ops[2].Address()).To(Equal(txs[0].GetTo()))
					Expect(ops[2].(*common.OpNewAccountBalance).Account.GetBalance()).To(Equal(util.String("1.000000000000000000")))
				})
			})

			Context("recipient has an account", func() {
				receiver := crypto.NewKeyFromIntSeed(3)

				BeforeEach(func() {
					account = &objects.Account{Type: objects.AccountTypeBalance, Address: util.String(receiver.Addr()), Balance: "0"}
					err = bc.CreateAccount(1, genesisChain, account)
					Expect(err).To(BeNil())
				})

				It(`should return operations and no error; expects 2 ops; 2 OpNewAccountBalance (sender and recipient);
					1st OpNewAccountBalance.Amount = 9; 2nd OpNewAccountBalance.Amount = 1`, func() {

					var txs = []core.Transaction{
						&objects.Transaction{Type: 1, Nonce: 1, To: util.String(receiver.Addr()), From: "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad", SenderPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
							Value: "1", Timestamp: 1532730724,
							Fee: "0.1", Sig: []uint8{},
							Hash: util.Hash{},
						},
					}

					ops, err := bc.ProcessTransactions(txs, genesisChain)
					Expect(err).To(BeNil())
					Expect(ops).To(HaveLen(2))

					Expect(ops[0]).To(BeAssignableToTypeOf(&common.OpNewAccountBalance{}))
					Expect(ops[0].Address()).To(Equal(txs[0].GetFrom()))
					Expect(ops[0].(*common.OpNewAccountBalance).Account.GetBalance()).To(Equal(util.String("8.900000000000000000")))

					Expect(ops[1]).To(BeAssignableToTypeOf(&common.OpNewAccountBalance{}))
					Expect(ops[1].Address()).To(Equal(txs[0].GetTo()))
					Expect(ops[1].(*common.OpNewAccountBalance).Account.GetBalance()).To(Equal(util.String("1.000000000000000000")))
				})
			})

			Context("sender has two transactions sending to a recipient", func() {

				Context("recipient has an account", func() {
					receiver := crypto.NewKeyFromIntSeed(3)

					BeforeEach(func() {
						account = &objects.Account{Type: objects.AccountTypeBalance, Address: util.String(receiver.Addr()), Balance: "0"}
						err = bc.CreateAccount(1, genesisChain, account)
						Expect(err).To(BeNil())
					})

					It(`should return operations and no error; expects 2 ops; 2 OpNewAccountBalance (sender and recipient);
													1st OpNewAccountBalance.Amount = 8; 2nd OpNewAccountBalance.Amount = 2`, func() {
						var txs = []core.Transaction{
							&objects.Transaction{Type: 1, Nonce: 1, To: util.String(receiver.Addr()), From: "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad", SenderPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
								Value: "1", Timestamp: 1532730724,
								Fee: "0.1", Sig: []uint8{},
								Hash: util.Hash{},
							},
							&objects.Transaction{Type: 1, Nonce: 1, To: util.String(receiver.Addr()), From: "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad", SenderPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
								Value: "1", Timestamp: 1532730724,
								Fee: "0.1", Sig: []uint8{},
								Hash: util.Hash{},
							},
						}

						ops, err := bc.ProcessTransactions(txs, genesisChain)
						Expect(err).To(BeNil())
						Expect(ops).To(HaveLen(2))

						Expect(ops[0]).To(BeAssignableToTypeOf(&common.OpNewAccountBalance{}))
						Expect(ops[0].Address()).To(Equal(txs[0].GetFrom()))
						Expect(ops[0].(*common.OpNewAccountBalance).Account.GetBalance()).To(Equal(util.String("7.800000000000000000")))

						Expect(ops[1]).To(BeAssignableToTypeOf(&common.OpNewAccountBalance{}))
						Expect(ops[1].Address()).To(Equal(txs[0].GetTo()))
						Expect(ops[1].(*common.OpNewAccountBalance).Account.GetBalance()).To(Equal(util.String("2.000000000000000000")))
					})
				})

				Context("recipient does not have an account", func() {
					It(`should return operations and no error; expects 3 ops; 1 OpCreateAccount (for recipient), 2 OpNewAccountBalance (sender and recipient);
														1st OpNewAccountBalance.Amount = 8; 2nd OpNewAccountBalance.Amount = 2`, func() {
						var txs = []core.Transaction{
							&objects.Transaction{Type: 1, Nonce: 1, To: util.String(receiver.Addr()), From: "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad", SenderPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
								Value: "1", Timestamp: 1532730724,
								Fee: "0.1", Sig: []uint8{},
								Hash: util.Hash{},
							},
							&objects.Transaction{Type: 1, Nonce: 1, To: util.String(receiver.Addr()), From: "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad", SenderPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
								Value: "1", Timestamp: 1532730724,
								Fee: "0.1", Sig: []uint8{},
								Hash: util.Hash{},
							},
						}

						ops, err := bc.ProcessTransactions(txs, genesisChain)
						Expect(err).To(BeNil())
						Expect(ops).To(HaveLen(3))

						Expect(ops[0]).To(BeAssignableToTypeOf(&common.OpCreateAccount{}))
						Expect(ops[0].Address()).To(Equal(txs[0].GetTo()))

						Expect(ops[1]).To(BeAssignableToTypeOf(&common.OpNewAccountBalance{}))
						Expect(ops[1].Address()).To(Equal(txs[0].GetFrom()))
						Expect(ops[1].(*common.OpNewAccountBalance).Account.GetBalance()).To(Equal(util.String("7.800000000000000000")))

						Expect(ops[2]).To(BeAssignableToTypeOf(&common.OpNewAccountBalance{}))
						Expect(ops[2].Address()).To(Equal(txs[0].GetTo()))
						Expect(ops[2].(*common.OpNewAccountBalance).Account.GetBalance()).To(Equal(util.String("2.000000000000000000")))
					})
				})
			})
		})

		Describe(".processTransactions (only TxTypeAllocCoin transactions)", func() {
			Context("only a single transaction", func() {
				When("recipient account does not exist", func() {
					It("should successfully return one state object = OpNewAccountBalance", func() {
						var txs = []core.Transaction{
							&objects.Transaction{Type: objects.TxTypeAlloc, Nonce: 1, To: "e6i7rxApBYUt7w94gGDKTz45A5J567JfkS", From: "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad", SenderPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
								Value: "10", Timestamp: 1532730724,
								Fee: "0.1", Sig: []uint8{},
								Hash: util.Hash{},
							},
						}

						ops, err := bc.ProcessTransactions(txs, genesisChain)
						Expect(err).To(BeNil())
						Expect(ops).To(HaveLen(1))
						Expect(ops[0]).To(BeAssignableToTypeOf(&common.OpNewAccountBalance{}))
						Expect(ops[0].(*common.OpNewAccountBalance).Account.GetBalance()).To(Equal(util.String("10.000000000000000000")))
					})
				})

				When("recipient account already exists with account balance = 100", func() {
					BeforeEach(func() {
						Expect(bc.CreateAccount(1, genesisChain, &objects.Account{
							Type:    objects.AccountTypeBalance,
							Address: "e6i7rxApBYUt7w94gGDKTz45A5J567JfkS",
							Balance: "100",
						})).To(BeNil())
					})

					It("should successfully return one state object = OpNewAccountBalance and Balance = 110.000000000000000000", func() {
						var txs = []core.Transaction{
							&objects.Transaction{Type: objects.TxTypeAlloc, Nonce: 1, To: "e6i7rxApBYUt7w94gGDKTz45A5J567JfkS", From: "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad", SenderPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
								Value: "10", Timestamp: 1532730724,
								Fee: "0.1", Sig: []uint8{},
								Hash: util.Hash{},
							},
						}
						ops, err := bc.ProcessTransactions(txs, genesisChain)
						Expect(err).To(BeNil())
						Expect(ops).To(HaveLen(1))
						Expect(ops[0]).To(BeAssignableToTypeOf(&common.OpNewAccountBalance{}))
						Expect(ops[0].(*common.OpNewAccountBalance).Account.GetBalance()).To(Equal(util.String("110.000000000000000000")))
					})
				})
			})
		})

		Describe(".ComputeTxsRoot", func() {
			It("should return expected root", func() {
				txs := []core.Transaction{
					&objects.Transaction{Sig: []byte("b"), Hash: util.StrToHash("hash_1")},
					&objects.Transaction{Sig: []byte("a"), Hash: util.StrToHash("hash_2")},
				}
				root := common.ComputeTxsRoot(txs)
				expected := util.Hash{
					0x37, 0xef, 0x76, 0x42, 0x69, 0x87, 0x67, 0xba, 0x8b, 0xfe, 0xf7, 0x5d, 0x66, 0x91, 0xda, 0x12,
					0x20, 0xb1, 0x2d, 0x11, 0x81, 0xeb, 0x85, 0x9e, 0x5a, 0x0a, 0xb3, 0xbb, 0x11, 0x7d, 0x75, 0xdb,
				}
				Expect(root).To(Equal(expected))
				Expect(root).To(HaveLen(32))
			})
		})

		Describe(".execBlock", func() {
			var block core.Block

			BeforeEach(func() {
				block = MakeTestBlock(bc, genesisChain, &core.GenerateBlockParams{
					Transactions: []core.Transaction{
						objects.NewTx(objects.TxTypeBalance, 1, util.String(receiver.Addr()), sender, "1", "2.36", 1532730724),
					},
					Creator:    sender,
					Nonce:      core.EncodeNonce(1),
					Difficulty: new(big.Int).SetInt64(131072),
				})
			})

			Context("when block has invalid transaction", func() {
				It("should return err ", func() {
					newSender := crypto.NewKeyFromIntSeed(3)
					block.GetTransactions()[0].SetFrom(util.String(newSender.Addr()))
					_, _, err := bc.execBlock(genesisChain, block)
					Expect(err).ToNot(BeNil())
					Expect(err.Error()).To(Equal("transaction error: index{0}: failed to get sender's account: account not found"))
				})
			})

			Context("sender and recipient are the same", func() {

				BeforeEach(func() {
					err = bc.CreateAccount(1, genesisChain, &objects.Account{
						Type:    objects.AccountTypeBalance,
						Address: util.String(sender.Addr()),
						Balance: "100",
					})
					Expect(err).To(BeNil())
				})

				When("block does not include fee allocation transaction", func() {

					BeforeEach(func() {
						block = MakeTestBlock(bc, genesisChain, &core.GenerateBlockParams{
							Transactions: []core.Transaction{
								objects.NewTx(objects.TxTypeBalance, 1, util.String(sender.Addr()), sender, "1", "2.36", 1532730724),
							},
							Creator:    sender,
							Nonce:      core.EncodeNonce(1),
							Difficulty: new(big.Int).SetInt64(131072),
						})
					})

					Specify("balance must be less than initial balance because no fee is paid back. The fee is lost", func() {
						_, stateObjs, err := bc.execBlock(genesisChain, block)
						Expect(err).To(BeNil())
						Expect(stateObjs).To(HaveLen(1))

						var m map[string]interface{}
						util.BytesToObject(stateObjs[0].Value, &m)
						Expect(m["balance"]).To(Equal("97.640000000000000000"))
					})
				})

				When("block includes a fee allocation transaction", func() {

					BeforeEach(func() {
						block = MakeTestBlock(bc, genesisChain, &core.GenerateBlockParams{
							Transactions: []core.Transaction{
								objects.NewTx(objects.TxTypeBalance, 1, util.String(sender.Addr()), sender, "1", "2.36", 1532730724),
							},
							Creator:     sender,
							Nonce:       core.EncodeNonce(1),
							Difficulty:  new(big.Int).SetInt64(131072),
							AddFeeAlloc: true,
						})
					})

					Specify("balance is equal to initial balance; fee is paid back; fee is not lost", func() {
						_, stateObjs, err := bc.execBlock(genesisChain, block)
						Expect(err).To(BeNil())
						Expect(stateObjs).To(HaveLen(1))

						var m map[string]interface{}
						util.BytesToObject(stateObjs[0].Value, &m)
						Expect(m["balance"]).To(Equal("100.000000000000000000"))
					})
				})
			})
		})

		Describe(".ProcessBlock", func() {

			var block core.Block

			BeforeEach(func() {
				block = MakeTestBlock(bc, genesisChain, &core.GenerateBlockParams{
					Transactions: []core.Transaction{
						objects.NewTx(objects.TxTypeBalance, 1, util.String(receiver.Addr()), sender, "1", "2.36", 1532730722),
						objects.NewTx(objects.TxTypeAlloc, 1, util.String(sender.Addr()), sender, "2.36", "0", 1532730722),
					},
					Creator:    sender,
					Nonce:      core.EncodeNonce(1),
					Difficulty: new(big.Int).SetInt64(131072),
				})
			})

			It("should reject the block if it has been added to the rejected cache", func() {
				bc.rejectedBlocks.Add(block.HashToHex(), struct{}{})
				_, err = bc.ProcessBlock(block)
				Expect(err).ToNot(BeNil())
				Expect(err).To(Equal(core.ErrBlockRejected))
			})

			It("should return error if block already exists in one of the known chains", func() {
				err = genesisChain.append(block)
				Expect(err).To(BeNil())
				_, err = bc.ProcessBlock(block)
				Expect(err).ToNot(BeNil())
				Expect(err).To(Equal(fmt.Errorf("block already exists")))
			})

			It("should return error if block has been added to the orphaned cache", func() {
				bc.orphanBlocks.Add(block.HashToHex(), block)
				_, err = bc.ProcessBlock(block)
				Expect(err).ToNot(BeNil())
				Expect(err).To(Equal(fmt.Errorf("orphan block")))
			})

			When("a block's parent does not exist in any chain", func() {

				BeforeEach(func() {
					block.GetHeader().SetParentHash(util.StrToHash("unknown"))
					block.SetHash(block.ComputeHash())
					blockSig, _ := objects.BlockSign(block, sender.PrivKey().Base58())
					block.SetSignature(blockSig)
				})

				It("should return nil and be added to the orphan block cache", func() {
					_, err = bc.ProcessBlock(block)
					Expect(err).To(BeNil())
					Expect(bc.orphanBlocks.Has(block.HashToHex())).To(BeTrue())
				})
			})

			When("a block's parent exists in a chain", func() {

				When("block's timestamp is lesser than its parent's timestamp", func() {

					var block2 core.Block

					BeforeEach(func() {
						block2 = MakeTestBlock(bc, genesisChain, &core.GenerateBlockParams{
							Transactions: []core.Transaction{
								objects.NewTx(objects.TxTypeBalance, 1, util.String(receiver.Addr()), sender, "1", "2.36", 1532730724),
								objects.NewTx(objects.TxTypeAlloc, 1, util.String(sender.Addr()), sender, "2.36", "0", 1532730724),
							},
							Creator:           sender,
							Nonce:             core.EncodeNonce(1),
							Difficulty:        new(big.Int).SetInt64(131072),
							OverrideTimestamp: genesisBlock.GetHeader().GetTimestamp() - 1,
						})
					})

					It("should return error", func() {
						_, err = bc.ProcessBlock(block2)
						Expect(err).ToNot(BeNil())
						Expect(err.Error()).To(Equal("block timestamp must be greater than its parent's"))
						Expect(bc.isRejected(block2)).To(BeTrue())
					})
				})

				When("block number is less than the chain tip block number", func() {

					var staleBlock2 objects.Block

					BeforeEach(func() {
						block2 := MakeTestBlock(bc, genesisChain, &core.GenerateBlockParams{
							Transactions: []core.Transaction{
								objects.NewTx(objects.TxTypeBalance, 1, util.String(receiver.Addr()), sender, "1", "2.36", 1532730724),
								objects.NewTx(objects.TxTypeAlloc, 1, util.String(sender.Addr()), sender, "2.36", "0", 1532730724),
							},
							Creator:           sender,
							Nonce:             core.EncodeNonce(1),
							Difficulty:        new(big.Int).SetInt64(131072),
							OverrideTimestamp: time.Now().Add(2 * time.Second).Unix(),
						})
						err = genesisChain.append(block2)
						Expect(err).To(BeNil())

						block3 := MakeTestBlock(bc, genesisChain, &core.GenerateBlockParams{
							Transactions: []core.Transaction{
								objects.NewTx(objects.TxTypeBalance, 1, util.String(receiver.Addr()), sender, "1", "2.36", 1532730725),
								objects.NewTx(objects.TxTypeAlloc, 1, util.String(sender.Addr()), sender, "2.36", "0", 1532730725),
							},
							Creator:           sender,
							Nonce:             core.EncodeNonce(2),
							Difficulty:        new(big.Int).SetInt64(131072),
							OverrideTimestamp: time.Now().Add(3 * time.Second).Unix(),
						})
						err = genesisChain.append(block3)
						Expect(err).To(BeNil())

						copier.Copy(&staleBlock2, block2)
						staleBlock2.GetHeader().SetNumber(2)
						staleBlock2.GetHeader().SetNonce(core.EncodeNonce(3))
						staleBlock2.Hash = staleBlock2.ComputeHash()
						staleBlock2.Sig, _ = objects.BlockSign(&staleBlock2, sender.PrivKey().Base58())
					})

					It("should return nil and create a new chain", func() {
						Expect(bc.chains).To(HaveLen(1))
						_, err = bc.ProcessBlock(&staleBlock2)
						Expect(err).To(BeNil())
						Expect(bc.chains).To(HaveLen(2))
					})
				})

				When("a block has same number as the chainTip", func() {
					var block2, block2_2 core.Block
					BeforeEach(func() {
						block2 = MakeTestBlock(bc, genesisChain, &core.GenerateBlockParams{
							Transactions: []core.Transaction{
								objects.NewTx(objects.TxTypeBalance, 1, util.String(receiver.Addr()), sender, "1", "2.36", 1532730724),
								objects.NewTx(objects.TxTypeAlloc, 1, util.String(sender.Addr()), sender, "2.36", "0", 1532730725),
							},
							Creator:    sender,
							Nonce:      core.EncodeNonce(1),
							Difficulty: new(big.Int).SetInt64(131072),
						})

						block2_2 = MakeTestBlock(bc, genesisChain, &core.GenerateBlockParams{
							Transactions: []core.Transaction{
								objects.NewTx(objects.TxTypeBalance, 1, util.String(receiver.Addr()), sender, "1", "2.36", 1532730725),
								objects.NewTx(objects.TxTypeAlloc, 1, util.String(sender.Addr()), sender, "2.36", "0", 1532730725),
							},
							Creator:    sender,
							Nonce:      core.EncodeNonce(4),
							Difficulty: new(big.Int).SetInt64(131072),
						})

						err = genesisChain.append(block2)
						Expect(err).To(BeNil())
					})

					It("should create a new chain tree; return nil; new tree's parent should be expected; tree must include the new block", func() {
						ch, err := bc.ProcessBlock(block2_2)
						Expect(err).To(BeNil())
						Expect(bc.chains).To(HaveLen(2))
						Expect(bc.chains[ch.GetID()].parentBlock.GetHash()).To(Equal(genesisBlock.GetHash()))
						hasBlock, err := bc.chains[ch.GetID()].hasBlock(block2_2.GetHash())
						Expect(err).To(BeNil())
						Expect(hasBlock).To(BeTrue())
					})
				})

				When("a block number is the same as the chain height", func() {

					var block2 core.Block

					BeforeEach(func() {

						block2 = MakeTestBlock(bc, genesisChain, &core.GenerateBlockParams{
							Transactions: []core.Transaction{
								objects.NewTx(objects.TxTypeBalance, 1, util.String(receiver.Addr()), sender, "1", "2.36", 1532730724),
								objects.NewTx(objects.TxTypeAlloc, 1, util.String(sender.Addr()), sender, "2.36", "0", 1532730725),
							},
							Creator:    sender,
							Nonce:      core.EncodeNonce(1),
							Difficulty: new(big.Int).SetInt64(131072),
						})

						block2.GetHeader().SetNumber(2)
						block2.SetHash(block2.ComputeHash())
						block3Sig, _ := objects.BlockSign(block2, sender.PrivKey().Base58())
						block2.SetSignature(block3Sig)

						err = genesisChain.append(block)
						Expect(err).To(BeNil())
					})

					It("should return nil and create a new chain", func() {
						Expect(bc.chains).To(HaveLen(1))
						_, err := bc.ProcessBlock(block2)
						Expect(err).To(BeNil())
						Expect(bc.chains).To(HaveLen(2))
					})
				})
			})

			Context("state root comparison", func() {

				var blockInvalidStateRoot, okStateRoot core.Block

				BeforeEach(func() {
					blockInvalidStateRoot = MakeTestBlock(bc, genesisChain, &core.GenerateBlockParams{
						Transactions: []core.Transaction{
							objects.NewTx(objects.TxTypeBalance, 1, util.String(receiver.Addr()), sender, "1", "2.36", 1532730724),
							objects.NewTx(objects.TxTypeAlloc, 1, util.String(sender.Addr()), sender, "2.36", "0", 1532730725),
						},
						Creator:    sender,
						Nonce:      core.EncodeNonce(1),
						Difficulty: new(big.Int).SetInt64(131072),
					})

					blockInvalidStateRoot.GetHeader().SetStateRoot(util.StrToHash("incorrect"))
					blockInvalidStateRoot.SetHash(blockInvalidStateRoot.ComputeHash())
					blockInvalidStateRootSig, _ := objects.BlockSign(blockInvalidStateRoot, sender.PrivKey().Base58())
					blockInvalidStateRoot.SetSignature(blockInvalidStateRootSig)

					okStateRoot = MakeTestBlock(bc, genesisChain, &core.GenerateBlockParams{
						Transactions: []core.Transaction{
							objects.NewTx(objects.TxTypeBalance, 1, util.String(receiver.Addr()), sender, "1", "2.36", 1532730724),
							objects.NewTx(objects.TxTypeAlloc, 1, util.String(sender.Addr()), sender, "2.36", "0", 1532730725),
						},
						Creator:    sender,
						Nonce:      core.EncodeNonce(1),
						Difficulty: new(big.Int).SetInt64(131072),
					})
				})

				It("should return error when block state root does not match", func() {
					_, err = bc.ProcessBlock(blockInvalidStateRoot)
					Expect(err).ToNot(BeNil())
					Expect(err).To(Equal(core.ErrBlockStateRootInvalid))
				})

				It("should successfully accept state root of block", func() {
					_, stateObjs, err := bc.execBlock(genesisChain, okStateRoot)
					Expect(err).To(BeNil())

					_, err = bc.ProcessBlock(okStateRoot)
					Expect(err).To(BeNil())

					Describe("chain should contain newly added block", func() {
						mBlock, err := genesisChain.getBlockByHash(okStateRoot.GetHash())
						Expect(err).To(BeNil())
						Expect(mBlock.Bytes()).To(Equal(okStateRoot.Bytes()))
					})

					Describe("all state objects must be persisted", func() {
						for _, so := range stateObjs {
							var result = db.GetByPrefix(so.Key)
							Expect(result).To(HaveLen(1))
						}
					})

					Describe("all transactions must be persisted", func() {
						for _, tx := range okStateRoot.GetTransactions() {
							txKey := common.MakeTxQueryKey(genesisChain.GetID().Bytes(), tx.GetHash().Bytes())
							var result = db.GetByPrefix(txKey)
							Expect(result).To(HaveLen(1))
						}
					})
				})
			})
		})

		Describe(".ProcessBlock: Test internal call of .processOrphanBlocks", func() {

			var parent1, orphanParent, orphan core.Block

			// add blocks in their correct other.
			// The tests will create new chains attempting to add the blocks
			// in different order just to test orphan handling. This other chains are named replayChain
			BeforeEach(func() {
				parent1 = MakeTestBlock(bc, genesisChain, &core.GenerateBlockParams{
					Transactions: []core.Transaction{
						objects.NewTx(objects.TxTypeBalance, 1, util.String(receiver.Addr()), sender, "1", "2.36", 1532730724),
						objects.NewTx(objects.TxTypeAlloc, 1, util.String(sender.Addr()), sender, "2.36", "0", 1532730725),
					},
					Creator:    sender,
					Nonce:      core.EncodeNonce(1),
					Difficulty: new(big.Int).SetInt64(131072),
				})
				_, err = bc.ProcessBlock(parent1)
				Expect(err).To(BeNil())

				orphanParent = MakeTestBlock(bc, genesisChain, &core.GenerateBlockParams{
					Transactions: []core.Transaction{
						objects.NewTx(objects.TxTypeBalance, 2, util.String(receiver.Addr()), sender, "1", "2.36", 1532730730),
						objects.NewTx(objects.TxTypeAlloc, 1, util.String(sender.Addr()), sender, "2.36", "0", 1532730731),
					},
					Creator:           sender,
					Nonce:             core.EncodeNonce(1),
					Difficulty:        new(big.Int).SetInt64(131072),
					OverrideTimestamp: time.Now().Add(2 * time.Second).Unix(),
				})
				_, err = bc.ProcessBlock(orphanParent)
				Expect(err).To(BeNil())

				orphan = MakeTestBlock(bc, genesisChain, &core.GenerateBlockParams{
					Transactions: []core.Transaction{
						objects.NewTx(objects.TxTypeBalance, 3, util.String(receiver.Addr()), sender, "1", "2.36", 1532730726),
						objects.NewTx(objects.TxTypeAlloc, 1, util.String(sender.Addr()), sender, "2.36", "0", 1532730727),
					},
					Creator:           sender,
					Nonce:             core.EncodeNonce(1),
					Difficulty:        new(big.Int).SetInt64(131072),
					OverrideTimestamp: time.Now().Add(4 * time.Second).Unix(),
				})
				_, err = bc.ProcessBlock(orphan)
				Expect(err).To(BeNil())
			})

			Context("with one orphan block", func() {

				var replayChain *Chain

				BeforeEach(func() {
					replayChain = NewChain("c2", db, cfg, log)
					bc.bestChain = replayChain
					bc.chains = map[util.String]*Chain{replayChain.id: replayChain}

					Expect(bc.CreateAccount(1, replayChain, &objects.Account{
						Type:    objects.AccountTypeBalance,
						Address: util.String(sender.Addr()),
						Balance: "1000",
					})).To(BeNil())

					err = replayChain.append(genesisBlock)
					Expect(err).To(BeNil())
				})

				BeforeEach(func() {
					_, err = bc.ProcessBlock(orphanParent)
					Expect(err).To(BeNil())
					Expect(bc.orphanBlocks.Len()).To(Equal(1))
				})

				It("should successfully add block and recursively process add all orphans when their parent exists in a chain", func() {
					_, err = bc.ProcessBlock(parent1)
					Expect(err).To(BeNil())
					Expect(replayChain.hasBlock(parent1.GetHash())).To(BeTrue())
					Expect(replayChain.hasBlock(orphanParent.GetHash())).To(BeTrue())
					Expect(bc.orphanBlocks.Len()).To(Equal(0))

					Describe("chain must contain the previously orphaned block as the tip", func() {
						has, err := replayChain.hasBlock(orphanParent.GetHash())
						Expect(err).To(BeNil())
						Expect(has).To(BeTrue())
						tipHeader, err := replayChain.Current()
						Expect(err).To(BeNil())
						Expect(tipHeader.ComputeHash()).To(Equal(orphanParent.GetHeader().ComputeHash()))
					})
				})
			})

			Context("with more than one orphan block", func() {

				var replayChain *Chain

				BeforeEach(func() {
					replayChain = NewChain("c2", db, cfg, log)
					bc.bestChain = replayChain
					bc.chains = map[util.String]*Chain{replayChain.id: replayChain}

					Expect(bc.CreateAccount(1, replayChain, &objects.Account{
						Type:    objects.AccountTypeBalance,
						Address: util.String(sender.Addr()),
						Balance: "1000",
					})).To(BeNil())

					err = replayChain.append(genesisBlock)
					Expect(err).To(BeNil())
				})

				BeforeEach(func() {
					_, err = bc.ProcessBlock(orphan)
					Expect(err).To(BeNil())
					_, err = bc.ProcessBlock(orphanParent)
					Expect(err).To(BeNil())
					Expect(bc.orphanBlocks.Len()).To(Equal(2))
				})

				It("should successfully add block and recursively process add all orphans when their parent exists in a chain", func() {
					_, err = bc.ProcessBlock(parent1)
					Expect(err).To(BeNil())
					Expect(replayChain.hasBlock(parent1.GetHash())).To(BeTrue())
					Expect(replayChain.hasBlock(orphanParent.GetHash())).To(BeTrue())
					Expect(replayChain.hasBlock(orphan.GetHash())).To(BeTrue())
					Expect(bc.orphanBlocks.Len()).To(Equal(0))

					Describe("chain must contain the previously orphaned block as the tip", func() {
						has, err := replayChain.hasBlock(orphan.GetHash())
						Expect(err).To(BeNil())
						Expect(has).To(BeTrue())
						tipHeader, err := replayChain.Current()
						Expect(err).To(BeNil())
						Expect(tipHeader.ComputeHash()).To(Equal(orphan.GetHeader().ComputeHash()))
					})
				})
			})
		})
	})
}
