package blockchain

import (
	"fmt"

	"github.com/ellcrys/elld/blockchain/common"
	"github.com/ellcrys/elld/blockchain/testdata"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/elldb"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/wire"
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

			var sender = crypto.NewKeyFromIntSeed(1)
			var account *wire.Account

			BeforeEach(func() {
				account = &wire.Account{Type: wire.AccountTypeBalance, Address: sender.Addr(), Balance: "10"}
				err = bc.putAccount(1, chain, account)
				Expect(err).To(BeNil())
			})

			It("should return error if sender does not exist in the best chain", func() {
				_, err := bc.processTransactions(testdata.TransactionSet1[0], chain)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("index{0}: failed to get sender's account: account not found"))
			})

			It("should return error if sender account has insufficient value", func() {
				_, err := bc.processTransactions(testdata.TransactionSet1[1], chain)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("index{0}: insufficient sender account balance"))
			})

			It("should return error if sender value is invalid", func() {
				_, err := bc.processTransactions(testdata.TransactionSet1[2], chain)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("index{0}: sending amount error: can't convert 100_333 to decimal"))
			})

			Context("recipient does not have an account", func() {
				It(`should return operations and no error; expects 3 ops; 1 OpCreateAccount (for recipient) and 2 OpNewAccountBalance (sender and recipient);
									1st OpNewAccountBalance.Amount = 9; 2nd OpNewAccountBalance.Amount = 1`, func() {
					txs := testdata.TransactionSet1[3]

					ops, err := bc.processTransactions(txs, chain)
					Expect(err).To(BeNil())
					Expect(ops).To(HaveLen(3))

					Expect(ops[0]).To(BeAssignableToTypeOf(&common.OpCreateAccount{}))
					Expect(ops[0].Address()).To(Equal(txs[0].To))

					Expect(ops[1]).To(BeAssignableToTypeOf(&common.OpNewAccountBalance{}))
					Expect(ops[1].Address()).To(Equal(txs[0].From))
					Expect(ops[1].(*common.OpNewAccountBalance).Account.Balance).To(Equal("9"))

					Expect(ops[2]).To(BeAssignableToTypeOf(&common.OpNewAccountBalance{}))
					Expect(ops[2].Address()).To(Equal(txs[0].To))
					Expect(ops[2].(*common.OpNewAccountBalance).Account.Balance).To(Equal("1"))
				})
			})

			Context("recipient has an account", func() {
				receiver := crypto.NewKeyFromIntSeed(3)

				BeforeEach(func() {
					account = &wire.Account{Type: wire.AccountTypeBalance, Address: receiver.Addr(), Balance: "0"}
					err = bc.putAccount(1, chain, account)
					Expect(err).To(BeNil())
				})

				It(`should return operations and no error; expects 2 ops; 2 OpNewAccountBalance (sender and recipient);
											1st OpNewAccountBalance.Amount = 9; 2nd OpNewAccountBalance.Amount = 1`, func() {

					txs := testdata.TransactionSet1[4]

					ops, err := bc.processTransactions(txs, chain)
					Expect(err).To(BeNil())
					Expect(ops).To(HaveLen(2))

					Expect(ops[0]).To(BeAssignableToTypeOf(&common.OpNewAccountBalance{}))
					Expect(ops[0].Address()).To(Equal(txs[0].From))
					Expect(ops[0].(*common.OpNewAccountBalance).Account.Balance).To(Equal("9"))

					Expect(ops[1]).To(BeAssignableToTypeOf(&common.OpNewAccountBalance{}))
					Expect(ops[1].Address()).To(Equal(txs[0].To))
					Expect(ops[1].(*common.OpNewAccountBalance).Account.Balance).To(Equal("1"))
				})
			})

			Context("sender has two transactions sending to a recipient", func() {

				Context("recipient has an account", func() {
					receiver := crypto.NewKeyFromIntSeed(3)

					BeforeEach(func() {
						account = &wire.Account{Type: wire.AccountTypeBalance, Address: receiver.Addr(), Balance: "0"}
						err = bc.putAccount(1, chain, account)
						Expect(err).To(BeNil())
					})

					It(`should return operations and no error; expects 2 ops; 2 OpNewAccountBalance (sender and recipient);
											1st OpNewAccountBalance.Amount = 8; 2nd OpNewAccountBalance.Amount = 2`, func() {
						txs := testdata.TransactionSet1[5]

						ops, err := bc.processTransactions(txs, chain)
						Expect(err).To(BeNil())
						Expect(ops).To(HaveLen(2))

						Expect(ops[0]).To(BeAssignableToTypeOf(&common.OpNewAccountBalance{}))
						Expect(ops[0].Address()).To(Equal(txs[0].From))
						Expect(ops[0].(*common.OpNewAccountBalance).Account.Balance).To(Equal("8"))

						Expect(ops[1]).To(BeAssignableToTypeOf(&common.OpNewAccountBalance{}))
						Expect(ops[1].Address()).To(Equal(txs[0].To))
						Expect(ops[1].(*common.OpNewAccountBalance).Account.Balance).To(Equal("2"))
					})
				})

				Context("recipient does not have an account", func() {
					It(`should return operations and no error; expects 3 ops; 1 OpCreateAccount (for recipient), 2 OpNewAccountBalance (sender and recipient);
												1st OpNewAccountBalance.Amount = 8; 2nd OpNewAccountBalance.Amount = 2`, func() {
						txs := testdata.TransactionSet1[5]

						ops, err := bc.processTransactions(txs, chain)
						Expect(err).To(BeNil())
						Expect(ops).To(HaveLen(3))

						Expect(ops[0]).To(BeAssignableToTypeOf(&common.OpCreateAccount{}))
						Expect(ops[0].Address()).To(Equal(txs[0].To))

						Expect(ops[1]).To(BeAssignableToTypeOf(&common.OpNewAccountBalance{}))
						Expect(ops[1].Address()).To(Equal(txs[0].From))
						Expect(ops[1].(*common.OpNewAccountBalance).Account.Balance).To(Equal("8"))

						Expect(ops[2]).To(BeAssignableToTypeOf(&common.OpNewAccountBalance{}))
						Expect(ops[2].Address()).To(Equal(txs[0].To))
						Expect(ops[2].(*common.OpNewAccountBalance).Account.Balance).To(Equal("2"))
					})
				})
			})
		})

		Describe(".processTransactions (only TxTypeAllocCoin transactions)", func() {
			Context("only a single transaction", func() {
				When("recipient account does not exist", func() {
					It("should successfully return one state object = OpNewAccountBalance", func() {
						txs := testdata.TransactionSet2[0]
						ops, err := bc.processTransactions(txs, chain)
						Expect(err).To(BeNil())
						Expect(ops).To(HaveLen(1))
						Expect(ops[0]).To(BeAssignableToTypeOf(&common.OpNewAccountBalance{}))
						Expect(ops[0].(*common.OpNewAccountBalance).Account.Balance).To(Equal("10.0000000000000000"))
					})
				})

				When("recipient account already exists with account balance = 100", func() {
					BeforeEach(func() {
						Expect(bc.putAccount(1, chain, &wire.Account{
							Type:    wire.AccountTypeBalance,
							Address: testdata.TransactionSet2[0][0].To,
							Balance: "100",
						})).To(BeNil())
					})

					It("should successfully return one state object = OpNewAccountBalance and Balance = 110.0000000000000000", func() {
						txs := testdata.TransactionSet2[0]
						ops, err := bc.processTransactions(txs, chain)
						Expect(err).To(BeNil())
						Expect(ops).To(HaveLen(1))
						Expect(ops[0]).To(BeAssignableToTypeOf(&common.OpNewAccountBalance{}))
						Expect(ops[0].(*common.OpNewAccountBalance).Account.Balance).To(Equal("110.0000000000000000"))
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
				It("should return err ", func() {
					block := testdata.InvalidTxInBlockSet1[0]
					_, _, err := bc.execBlock(chain, block)
					Expect(err).ToNot(BeNil())
					Expect(err.Error()).To(Equal("transaction error: index{0}: failed to get sender's account: account not found"))
				})
			})
		})

		Describe(".ProcessBlock", func() {

			It("should reject the block if it has been added to the rejected cache", func() {
				block = testdata.Block2
				bc.rejectedBlocks.Add(block.HashToHex(), struct{}{})
				_, err = bc.ProcessBlock(block)
				Expect(err).ToNot(BeNil())
				Expect(err).To(Equal(common.ErrBlockRejected))
			})

			It("should return error if block already exists in one of the known chains", func() {
				block = testdata.Block2
				err = chain.append(block)
				Expect(err).To(BeNil())
				_, err = bc.ProcessBlock(block)
				Expect(err).ToNot(BeNil())
				Expect(err).To(Equal(fmt.Errorf("error:block found in chain")))
			})

			It("should return error if block has been added to the orphaned cache", func() {
				block = testdata.Block2
				bc.orphanBlocks.Add(block.HashToHex(), block)
				_, err = bc.ProcessBlock(block)
				Expect(err).ToNot(BeNil())
				Expect(err).To(Equal(fmt.Errorf("error:block found in orphan cache")))
			})

			When("a block's parent does not exist in any chain", func() {
				It("should return nil and be added to the orphan block cache", func() {
					block = testdata.BlockSet2[1]
					_, err = bc.ProcessBlock(block)
					Expect(err).To(BeNil())
					Expect(bc.orphanBlocks.Has(block.HashToHex())).To(BeTrue())
				})
			})

			When("a block's parent exists in a chain", func() {
				// var block2, block2_2, block3, block4 *wire.Block

				When("block number is less than the chain tip block number", func() {
					// BeforeEach(func() {
					// 	blk, err := bc.GenerateBlock(&GenerateBlockParams{
					// 		Transactions: []*wire.Transaction{
					// 			wire.NewTx(wire.TxTypeBalance, 123, receiver.Addr(), sender, "1", "0.1", 1532730722),
					// 		},
					// 		Creator:    sender,
					// 		Nonce:      wire.EncodeNonce(1),
					// 		MixHash:    util.BytesToHash([]byte("mix hash")),
					// 		Difficulty: new(big.Int).SetInt64(500),
					// 	}, ChainOp{Chain: chain})
					// 	Expect(err).To(BeNil())
					// 	Expect(blk).ToNot(BeNil())
					// 	pp.Println(blk)
					// })

					BeforeEach(func() {
						err = chain.append(testdata.BlockSet2[0])
						Expect(err).To(BeNil())
						err = chain.append(testdata.BlockSet2[1])
						Expect(err).To(BeNil())
					})

					It("should return ErrVeryStaleBlock", func() {
						_, err = bc.ProcessBlock(testdata.BlockSet2[2])
						Expect(err).ToNot(BeNil())
						Expect(err).To(Equal(common.ErrVeryStaleBlock))
					})
				})

				When("a block has same number as the chainTip", func() {
					BeforeEach(func() {
						err = chain.append(testdata.BlockSet2[0])
						Expect(err).To(BeNil())
					})

					It("should create a new chain tree; return nil; new tree's parent should be expected; tree must include the new block", func() {
						ch, err := bc.ProcessBlock(testdata.BlockSet2[2])
						Expect(err).To(BeNil())
						Expect(bc.chains).To(HaveLen(2))
						Expect(bc.chains[ch.GetID()].parentBlock.Hash).To(Equal(block.Hash))
						hasBlock, err := bc.chains[ch.GetID()].hasBlock(testdata.BlockSet2[2].HashToHex())
						Expect(err).To(BeNil())
						Expect(hasBlock).To(BeTrue())
					})
				})

				When("a block number is greater than chain tip block number by 1", func() {

					BeforeEach(func() {
						err = chain.append(testdata.BlockSet2[0])
						Expect(err).To(BeNil())
					})

					It("should return error", func() {
						_, err := bc.ProcessBlock(testdata.BlockSet2[3])
						Expect(err).ToNot(BeNil())
						Expect(err).To(Equal(common.ErrBlockFailedValidation))
					})
				})
			})

			Context("state root comparison", func() {

				It("should return error when block state root does not match", func() {
					_, err = bc.ProcessBlock(testdata.StateRootSet1[0])
					Expect(err).ToNot(BeNil())
					Expect(err).To(Equal(common.ErrBlockStateRootInvalid))
				})

				It("should successfully accept state root of block", func() {
					block := testdata.StateRootSet1[1]

					_, stateObjs, err := bc.execBlock(chain, block)
					Expect(err).To(BeNil())

					_, err = bc.ProcessBlock(block)
					Expect(err).To(BeNil())

					Describe("chain should contain newly added block", func() {
						mBlock, err := chain.getBlockByHash(block.GetHash().HexStr())
						Expect(err).To(BeNil())
						Expect(mBlock).To(Equal(block))
					})

					Describe("all state objects must be persisted", func() {
						for _, so := range stateObjs {
							var result []*elldb.KVObject
							testStore.Get(so.Key, &result)
							Expect(result).To(HaveLen(1))
						}
					})

					Describe("all transactions must be persisted", func() {
						for _, tx := range block.Transactions {
							txKey := common.MakeTxKey(chain.GetID(), tx.ID())
							var result []*elldb.KVObject
							testStore.Get(txKey, &result)
							Expect(result).To(HaveLen(1))
						}
					})
				})
			})
		})

		Describe(".ProcessBlock: Test internal call of .processOrphanBlocks", func() {

			var parent1, orphanParent, orphan, orphan2 *wire.Block

			BeforeEach(func() {
				parent1 = testdata.OrphanSet1[0]
				orphanParent = testdata.OrphanSet1[1]
				orphan = testdata.OrphanSet1[2]
				orphan2 = testdata.OrphanSet1[3]
				_, err = bc.ProcessBlock(parent1)
				Expect(err).To(BeNil())
			})

			Context("with one orphan block", func() {

				BeforeEach(func() {
					_, err = bc.ProcessBlock(orphan)
					Expect(err).To(BeNil())
					Expect(bc.orphanBlocks.Len()).To(Equal(1))
				})

				It("should successfully add block and recursively process add all orphans when their parent exists in a chain", func() {
					_, err = bc.ProcessBlock(orphanParent)
					Expect(err).To(BeNil())
					Expect(chain.hasBlock(orphanParent.HashToHex())).To(BeTrue())
					Expect(bc.orphanBlocks.Len()).To(Equal(0))

					Describe("chain must contain the previously orphaned block as the tip", func() {
						has, err := chain.hasBlock(orphan.HashToHex())
						Expect(err).To(BeNil())
						Expect(has).To(BeTrue())
						tipHeader, err := chain.getTipHeader()
						Expect(err).To(BeNil())
						Expect(tipHeader.ComputeHash()).To(Equal(orphan.Header.ComputeHash()))
					})
				})
			})

			Context("with more than one orphan block", func() {

				BeforeEach(func() {
					_, err = bc.ProcessBlock(orphan)
					Expect(err).To(BeNil())
					_, err = bc.ProcessBlock(orphan2)
					Expect(err).To(BeNil())
					Expect(bc.orphanBlocks.Len()).To(Equal(2))
				})

				It("should successfully add block and recursively process add all orphans when their parent exists in a chain", func() {
					_, err = bc.ProcessBlock(orphanParent)
					Expect(err).To(BeNil())
					Expect(chain.hasBlock(orphanParent.HashToHex())).To(BeTrue())
					Expect(bc.orphanBlocks.Len()).To(Equal(0))

					Describe("chain must contain the previously orphaned blocks and orphan2 as the tip", func() {
						has, err := chain.hasBlock(orphan.HashToHex())
						Expect(err).To(BeNil())
						Expect(has).To(BeTrue())
						tipHeader, err := chain.getTipHeader()
						Expect(err).To(BeNil())
						Expect(tipHeader.ComputeHash()).To(Equal(orphan2.Header.ComputeHash()))
					})
				})
			})
		})
	})
}
