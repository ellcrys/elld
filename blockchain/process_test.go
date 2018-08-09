package blockchain

import (
	"fmt"
	"math/big"

	"github.com/ellcrys/elld/blockchain/common"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/wire"
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
					Account: &wire.Account{
						Balance: "10",
					},
				},
			}

			It("should add an additional op successfully", func() {
				op := &common.OpNewAccountBalance{
					OpBase: &common.OpBase{Addr: "addr2"},
					Account: &wire.Account{
						Balance: "10",
					},
				}
				newOps := addOp(curOps, op)
				Expect(newOps).To(HaveLen(2))
			})

			It("should replace any equal op found", func() {
				op := &common.OpNewAccountBalance{
					OpBase: &common.OpBase{Addr: "addr1"},
					Account: &wire.Account{
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

			var account *wire.Account

			BeforeEach(func() {
				account = &wire.Account{Type: wire.AccountTypeBalance, Address: util.String(sender.Addr()), Balance: "10"}
				err = bc.putAccount(1, genesisChain, account)
				Expect(err).To(BeNil())
			})

			It("should return error if sender does not exist in the best chain", func() {
				var txs = []*wire.Transaction{
					&wire.Transaction{Type: 1, Nonce: 123, To: "e6i7rxApBYUt7w94gGDKTz45A5J567JfkS", From: "unknown", SenderPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
						Value: "1", Timestamp: 1532730724,
						Fee: "0.1", Sig: []uint8{},
						Hash: util.Hash{},
					},
				}

				_, err := bc.processTransactions(txs, genesisChain)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("index{0}: failed to get sender's account: account not found"))
			})

			It("should return error if sender account has insufficient value", func() {
				var txs = []*wire.Transaction{
					&wire.Transaction{Type: 1, Nonce: 123, To: "e6i7rxApBYUt7w94gGDKTz45A5J567JfkS", From: "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad", SenderPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
						Value: "10000000", Timestamp: 1532730724,
						Fee: "0.1", Sig: []uint8{},
						Hash: util.Hash{},
					},
				}

				_, err := bc.processTransactions(txs, genesisChain)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("index{0}: insufficient sender account balance"))
			})

			It("should panic if sender value is could not be converted to decimal", func() {
				var txs = []*wire.Transaction{
					&wire.Transaction{Type: 1, Nonce: 123, To: "e6i7rxApBYUt7w94gGDKTz45A5J567JfkS", From: "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad", SenderPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
						Value: "100_333", Timestamp: 1532730724,
						Fee: "0.1", Sig: []uint8{},
						Hash: util.Hash{},
					},
				}
				Expect(func() {
					bc.processTransactions(txs, genesisChain)
				}).To(Panic())
			})

			Context("recipient does not have an account", func() {

				It(`should return operations and no error; expects 3 ops; 1 OpCreateAccount (for recipient) and 2 OpNewAccountBalance (sender and recipient);
					1st OpNewAccountBalance.Amount = 9; 2nd OpNewAccountBalance.Amount = 1`, func() {
					var txs = []*wire.Transaction{
						&wire.Transaction{Type: 1, Nonce: 123, To: "e6i7rxApBYUt7w94gGDKTz45A5J567JfkS", From: "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad", SenderPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
							Value: "1", Timestamp: 1532730724,
							Fee: "0.1", Sig: []uint8{},
							Hash: util.Hash{},
						},
					}

					ops, err := bc.processTransactions(txs, genesisChain)
					Expect(err).To(BeNil())
					Expect(ops).To(HaveLen(3))

					Expect(ops[0]).To(BeAssignableToTypeOf(&common.OpCreateAccount{}))
					Expect(ops[0].Address()).To(Equal(txs[0].To))

					Expect(ops[1]).To(BeAssignableToTypeOf(&common.OpNewAccountBalance{}))
					Expect(ops[1].Address()).To(Equal(txs[0].From))
					Expect(ops[1].(*common.OpNewAccountBalance).Account.Balance).To(Equal(util.String("9.0000000000000000")))

					Expect(ops[2]).To(BeAssignableToTypeOf(&common.OpNewAccountBalance{}))
					Expect(ops[2].Address()).To(Equal(txs[0].To))
					Expect(ops[2].(*common.OpNewAccountBalance).Account.Balance).To(Equal(util.String("1.0000000000000000")))
				})
			})

			Context("recipient has an account", func() {
				receiver := crypto.NewKeyFromIntSeed(3)

				BeforeEach(func() {
					account = &wire.Account{Type: wire.AccountTypeBalance, Address: util.String(receiver.Addr()), Balance: "0"}
					err = bc.putAccount(1, genesisChain, account)
					Expect(err).To(BeNil())
				})

				It(`should return operations and no error; expects 2 ops; 2 OpNewAccountBalance (sender and recipient);
					1st OpNewAccountBalance.Amount = 9; 2nd OpNewAccountBalance.Amount = 1`, func() {

					var txs = []*wire.Transaction{
						&wire.Transaction{Type: 1, Nonce: 123, To: util.String(receiver.Addr()), From: "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad", SenderPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
							Value: "1", Timestamp: 1532730724,
							Fee: "0.1", Sig: []uint8{},
							Hash: util.Hash{},
						},
					}

					ops, err := bc.processTransactions(txs, genesisChain)
					Expect(err).To(BeNil())
					Expect(ops).To(HaveLen(2))

					Expect(ops[0]).To(BeAssignableToTypeOf(&common.OpNewAccountBalance{}))
					Expect(ops[0].Address()).To(Equal(txs[0].From))
					Expect(ops[0].(*common.OpNewAccountBalance).Account.Balance).To(Equal(util.String("9.0000000000000000")))

					Expect(ops[1]).To(BeAssignableToTypeOf(&common.OpNewAccountBalance{}))
					Expect(ops[1].Address()).To(Equal(txs[0].To))
					Expect(ops[1].(*common.OpNewAccountBalance).Account.Balance).To(Equal(util.String("1.0000000000000000")))
				})
			})

			Context("sender has two transactions sending to a recipient", func() {

				Context("recipient has an account", func() {
					receiver := crypto.NewKeyFromIntSeed(3)

					BeforeEach(func() {
						account = &wire.Account{Type: wire.AccountTypeBalance, Address: util.String(receiver.Addr()), Balance: "0"}
						err = bc.putAccount(1, genesisChain, account)
						Expect(err).To(BeNil())
					})

					It(`should return operations and no error; expects 2 ops; 2 OpNewAccountBalance (sender and recipient);
													1st OpNewAccountBalance.Amount = 8; 2nd OpNewAccountBalance.Amount = 2`, func() {
						var txs = []*wire.Transaction{
							&wire.Transaction{Type: 1, Nonce: 123, To: util.String(receiver.Addr()), From: "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad", SenderPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
								Value: "1", Timestamp: 1532730724,
								Fee: "0.1", Sig: []uint8{},
								Hash: util.Hash{},
							},
							&wire.Transaction{Type: 1, Nonce: 123, To: util.String(receiver.Addr()), From: "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad", SenderPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
								Value: "1", Timestamp: 1532730724,
								Fee: "0.1", Sig: []uint8{},
								Hash: util.Hash{},
							},
						}

						ops, err := bc.processTransactions(txs, genesisChain)
						Expect(err).To(BeNil())
						Expect(ops).To(HaveLen(2))

						Expect(ops[0]).To(BeAssignableToTypeOf(&common.OpNewAccountBalance{}))
						Expect(ops[0].Address()).To(Equal(txs[0].From))
						Expect(ops[0].(*common.OpNewAccountBalance).Account.Balance).To(Equal(util.String("8.0000000000000000")))

						Expect(ops[1]).To(BeAssignableToTypeOf(&common.OpNewAccountBalance{}))
						Expect(ops[1].Address()).To(Equal(txs[0].To))
						Expect(ops[1].(*common.OpNewAccountBalance).Account.Balance).To(Equal(util.String("2.0000000000000000")))
					})
				})

				Context("recipient does not have an account", func() {
					It(`should return operations and no error; expects 3 ops; 1 OpCreateAccount (for recipient), 2 OpNewAccountBalance (sender and recipient);
														1st OpNewAccountBalance.Amount = 8; 2nd OpNewAccountBalance.Amount = 2`, func() {
						var txs = []*wire.Transaction{
							&wire.Transaction{Type: 1, Nonce: 123, To: util.String(receiver.Addr()), From: "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad", SenderPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
								Value: "1", Timestamp: 1532730724,
								Fee: "0.1", Sig: []uint8{},
								Hash: util.Hash{},
							},
							&wire.Transaction{Type: 1, Nonce: 123, To: util.String(receiver.Addr()), From: "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad", SenderPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
								Value: "1", Timestamp: 1532730724,
								Fee: "0.1", Sig: []uint8{},
								Hash: util.Hash{},
							},
						}

						ops, err := bc.processTransactions(txs, genesisChain)
						Expect(err).To(BeNil())
						Expect(ops).To(HaveLen(3))

						Expect(ops[0]).To(BeAssignableToTypeOf(&common.OpCreateAccount{}))
						Expect(ops[0].Address()).To(Equal(txs[0].To))

						Expect(ops[1]).To(BeAssignableToTypeOf(&common.OpNewAccountBalance{}))
						Expect(ops[1].Address()).To(Equal(txs[0].From))
						Expect(ops[1].(*common.OpNewAccountBalance).Account.Balance).To(Equal(util.String("8.0000000000000000")))

						Expect(ops[2]).To(BeAssignableToTypeOf(&common.OpNewAccountBalance{}))
						Expect(ops[2].Address()).To(Equal(txs[0].To))
						Expect(ops[2].(*common.OpNewAccountBalance).Account.Balance).To(Equal(util.String("2.0000000000000000")))
					})
				})
			})
		})

		Describe(".processTransactions (only TxTypeAllocCoin transactions)", func() {
			Context("only a single transaction", func() {
				When("recipient account does not exist", func() {
					It("should successfully return one state object = OpNewAccountBalance", func() {
						var txs = []*wire.Transaction{
							&wire.Transaction{Type: wire.TxTypeAllocCoin, Nonce: 123, To: "e6i7rxApBYUt7w94gGDKTz45A5J567JfkS", From: "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad", SenderPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
								Value: "10", Timestamp: 1532730724,
								Fee: "0.1", Sig: []uint8{},
								Hash: util.Hash{},
							},
						}

						ops, err := bc.processTransactions(txs, genesisChain)
						Expect(err).To(BeNil())
						Expect(ops).To(HaveLen(1))
						Expect(ops[0]).To(BeAssignableToTypeOf(&common.OpNewAccountBalance{}))
						Expect(ops[0].(*common.OpNewAccountBalance).Account.Balance).To(Equal(util.String("10.0000000000000000")))
					})
				})

				When("recipient account already exists with account balance = 100", func() {
					BeforeEach(func() {
						Expect(bc.putAccount(1, genesisChain, &wire.Account{
							Type:    wire.AccountTypeBalance,
							Address: "e6i7rxApBYUt7w94gGDKTz45A5J567JfkS",
							Balance: "100",
						})).To(BeNil())
					})

					It("should successfully return one state object = OpNewAccountBalance and Balance = 110.0000000000000000", func() {
						var txs = []*wire.Transaction{
							&wire.Transaction{Type: wire.TxTypeAllocCoin, Nonce: 123, To: "e6i7rxApBYUt7w94gGDKTz45A5J567JfkS", From: "eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad", SenderPubKey: "48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC",
								Value: "10", Timestamp: 1532730724,
								Fee: "0.1", Sig: []uint8{},
								Hash: util.Hash{},
							},
						}
						ops, err := bc.processTransactions(txs, genesisChain)
						Expect(err).To(BeNil())
						Expect(ops).To(HaveLen(1))
						Expect(ops[0]).To(BeAssignableToTypeOf(&common.OpNewAccountBalance{}))
						Expect(ops[0].(*common.OpNewAccountBalance).Account.Balance).To(Equal(util.String("110.0000000000000000")))
					})
				})
			})
		})

		Describe(".ComputeTxsRoot", func() {
			It("should return expected root", func() {
				txs := []*wire.Transaction{
					&wire.Transaction{Sig: []byte("b"), Hash: util.StrToHash("hash_1")},
					&wire.Transaction{Sig: []byte("a"), Hash: util.StrToHash("hash_2")},
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
			Context("when block has invalid transaction", func() {

				var block *wire.Block

				BeforeEach(func() {
					block = MakeTestBlock(bc, genesisChain, &common.GenerateBlockParams{
						Transactions: []*wire.Transaction{
							wire.NewTx(wire.TxTypeBalance, 123, util.String(receiver.Addr()), sender, "1", "0.1", 1532730724),
						},
						Creator:    sender,
						Nonce:      wire.EncodeNonce(1),
						MixHash:    util.BytesToHash([]byte("mix hash")),
						Difficulty: new(big.Int).SetInt64(500),
					})
				})

				It("should return err ", func() {
					newSender := crypto.NewKeyFromIntSeed(3)
					block.Transactions[0].From = util.String(newSender.Addr())
					_, _, err := bc.execBlock(genesisChain, block)
					Expect(err).ToNot(BeNil())
					Expect(err.Error()).To(Equal("transaction error: index{0}: failed to get sender's account: account not found"))
				})
			})
		})

		Describe(".ProcessBlock", func() {

			var block *wire.Block

			BeforeEach(func() {
				block = MakeTestBlock(bc, genesisChain, &common.GenerateBlockParams{
					Transactions: []*wire.Transaction{
						wire.NewTx(wire.TxTypeBalance, 123, util.String(receiver.Addr()), sender, "1", "0.1", 1532730722),
					},
					Creator:    sender,
					Nonce:      wire.EncodeNonce(1),
					MixHash:    util.BytesToHash([]byte("mix hash")),
					Difficulty: new(big.Int).SetInt64(500),
				})
			})

			It("should reject the block if it has been added to the rejected cache", func() {
				bc.rejectedBlocks.Add(block.HashToHex(), struct{}{})
				_, err = bc.ProcessBlock(block)
				Expect(err).ToNot(BeNil())
				Expect(err).To(Equal(common.ErrBlockRejected))
			})

			It("should return error if block already exists in one of the known chains", func() {
				err = genesisChain.append(block)
				Expect(err).To(BeNil())
				_, err = bc.ProcessBlock(block)
				Expect(err).ToNot(BeNil())
				Expect(err).To(Equal(fmt.Errorf("error:block found in chain")))
			})

			It("should return error if block has been added to the orphaned cache", func() {
				bc.orphanBlocks.Add(block.HashToHex(), block)
				_, err = bc.ProcessBlock(block)
				Expect(err).ToNot(BeNil())
				Expect(err).To(Equal(fmt.Errorf("error:block found in orphan cache")))
			})

			When("a block's parent does not exist in any chain", func() {

				BeforeEach(func() {
					block.Header.ParentHash = util.StrToHash("unknown")
					block.Hash = block.ComputeHash()
					block.Sig, _ = wire.BlockSign(block, sender.PrivKey().Base58())

				})

				It("should return nil and be added to the orphan block cache", func() {
					_, err = bc.ProcessBlock(block)
					Expect(err).To(BeNil())
					Expect(bc.orphanBlocks.Has(block.HashToHex())).To(BeTrue())
				})
			})

			When("a block's parent exists in a chain", func() {
				// var block2, block2_2, block3, block4 *wire.Block

				When("block number is less than the chain tip block number", func() {

					var staleBlock2 wire.Block

					BeforeEach(func() {
						block2 := MakeTestBlock(bc, genesisChain, &common.GenerateBlockParams{
							Transactions: []*wire.Transaction{
								wire.NewTx(wire.TxTypeBalance, 123, util.String(receiver.Addr()), sender, "1", "0.1", 1532730724),
							},
							Creator:    sender,
							Nonce:      wire.EncodeNonce(1),
							MixHash:    util.BytesToHash([]byte("mix hash")),
							Difficulty: new(big.Int).SetInt64(500),
						})
						err = genesisChain.append(block2)
						Expect(err).To(BeNil())

						block3 := MakeTestBlock(bc, genesisChain, &common.GenerateBlockParams{
							Transactions: []*wire.Transaction{
								wire.NewTx(wire.TxTypeBalance, 123, util.String(receiver.Addr()), sender, "1", "0.1", 1532730725),
							},
							Creator:    sender,
							Nonce:      wire.EncodeNonce(1),
							MixHash:    util.BytesToHash([]byte("mix hash 2")),
							Difficulty: new(big.Int).SetInt64(500),
						})
						err = genesisChain.append(block3)
						Expect(err).To(BeNil())

						copier.Copy(&staleBlock2, block2)
						staleBlock2.Header.Number = 2
						staleBlock2.Header.MixHash = util.BytesToHash([]byte("mix hash 3"))
						staleBlock2.Hash = staleBlock2.ComputeHash()
						staleBlock2.Sig, _ = wire.BlockSign(&staleBlock2, sender.PrivKey().Base58())
					})

					It("should return ErrVeryStaleBlock", func() {
						_, err = bc.ProcessBlock(&staleBlock2)
						Expect(err).ToNot(BeNil())
						Expect(err).To(Equal(common.ErrVeryStaleBlock))
					})
				})

				When("a block has same number as the chainTip", func() {
					var block2, block2_2 *wire.Block
					BeforeEach(func() {
						block2 = MakeTestBlock(bc, genesisChain, &common.GenerateBlockParams{
							Transactions: []*wire.Transaction{
								wire.NewTx(wire.TxTypeBalance, 123, util.String(receiver.Addr()), sender, "1", "0.1", 1532730724),
							},
							Creator:    sender,
							Nonce:      wire.EncodeNonce(1),
							MixHash:    util.BytesToHash([]byte("mix hash")),
							Difficulty: new(big.Int).SetInt64(500),
						})

						block2_2 = MakeTestBlock(bc, genesisChain, &common.GenerateBlockParams{
							Transactions: []*wire.Transaction{
								wire.NewTx(wire.TxTypeBalance, 123, util.String(receiver.Addr()), sender, "1", "0.1", 1532730725),
							},
							Creator:    sender,
							Nonce:      wire.EncodeNonce(4),
							MixHash:    util.BytesToHash([]byte("mix hash 2")),
							Difficulty: new(big.Int).SetInt64(500),
						})

						err = genesisChain.append(block2)
						Expect(err).To(BeNil())
					})

					It("should create a new chain tree; return nil; new tree's parent should be expected; tree must include the new block", func() {
						ch, err := bc.ProcessBlock(block2_2)
						Expect(err).To(BeNil())
						Expect(bc.chains).To(HaveLen(2))
						Expect(bc.chains[ch.GetID()].parentBlock.Hash).To(Equal(genesisBlock.Hash))
						hasBlock, err := bc.chains[ch.GetID()].hasBlock(block2_2.Hash)
						Expect(err).To(BeNil())
						Expect(hasBlock).To(BeTrue())
					})
				})

				When("a block number is greater than chain tip block number by 1", func() {

					var block3 *wire.Block

					BeforeEach(func() {
						block3 = MakeTestBlock(bc, genesisChain, &common.GenerateBlockParams{
							Transactions: []*wire.Transaction{
								wire.NewTx(wire.TxTypeBalance, 123, util.String(receiver.Addr()), sender, "1", "0.1", 1532730724),
							},
							Creator:    sender,
							Nonce:      wire.EncodeNonce(1),
							MixHash:    util.BytesToHash([]byte("mix hash")),
							Difficulty: new(big.Int).SetInt64(500),
						})

						block3.Header.Number = 3
						block3.Hash = block3.ComputeHash()
						block3.Sig, _ = wire.BlockSign(block3, sender.PrivKey().Base58())
					})

					It("should return error", func() {
						_, err := bc.ProcessBlock(block3)
						Expect(err).ToNot(BeNil())
						Expect(err).To(Equal(common.ErrBlockFailedValidation))
					})
				})
			})

			Context("state root comparison", func() {

				var blockInvalidStateRoot, okStateRoot *wire.Block

				BeforeEach(func() {
					blockInvalidStateRoot = MakeTestBlock(bc, genesisChain, &common.GenerateBlockParams{
						Transactions: []*wire.Transaction{
							wire.NewTx(wire.TxTypeBalance, 123, util.String(receiver.Addr()), sender, "1", "0.1", 1532730724),
						},
						Creator:    sender,
						Nonce:      wire.EncodeNonce(1),
						MixHash:    util.BytesToHash([]byte("mix hash")),
						Difficulty: new(big.Int).SetInt64(500),
					})

					blockInvalidStateRoot.Header.StateRoot = util.StrToHash("incorrect")
					blockInvalidStateRoot.Hash = blockInvalidStateRoot.ComputeHash()
					blockInvalidStateRoot.Sig, _ = wire.BlockSign(blockInvalidStateRoot, sender.PrivKey().Base58())

					okStateRoot = MakeTestBlock(bc, genesisChain, &common.GenerateBlockParams{
						Transactions: []*wire.Transaction{
							wire.NewTx(wire.TxTypeBalance, 123, util.String(receiver.Addr()), sender, "1", "0.1", 1532730724),
						},
						Creator:    sender,
						Nonce:      wire.EncodeNonce(1),
						MixHash:    util.BytesToHash([]byte("mix hash")),
						Difficulty: new(big.Int).SetInt64(500),
					})
				})

				It("should return error when block state root does not match", func() {
					_, err = bc.ProcessBlock(blockInvalidStateRoot)
					Expect(err).ToNot(BeNil())
					Expect(err).To(Equal(common.ErrBlockStateRootInvalid))
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
						for _, tx := range okStateRoot.Transactions {
							txKey := common.MakeTxKey(genesisChain.GetID().Bytes(), tx.Hash.Bytes())
							var result = db.GetByPrefix(txKey)
							Expect(result).To(HaveLen(1))
						}
					})
				})
			})
		})

		Describe(".ProcessBlock: Test internal call of .processOrphanBlocks", func() {

			var parent1, orphanParent, orphan *wire.Block

			// add blocks in their correct other.
			// The tests will create new chains attempting to add the blocks
			// in different order just to test orphan handling. This other chains are named replayChain
			BeforeEach(func() {
				parent1 = MakeTestBlock(bc, genesisChain, &common.GenerateBlockParams{
					Transactions: []*wire.Transaction{
						wire.NewTx(wire.TxTypeBalance, 123, util.String(receiver.Addr()), sender, "1", "0.1", 1532730724),
					},
					Creator:    sender,
					Nonce:      wire.EncodeNonce(1),
					MixHash:    util.BytesToHash([]byte("mix hash")),
					Difficulty: new(big.Int).SetInt64(500),
				})
				_, err = bc.ProcessBlock(parent1)
				Expect(err).To(BeNil())

				orphanParent = MakeTestBlock(bc, genesisChain, &common.GenerateBlockParams{
					Transactions: []*wire.Transaction{
						wire.NewTx(wire.TxTypeBalance, 123, util.String(receiver.Addr()), sender, "1", "0.1", 1532730730),
					},
					Creator:    sender,
					Nonce:      wire.EncodeNonce(1),
					MixHash:    util.BytesToHash([]byte("mix hash")),
					Difficulty: new(big.Int).SetInt64(500),
				})
				_, err = bc.ProcessBlock(orphanParent)
				Expect(err).To(BeNil())

				orphan = MakeTestBlock(bc, genesisChain, &common.GenerateBlockParams{
					Transactions: []*wire.Transaction{
						wire.NewTx(wire.TxTypeBalance, 123, util.String(receiver.Addr()), sender, "1", "0.1", 1532730726),
					},
					Creator:    sender,
					Nonce:      wire.EncodeNonce(1),
					MixHash:    util.BytesToHash([]byte("mix hash")),
					Difficulty: new(big.Int).SetInt64(500),
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

					Expect(bc.putAccount(1, replayChain, &wire.Account{
						Type:    wire.AccountTypeBalance,
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
					Expect(replayChain.hasBlock(parent1.Hash)).To(BeTrue())
					Expect(replayChain.hasBlock(orphanParent.Hash)).To(BeTrue())
					Expect(bc.orphanBlocks.Len()).To(Equal(0))

					Describe("chain must contain the previously orphaned block as the tip", func() {
						has, err := replayChain.hasBlock(orphanParent.Hash)
						Expect(err).To(BeNil())
						Expect(has).To(BeTrue())
						tipHeader, err := replayChain.Current()
						Expect(err).To(BeNil())
						Expect(tipHeader.ComputeHash()).To(Equal(orphanParent.Header.ComputeHash()))
					})
				})
			})

			Context("with more than one orphan block", func() {

				var replayChain *Chain

				BeforeEach(func() {
					replayChain = NewChain("c2", db, cfg, log)
					bc.bestChain = replayChain
					bc.chains = map[util.String]*Chain{replayChain.id: replayChain}

					Expect(bc.putAccount(1, replayChain, &wire.Account{
						Type:    wire.AccountTypeBalance,
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
					Expect(replayChain.hasBlock(parent1.Hash)).To(BeTrue())
					Expect(replayChain.hasBlock(orphanParent.Hash)).To(BeTrue())
					Expect(replayChain.hasBlock(orphan.Hash)).To(BeTrue())
					Expect(bc.orphanBlocks.Len()).To(Equal(0))

					Describe("chain must contain the previously orphaned block as the tip", func() {
						has, err := replayChain.hasBlock(orphan.Hash)
						Expect(err).To(BeNil())
						Expect(has).To(BeTrue())
						tipHeader, err := replayChain.Current()
						Expect(err).To(BeNil())
						Expect(tipHeader.ComputeHash()).To(Equal(orphan.Header.ComputeHash()))
					})
				})
			})
		})
	})
}
