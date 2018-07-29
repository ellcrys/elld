package blockchain

import (
	"fmt"

	"github.com/ellcrys/elld/blockchain/common"
	"github.com/ellcrys/elld/blockchain/testdata"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/database"
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

		Describe(".processTransactions", func() {

			var sender = crypto.NewKeyFromIntSeed(1)
			var account *wire.Account

			// BeforeEach(func() {
			// 	blk, err := bc.GenerateBlock(&GenerateBlockParams{
			// 		Transactions: []*wire.Transaction{
			// 			wire.NewTx(wire.TxTypeBalance, 123, receiver.Addr(), sender, "1", "0.1", 1532730722),
			// 		},
			// 		Creator:    sender,
			// 		Nonce:      384760,
			// 		MixHash:    "0x1cdf0e214bcdb7af36885316506f7388f262f7b710a28a00d21706550cdd72c2",
			// 		Difficulty: "102994",
			// 	}, ChainOp{Chain: chain})
			// 	Expect(err).To(BeNil())
			// 	Expect(blk).ToNot(BeNil())
			// 	pretty.Println(blk)
			// })

			BeforeEach(func() {
				account = &wire.Account{Type: wire.AccountTypeBalance, Address: sender.Addr(), Balance: "10"}
				err = bc.putAccount(1, chain, account)
				Expect(err).To(BeNil())
			})

			It("should return error if sender does not exist in the best chain", func() {
				_, err := bc.processTransactions(testdata.ProcessTransaction[0].Transactions, chain)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("index{0}: failed to get sender's account: account not found"))
			})

			It("should return error if sender account has insufficient value", func() {
				_, err := bc.processTransactions(testdata.ProcessTransaction[1].Transactions, chain)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("index{0}: insufficient sender account balance"))
			})

			It("should return error if sender value is invalid", func() {
				_, err := bc.processTransactions(testdata.ProcessTransaction[2].Transactions, chain)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("index{0}: sending amount error: can't convert 100_333 to decimal"))
			})

			Context("recipient does not have an account", func() {
				It(`should return operations and no error; expects 3 ops; 1 OpCreateAccount (for recipient) and 2 OpNewAccountBalance (sender and recipient);
							1st OpNewAccountBalance.Amount = 9; 2nd OpNewAccountBalance.Amount = 1`, func() {
					block := testdata.ProcessTransaction[3]

					ops, err := bc.processTransactions(block.Transactions, chain)
					Expect(err).To(BeNil())
					Expect(ops).To(HaveLen(3))

					Expect(ops[0]).To(BeAssignableToTypeOf(&common.OpCreateAccount{}))
					Expect(ops[0].Address()).To(Equal(block.Transactions[0].To))

					Expect(ops[1]).To(BeAssignableToTypeOf(&common.OpNewAccountBalance{}))
					Expect(ops[1].Address()).To(Equal(block.Transactions[0].From))
					Expect(ops[1].(*common.OpNewAccountBalance).Account.Balance).To(Equal("9"))

					Expect(ops[2]).To(BeAssignableToTypeOf(&common.OpNewAccountBalance{}))
					Expect(ops[2].Address()).To(Equal(block.Transactions[0].To))
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

					block := testdata.ProcessTransaction[4]

					ops, err := bc.processTransactions(block.Transactions, chain)
					Expect(err).To(BeNil())
					Expect(ops).To(HaveLen(2))

					Expect(ops[0]).To(BeAssignableToTypeOf(&common.OpNewAccountBalance{}))
					Expect(ops[0].Address()).To(Equal(block.Transactions[0].From))
					Expect(ops[0].(*common.OpNewAccountBalance).Account.Balance).To(Equal("9"))

					Expect(ops[1]).To(BeAssignableToTypeOf(&common.OpNewAccountBalance{}))
					Expect(ops[1].Address()).To(Equal(block.Transactions[0].To))
					Expect(ops[1].(*common.OpNewAccountBalance).Account.Balance).To(Equal("1"))
				})
			})

			Context("sender has two transactions sending to a recipient", func() {

				BeforeEach(func() {
					account = &wire.Account{Type: wire.AccountTypeBalance, Address: receiver.Addr(), Balance: "0"}
					err = bc.putAccount(1, chain, account)
					Expect(err).To(BeNil())
				})

				Context("recipient has an account", func() {
					It(`should return operations and no error; expects 2 ops; 2 OpNewAccountBalance (sender and recipient);
								1st OpNewAccountBalance.Amount = 8; 2nd OpNewAccountBalance.Amount = 2`, func() {
						block := testdata.ProcessTransaction[5]

						ops, err := bc.processTransactions(block.Transactions, chain)
						Expect(err).To(BeNil())
						Expect(ops).To(HaveLen(2))

						Expect(ops[0]).To(BeAssignableToTypeOf(&common.OpNewAccountBalance{}))
						Expect(ops[0].Address()).To(Equal(block.Transactions[0].From))
						Expect(ops[0].(*common.OpNewAccountBalance).Account.Balance).To(Equal("8"))

						Expect(ops[1]).To(BeAssignableToTypeOf(&common.OpNewAccountBalance{}))
						Expect(ops[1].Address()).To(Equal(block.Transactions[0].To))
						Expect(ops[1].(*common.OpNewAccountBalance).Account.Balance).To(Equal("2"))
					})
				})

				Context("recipient does not have an account", func() {
					It(`should return operations and no error; expects 3 ops; 1 OpCreateAccount (for recipient), 2 OpNewAccountBalance (sender and recipient);
									1st OpNewAccountBalance.Amount = 8; 2nd OpNewAccountBalance.Amount = 2`, func() {
						block := testdata.ProcessTransaction[6]

						ops, err := bc.processTransactions(block.Transactions, chain)
						Expect(err).To(BeNil())
						Expect(ops).To(HaveLen(3))

						Expect(ops[0]).To(BeAssignableToTypeOf(&common.OpCreateAccount{}))
						Expect(ops[0].Address()).To(Equal(block.Transactions[0].To))

						Expect(ops[1]).To(BeAssignableToTypeOf(&common.OpNewAccountBalance{}))
						Expect(ops[1].Address()).To(Equal(block.Transactions[0].From))
						Expect(ops[1].(*common.OpNewAccountBalance).Account.Balance).To(Equal("8"))

						Expect(ops[2]).To(BeAssignableToTypeOf(&common.OpNewAccountBalance{}))
						Expect(ops[2].Address()).To(Equal(block.Transactions[0].To))
						Expect(ops[2].(*common.OpNewAccountBalance).Account.Balance).To(Equal("2"))
					})
				})
			})
		})

		Describe(".ComputeTxsRoot", func() {
			It("should return expected root", func() {
				txs := []*wire.Transaction{
					&wire.Transaction{Sig: "b", Hash: util.ToHex([]byte("b"))},
					&wire.Transaction{Sig: "a", Hash: util.ToHex([]byte("a"))},
				}
				root := common.ComputeTxsRoot(txs)
				expected := []uint8{24, 12, 16, 250, 132, 0, 212, 108, 188, 32, 33, 161, 7, 17, 165, 111, 225, 219, 223, 95, 216, 242, 164, 222, 250, 56, 209, 66, 180, 143, 128, 36}
				Expect(root).To(Equal(expected))
				Expect(root).To(HaveLen(32))
			})
		})

		Describe(".execBlock", func() {
			var block *wire.Block

			Context("when block has invalid transaction", func() {

				BeforeEach(func() {
					block, _ = wire.BlockFromString(testdata.ProcessMockBlockData[0])
				})

				It("should return err ", func() {
					_, _, err := bc.execBlock(chain, block)
					Expect(err).ToNot(BeNil())
					Expect(err.Error()).To(Equal("transaction error: index{0}: failed to get sender's account: account not found"))
				})
			})
		})

		Describe(".ProcessBlock", func() {

			It("should reject the block if it has been added to the rejected cache", func() {
				block = testdata.Block2
				bc.rejectedBlocks.Add(block.GetHash(), struct{}{})
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
				bc.orphanBlocks.Add(block.GetHash(), block)
				_, err = bc.ProcessBlock(block)
				Expect(err).ToNot(BeNil())
				Expect(err).To(Equal(fmt.Errorf("error:block found in orphan cache")))
			})

			When("a block's parent does not exist in any chain", func() {
				It("should return nil and be added to the orphan block cache", func() {
					block = testdata.OrphanBlock1
					_, err = bc.ProcessBlock(block)
					Expect(err).To(BeNil())
					Expect(bc.orphanBlocks.Has(block.GetHash())).To(BeTrue())
				})
			})

			When("a block's parent exists in a chain", func() {
				var block2, block2_2, block3, block4 *wire.Block

				When("block number is less than the chain tip block number", func() {
					BeforeEach(func() {
						block2 = testdata.Stale1[0]
						err = chain.append(block2)
						Expect(err).To(BeNil())
						block2_2 = testdata.Stale1[1]
						block3 = testdata.Stale1[2]
						err = chain.append(block3)
						Expect(err).To(BeNil())
					})

					It("should return ErrVeryStaleBlock", func() {
						_, err = bc.ProcessBlock(block2_2)
						Expect(err).ToNot(BeNil())
						Expect(err).To(Equal(common.ErrVeryStaleBlock))
					})
				})

				When("a block has same number as the chainTip", func() {

					BeforeEach(func() {
						Expect(bc.chains).To(HaveLen(1))
						block2 = testdata.Stale1[0]
						err = chain.append(block2)
						Expect(err).To(BeNil())
						block2_2 = testdata.Stale1[1]
					})

					It("should create a new chain tree; return nil; new tree's parent should be expected; tree must include the new block", func() {
						ch, err := bc.ProcessBlock(block2_2)
						Expect(err).To(BeNil())
						Expect(bc.chains).To(HaveLen(2))
						Expect(bc.chains[ch.GetID()].parentBlock.Hash).To(Equal(block.Hash))
						hasBlock, err := bc.chains[ch.GetID()].hasBlock(block2_2.GetHash())
						Expect(err).To(BeNil())
						Expect(hasBlock).To(BeTrue())
					})
				})

				When("a block number is greater than chain tip block number by 1", func() {

					BeforeEach(func() {
						Expect(bc.chains).To(HaveLen(1))
						Expect(chain.append(testdata.Stale1[0])).To(BeNil())
						block4 = testdata.Stale1[3]
					})

					It("should return error", func() {
						_, err := bc.ProcessBlock(block4)
						Expect(err).ToNot(BeNil())
						Expect(err).To(Equal(common.ErrBlockFailedValidation))
					})
				})
			})

			Context("state root comparison", func() {

				It("should return error when block state root does not match", func() {
					_, err = bc.ProcessBlock(testdata.StateRoot1[0])
					Expect(err).ToNot(BeNil())
					Expect(err).To(Equal(common.ErrBlockStateRootInvalid))
				})

				It("should successfully accept state root of block", func() {
					block := testdata.StateRoot1[1]

					_, stateObjs, err := bc.execBlock(chain, block)
					Expect(err).To(BeNil())

					_, err = bc.ProcessBlock(block)
					Expect(err).To(BeNil())

					Describe("chain should contain newly added block", func() {
						mBlock, err := chain.getBlockByHash(block.GetHash())
						Expect(err).To(BeNil())
						Expect(mBlock).To(Equal(block))
					})

					Describe("all state objects must be persisted", func() {
						for _, so := range stateObjs {
							var result []*database.KVObject
							store.Get(so.Key, &result)
							Expect(result).To(HaveLen(1))
						}
					})

					Describe("all transactions must be persisted", func() {
						for _, tx := range block.Transactions {
							txKey := common.MakeTxKey(chain.GetID(), tx.ID())
							var result []*database.KVObject
							store.Get(txKey, &result)
							Expect(result).To(HaveLen(1))
						}
					})
				})
			})
		})

		Describe(".processOrphanBlocks", func() {

			var parent1, orphanParent, orphan, orphan2 *wire.Block

			BeforeEach(func() {
				parent1 = testdata.Orphans[0]
				orphanParent = testdata.Orphans[1]
				orphan = testdata.Orphans[2]
				orphan2 = testdata.Orphans[3]

				_, err = bc.ProcessBlock(parent1)
				Expect(err).To(BeNil())
			})

			Context("with one orphan block", func() {

				BeforeEach(func() {
					_, err = bc.ProcessBlock(orphan)
					Expect(err).To(BeNil())
					Expect(bc.orphanBlocks.Len()).To(Equal(1))
				})

				BeforeEach(func() {
					_, err = bc.ProcessBlock(orphanParent)
					Expect(err).To(BeNil())
					Expect(chain.hasBlock(orphanParent.Hash)).To(BeTrue())
				})

				It("should add orphan when its parent exists in a chain", func() {
					err = bc.processOrphanBlocks(orphanParent.Hash)
					Expect(err).To(BeNil())
					Expect(bc.orphanBlocks.Len()).To(Equal(0))

					Describe("chain must contain the previously orphaned block as the tip", func() {
						has, err := chain.hasBlock(orphan.Hash)
						Expect(err).To(BeNil())
						Expect(has).To(BeTrue())
						tipHeader, err := chain.getTipHeader()
						Expect(err).To(BeNil())
						Expect(tipHeader.ComputeHash()).To(Equal(orphan.Header.ComputeHash()))
					})
				})
			})

			Context("with more than one orphan block", func() {

				// add orphan blocks
				BeforeEach(func() {
					_, err = bc.ProcessBlock(orphan)
					Expect(err).To(BeNil())
					_, err = bc.ProcessBlock(orphan2)
					Expect(err).To(BeNil())
					Expect(bc.orphanBlocks.Len()).To(Equal(2))
				})

				// add the parent that is linked by one of the orphan
				BeforeEach(func() {
					_, err = bc.ProcessBlock(orphanParent)
					Expect(err).To(BeNil())
					Expect(chain.hasBlock(orphanParent.Hash)).To(BeTrue())
				})

				It("should recursively add all orphans when their parent exists in a chain", func() {
					err = bc.processOrphanBlocks(orphanParent.Hash)
					Expect(err).To(BeNil())
					Expect(bc.orphanBlocks.Len()).To(Equal(0))

					Describe("chain must contain the previously orphaned blocks and orphan2 as the tip", func() {
						has, err := chain.hasBlock(orphan.Hash)
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
