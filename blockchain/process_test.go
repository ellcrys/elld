package blockchain

import (
	"fmt"

	"github.com/ellcrys/elld/blockchain/common"
	"github.com/ellcrys/elld/blockchain/leveldb"
	"github.com/ellcrys/elld/blockchain/testdata"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/database"
	"github.com/ellcrys/elld/testutil"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/wire"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Blockchain", func() {

	var err error
	var store common.Store
	var db database.DB
	var chainID = "chain1"
	var chain *Chain
	var bc *Blockchain

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
		bc = New(cfg, log)
		bc.SetStore(store)
		chain, err = NewChain(chainID, store, cfg, log)
		Expect(err).To(BeNil())
		bc.bestChain = chain
		bc.addChain(chain)
	})

	AfterEach(func() {
		db.Close()
	})

	AfterEach(func() {
		Expect(testutil.RemoveTestCfgDir()).To(BeNil())
	})

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

		var block2, block3, block4, block5, block6, block7, block8 *wire.Block
		var key, key2, key3 = crypto.NewKeyFromIntSeed(1), crypto.NewKeyFromIntSeed(2), crypto.NewKeyFromIntSeed(3)
		var account, account2, account3 *wire.Account

		BeforeEach(func() {
			block2, _ = wire.BlockFromString(testdata.ProcessDotGoJSON[1])
			block3, _ = wire.BlockFromString(testdata.ProcessDotGoJSON[2])
			block4, _ = wire.BlockFromString(testdata.ProcessDotGoJSON[3])
			block5, _ = wire.BlockFromString(testdata.ProcessDotGoJSON[4])
			block6, _ = wire.BlockFromString(testdata.ProcessDotGoJSON[5])
			block7, _ = wire.BlockFromString(testdata.ProcessDotGoJSON[6])
			block8, _ = wire.BlockFromString(testdata.ProcessDotGoJSON[7])
		})

		BeforeEach(func() {
			err = chain.init(testdata.TestGenesisBlock)
			Expect(err).To(BeNil())
		})

		BeforeEach(func() {
			account = &wire.Account{Type: wire.AccountTypeBalance, Address: key.Addr(), Balance: "0"}
			err = bc.putAccount(1, chain, account)
			Expect(err).To(BeNil())
			account2 = &wire.Account{Type: wire.AccountTypeBalance, Address: key2.Addr(), Balance: "10"}
			err = bc.putAccount(1, chain, account2)
			Expect(err).To(BeNil())
			account3 = &wire.Account{Type: wire.AccountTypeBalance, Address: key3.Addr(), Balance: "5"}
			err = bc.putAccount(1, chain, account3)
			Expect(err).To(BeNil())
		})

		It("should return error if sender does not exist in the best chain", func() {
			_, err := bc.processTransactions(block2.Transactions, chain)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("failed to get sender's account: account not found"))
		})

		It("should return error if sender account has insufficient value", func() {
			_, err := bc.processTransactions(block3.Transactions, chain)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("insufficient sender account balance"))
		})

		It("should return error if sender value is invalid", func() {
			_, err := bc.processTransactions(block4.Transactions, chain)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("sending amount error: can't convert 100_333 to decimal"))
		})

		Context("recipient does not have an account", func() {

			It(`should return operations and no error; expects 3 ops; 1 OpCreateAccount (for recipient) and 2 OpNewAccountBalance (sender and recipient);
			1st OpNewAccountBalance.Amount = 9; 2nd OpNewAccountBalance.Amount = 1`, func() {
				ops, err := bc.processTransactions(block5.Transactions, chain)
				Expect(err).To(BeNil())
				Expect(ops).To(HaveLen(3))

				Expect(ops[0]).To(BeAssignableToTypeOf(&common.OpCreateAccount{}))
				Expect(ops[0].Address()).To(Equal(block5.Transactions[0].To))

				Expect(ops[1]).To(BeAssignableToTypeOf(&common.OpNewAccountBalance{}))
				Expect(ops[1].Address()).To(Equal(block5.Transactions[0].From))
				Expect(ops[1].(*common.OpNewAccountBalance).Account.Balance).To(Equal("9"))

				Expect(ops[2]).To(BeAssignableToTypeOf(&common.OpNewAccountBalance{}))
				Expect(ops[2].Address()).To(Equal(block5.Transactions[0].To))
				Expect(ops[2].(*common.OpNewAccountBalance).Account.Balance).To(Equal("1"))
			})
		})

		Context("recipient has an account", func() {

			It(`should return operations and no error; expects 2 ops; 2 OpNewAccountBalance (sender and recipient);
				1st OpNewAccountBalance.Amount = 9; 2nd OpNewAccountBalance.Amount = 6`, func() {
				ops, err := bc.processTransactions(block6.Transactions, chain)
				Expect(err).To(BeNil())
				Expect(ops).To(HaveLen(2))

				Expect(ops[0]).To(BeAssignableToTypeOf(&common.OpNewAccountBalance{}))
				Expect(ops[0].Address()).To(Equal(block6.Transactions[0].From))
				Expect(ops[0].(*common.OpNewAccountBalance).Account.Balance).To(Equal("9"))

				Expect(ops[1]).To(BeAssignableToTypeOf(&common.OpNewAccountBalance{}))
				Expect(ops[1].Address()).To(Equal(block6.Transactions[0].To))
				Expect(ops[1].(*common.OpNewAccountBalance).Account.Balance).To(Equal("6"))
			})
		})

		Context("sender has two transactions sending to a recipient", func() {

			Context("recipient has an account", func() {
				It(`should return operations and no error; expects 2 ops; 2 OpNewAccountBalance (sender and recipient);
				1st OpNewAccountBalance.Amount = 8; 2nd OpNewAccountBalance.Amount = 7`, func() {
					ops, err := bc.processTransactions(block7.Transactions, chain)
					Expect(err).To(BeNil())
					Expect(ops).To(HaveLen(2))

					Expect(ops[0]).To(BeAssignableToTypeOf(&common.OpNewAccountBalance{}))
					Expect(ops[0].Address()).To(Equal(block7.Transactions[0].From))
					Expect(ops[0].(*common.OpNewAccountBalance).Account.Balance).To(Equal("8"))

					Expect(ops[1]).To(BeAssignableToTypeOf(&common.OpNewAccountBalance{}))
					Expect(ops[1].Address()).To(Equal(block7.Transactions[0].To))
					Expect(ops[1].(*common.OpNewAccountBalance).Account.Balance).To(Equal("7"))
				})
			})

			Context("recipient does not have an account", func() {
				It(`should return operations and no error; expects 3 ops; 1 OpCreateAccount (for recipient), 2 OpNewAccountBalance (sender and recipient);
				1st OpNewAccountBalance.Amount = 8; 2nd OpNewAccountBalance.Amount = 7`, func() {
					ops, err := bc.processTransactions(block8.Transactions, chain)
					Expect(err).To(BeNil())
					Expect(ops).To(HaveLen(3))

					Expect(ops[0]).To(BeAssignableToTypeOf(&common.OpCreateAccount{}))
					Expect(ops[0].Address()).To(Equal(block8.Transactions[0].To))

					Expect(ops[1]).To(BeAssignableToTypeOf(&common.OpNewAccountBalance{}))
					Expect(ops[1].Address()).To(Equal(block8.Transactions[0].From))
					Expect(ops[1].(*common.OpNewAccountBalance).Account.Balance).To(Equal("8"))

					Expect(ops[2]).To(BeAssignableToTypeOf(&common.OpNewAccountBalance{}))
					Expect(ops[2].Address()).To(Equal(block8.Transactions[0].To))
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
			root := ComputeTxsRoot(txs)
			expected := []uint8{234, 95, 233, 24, 60, 179, 167, 159, 131, 70, 11, 85, 134, 71, 67, 32, 84, 172, 75, 236, 136, 186, 70, 110, 122, 207, 126, 67, 15, 235, 175, 85, 84, 35, 85, 247, 87, 193, 144, 17, 79, 219, 209, 140, 38, 195, 170, 137, 73, 61, 114, 113, 3, 163, 54, 213, 142, 166, 244, 116, 118, 148, 224, 88}
			Expect(root).To(Equal(expected))
		})
	})

	Describe(".MockExecBlock", func() {

		var block *wire.Block
		var sender *crypto.Key

		Context("when block has invalid transaction", func() {

			BeforeEach(func() {
				block, _ = wire.BlockFromString(testdata.ProcessMockBlockData[0])
			})

			It("should return err ", func() {
				_, _, err := bc.mockExecBlock(chain, block)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("failed to process transactions: rejected: failed to get sender's account: account not found"))
			})
		})

		Context("with valid transactions", func() {

			var recipient *crypto.Key

			BeforeEach(func() {
				block, _ = wire.BlockFromString(testdata.ProcessMockBlockData[0])
			})

			BeforeEach(func() {
				sender = crypto.NewKeyFromIntSeed(1)
				recipient = crypto.NewKeyFromIntSeed(2)
				senderAddrKey := common.MakeAccountKey(block.GetNumber(), chain.id, sender.Addr())
				err = chain.store.Put(senderAddrKey, util.ObjectToBytes(&wire.Account{Address: sender.Addr(), Balance: "10"}))
				Expect(err).To(BeNil())
			})

			It("should compute root, state objects without changing the current chain state root", func() {
				block.Transactions[0].SenderPubKey = sender.PubKey().Base58()
				block.Transactions[0].From = sender.Addr()
				block.Transactions[0].To = recipient.Addr()

				curRoot, curRootNode, err := chain.stateTree.Root()
				Expect(err).To(BeNil())
				Expect(curRoot).To(BeNil())
				Expect(curRootNode).To(BeNil())

				root, stateObjs, _ := bc.mockExecBlock(chain, block)
				Expect(stateObjs).To(HaveLen(3))
				Expect(root).ToNot(BeEmpty())
				Expect(root).To(HaveLen(64))

				curRoot, curRootNode, err = chain.stateTree.Root()
				Expect(err).To(BeNil())
				Expect(curRoot).To(BeNil())
				Expect(curRootNode).To(BeNil())
			})
		})

	})

	Describe(".ProcessBlock", func() {

		var block *wire.Block

		It("should reject the block if it has been added to the rejected cache", func() {
			block, _ = wire.BlockFromString(testdata.ProcessDotGoJSON[0])
			bc.rejectedBlocks.Add(block.GetHash(), struct{}{})
			err = bc.ProcessBlock(block)
			Expect(err).ToNot(BeNil())
			Expect(err).To(Equal(common.ErrBlockRejected))
		})

		It("should return error if block already exists in one of the known chains", func() {
			block, _ = wire.BlockFromString(testdata.ProcessDotGoJSON[0])
			err = chain.appendBlock(block)
			Expect(err).To(BeNil())
			err = bc.ProcessBlock(block)
			Expect(err).ToNot(BeNil())
			Expect(err).To(Equal(common.ErrBlockExists))
		})

		It("should return error if block has been added to the orphaned cache", func() {
			block, _ = wire.BlockFromString(testdata.ProcessDotGoJSON[0])
			bc.orphanBlocks.Add(block.GetHash(), block)
			err = bc.ProcessBlock(block)
			Expect(err).ToNot(BeNil())
			Expect(err).To(Equal(common.ErrOrphanBlock))
		})

		When("a block's parent does not exist in any chain", func() {
			It("should return nil and add be added to the orphan block cache", func() {
				block, _ = wire.BlockFromString(testdata.ProcessDotGoJSON[0])
				err = bc.ProcessBlock(block)
				Expect(err).To(BeNil())
				Expect(bc.orphanBlocks.Has(block.GetHash())).To(BeTrue())
			})
		})

		Describe("how stale blocks are handled", func() {
			When("a block's parent exists in a chain", func() {
				var genesis, block2, chainTip, veryStaleBlock, staleBlock, futureNumberedBlock *wire.Block

				BeforeEach(func() {
					genesis, _ = wire.BlockFromString(testdata.ProcessStaleOrInvalidBlockData[0])
					block2, _ = wire.BlockFromString(testdata.ProcessStaleOrInvalidBlockData[1])
					chainTip, _ = wire.BlockFromString(testdata.ProcessStaleOrInvalidBlockData[2])
					veryStaleBlock, _ = wire.BlockFromString(testdata.ProcessStaleOrInvalidBlockData[3])
					staleBlock, _ = wire.BlockFromString(testdata.ProcessStaleOrInvalidBlockData[4])
					futureNumberedBlock, _ = wire.BlockFromString(testdata.ProcessStaleOrInvalidBlockData[5])
					err = chain.appendBlock(genesis)
					Expect(err).To(BeNil())
					err = chain.appendBlock(block2)
					Expect(err).To(BeNil())
					err = chain.appendBlock(chainTip)
					Expect(err).To(BeNil())
				})

				It("should return ErrVeryStaleBlock when block number less than the latest block", func() {
					err = bc.ProcessBlock(veryStaleBlock)
					Expect(err).ToNot(BeNil())
					Expect(err).To(Equal(common.ErrVeryStaleBlock))
				})

				When("stale block has same number as the chainTip", func() {

					BeforeEach(func() {
						Expect(bc.chains).To(HaveLen(1))
					})

					It("should create a new chain tree; return nil; new tree's parent should be expected; tree must include the new block", func() {
						err = bc.ProcessBlock(staleBlock)
						Expect(err).To(BeNil())
						Expect(bc.chains).To(HaveLen(2))
						Expect(bc.chains[1].parentBlock.Hash).To(Equal(block2.Hash))
						hasBlock, err := bc.chains[1].hasBlock(staleBlock.GetHash())
						Expect(err).To(BeNil())
						Expect(hasBlock).To(BeTrue())
					})
				})

				It("should return error when the difference between block number and chain  block number is greater than 1", func() {
					err = bc.ProcessBlock(futureNumberedBlock)
					Expect(err).ToNot(BeNil())
					Expect(err).To(Equal(common.ErrBlockFailedValidation))
				})
			})
		})

		Context("state root comparison", func() {

			var genesis *wire.Block
			var recipient, sender *crypto.Key

			BeforeEach(func() {
				genesis, _ = wire.BlockFromString(testdata.ProcessStateRootData[1])
				block, _ = wire.BlockFromString(testdata.ProcessStateRootData[2])
				chain.appendBlock(genesis)
			})

			BeforeEach(func() {
				sender = crypto.NewKeyFromIntSeed(1)
				recipient = crypto.NewKeyFromIntSeed(2)
				senderAddrKey := common.MakeAccountKey(block.GetNumber(), chain.id, sender.Addr())
				err = chain.store.Put(senderAddrKey, util.ObjectToBytes(&wire.Account{Address: sender.Addr(), Balance: "10"}))
				Expect(err).To(BeNil())
			})

			It("should return error when block state root does not match", func() {
				block.Transactions[0].SenderPubKey = sender.PubKey().Base58()
				block.Transactions[0].From = sender.Addr()
				block.Transactions[0].To = recipient.Addr()
				err = bc.ProcessBlock(block)
				Expect(err).ToNot(BeNil())
				Expect(err).To(Equal(common.ErrBlockStateRootInvalid))
			})

			It("should successfully accept state root of block", func() {
				block.Transactions[0].SenderPubKey = sender.PubKey().Base58()
				block.Transactions[0].From = sender.Addr()
				block.Transactions[0].To = recipient.Addr()

				root, stateObjs, err := bc.mockExecBlock(chain, block)
				Expect(err).To(BeNil())
				block.Header.StateRoot = util.ToHex(root)

				err = bc.ProcessBlock(block)
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
			})
		})
	})

	Describe(".processOrphanBlocks", func() {

		var parent1, orphanParent, orphan, orphan2 *wire.Block

		BeforeEach(func() {
			parent1, _ = wire.BlockFromString(testdata.ProcessOrphanBlockData[0])
			orphanParent, _ = wire.BlockFromString(testdata.ProcessOrphanBlockData[1])
			orphan, _ = wire.BlockFromString(testdata.ProcessOrphanBlockData[2])
			orphan2, _ = wire.BlockFromString(testdata.ProcessOrphanBlockData[3])
			chain.appendBlock(parent1)
		})

		Context("with one orphan block", func() {

			BeforeEach(func() {
				err = bc.ProcessBlock(orphan)
				Expect(err).To(BeNil())
				Expect(bc.orphanBlocks.Len()).To(Equal(1))
			})

			BeforeEach(func() {
				err = chain.appendBlock(orphanParent)
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
				err = bc.ProcessBlock(orphan)
				Expect(err).To(BeNil())
				err = bc.ProcessBlock(orphan2)
				Expect(err).To(BeNil())
				Expect(bc.orphanBlocks.Len()).To(Equal(2))
			})

			// add the parent that is linked by one of the orphan
			BeforeEach(func() {
				err = chain.appendBlock(orphanParent)
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
					root, rn, err := chain.stateTree.Root()
					fmt.Println(root, rn, err)
				})
			})
		})
	})
})
