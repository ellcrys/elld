package blockchain

import (
	"math/big"
	"time"

	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/wire"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var BlockchainTest = func() bool {

	return Describe("Blockchain", func() {

		// Describe(".SetStore", func() {
		// 	It("should set store", func() {
		// 		bc := New(nil, cfg, log)
		// 		bc.SetDB(db)
		// 	})
		// })

		// Describe(".addChain", func() {
		// 	It("should add chain", func() {
		// 		chain := NewChain("chain_id", db, cfg, log)
		// 		Expect(err).To(BeNil())
		// 		Expect(bc.chains).To(HaveLen(1))
		// 		err = bc.addChain(chain)
		// 		Expect(err).To(BeNil())
		// 		Expect(bc.chains).To(HaveLen(2))
		// 	})
		// })

		// Describe(".removeChain", func() {

		// 	var chain *Chain

		// 	BeforeEach(func() {
		// 		chain = NewChain("chain_id", db, cfg, log)
		// 		Expect(err).To(BeNil())
		// 		err = bc.addChain(chain)
		// 		Expect(err).To(BeNil())
		// 		Expect(bc.chains).To(HaveLen(2))
		// 	})

		// 	It("should remove chain", func() {
		// 		bc.removeChain(chain)
		// 		Expect(bc.chains).To(HaveLen(1))
		// 		Expect(bc.chains[chain.GetID()]).To(BeNil())
		// 	})
		// })

		// Describe(".hasChain", func() {

		// 	var chain *Chain

		// 	BeforeEach(func() {
		// 		chain = NewChain("chain_id", db, cfg, log)
		// 		Expect(err).To(BeNil())
		// 	})

		// 	It("should return true if chain exists", func() {
		// 		Expect(bc.hasChain(chain)).To(BeFalse())
		// 		err = bc.addChain(chain)
		// 		Expect(err).To(BeNil())
		// 		Expect(bc.hasChain(chain)).To(BeTrue())
		// 	})
		// })

		Describe(".Up", func() {

			var block2 core.Block

			BeforeEach(func() {
				chain := NewChain("c1", db, cfg, log)
				block2 = MakeTestBlock(bc, chain, &core.GenerateBlockParams{
					Transactions: []core.Transaction{
						wire.NewTx(wire.TxTypeAlloc, 123, util.String(sender.Addr()), sender, "1", "0.1", 1532730722),
					},
					Creator:    sender,
					Nonce:      core.EncodeNonce(1),
					Difficulty: new(big.Int).SetInt64(131072),
				})
			})

			It("should return error if store is not set", func() {
				bc := New(nil, cfg, log)
				err = bc.Up()
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("db has not been initialized"))
			})

			When("genesis block is not valid: block number is > 1", func() {
				BeforeEach(func() {
					bc.chains = make(map[util.String]*Chain)
					chain := NewChain("c1", db, cfg, log)
					block := MakeTestBlock(bc, chain, &core.GenerateBlockParams{
						Transactions: []core.Transaction{
							wire.NewTx(wire.TxTypeAlloc, 123, util.String(sender.Addr()), sender, "1", "0.1", 1532730722),
						},
						Creator:    sender,
						Nonce:      core.EncodeNonce(1),
						Difficulty: new(big.Int).SetInt64(131072),
					})
					block.GetHeader().SetNumber(2)
					GenesisBlock = block
				})

				It("should return error if genesis block number is not equal to 1", func() {
					err = bc.Up()
					Expect(err).ToNot(BeNil())
					Expect(err.Error()).To(Equal("genesis block error: expected block number 1"))
				})
			})

			When("genesis block is not valid: sender account does not exist", func() {
				BeforeEach(func() {
					bc.chains = make(map[util.String]*Chain)
					chain := NewChain("c1", db, cfg, log)

					Expect(bc.putAccount(1, chain, &wire.Account{
						Type:    wire.AccountTypeBalance,
						Address: util.String(sender.Addr()),
						Balance: "1000",
					})).To(BeNil())

					block := MakeTestBlock(bc, chain, &core.GenerateBlockParams{
						Transactions: []core.Transaction{
							wire.NewTx(wire.TxTypeBalance, 123, util.String(sender.Addr()), sender, "1", "0.1", 1532730722),
						},
						Creator:    sender,
						Nonce:      core.EncodeNonce(1),
						Difficulty: new(big.Int).SetInt64(131072),
					})
					block.GetTransactions()[0].SetFrom("unknown_account")
					block.SetHash(block.ComputeHash())
					blockSig, _ := wire.BlockSign(block, sender.PrivKey().Base58())
					block.SetSignature(blockSig)
					GenesisBlock = block
				})

				It("should return error if a transaction's sender does not exists", func() {
					err = bc.Up()
					Expect(err).ToNot(BeNil())
					Expect(err.Error()).To(Equal("genesis block error: transaction error: index{0}: failed to get sender's account: account not found"))
				})
			})

			When("genesis block is valid", func() {

				BeforeEach(func() {
					bc.chains = make(map[util.String]*Chain)
					chain := NewChain("c1", db, cfg, log)
					block := MakeTestBlock(bc, chain, &core.GenerateBlockParams{
						Transactions: []core.Transaction{
							wire.NewTx(wire.TxTypeAlloc, 123, util.String(sender.Addr()), sender, "1", "0.1", 1532730722),
						},
						Creator:    sender,
						Nonce:      core.EncodeNonce(1),
						Difficulty: new(big.Int).SetInt64(131072),
					})
					GenesisBlock = block
				})

				It("should assign new chain as the best chain if no chain is known", func() {
					err = bc.Up()
					Expect(err).To(BeNil())
					Expect(bc.bestChain).ToNot(BeNil())
				})
			})

			It("should load all chains", func() {
				GenesisBlock = block2

				c1 := NewChain("c_1", db, cfg, log)
				c2 := NewChain("c_2", db, cfg, log)

				err = bc.saveChain(c1, "", 0)
				Expect(err).To(BeNil())

				err = bc.saveChain(c2, "", 0)
				Expect(err).To(BeNil())

				err = c1.append(GenesisBlock)
				Expect(err).To(BeNil())

				err = bc.Up()
				Expect(err).To(BeNil())

				Expect(bc.chains).To(HaveLen(3))
				Expect(bc.chains["c_1"].id).To(Equal(c1.id))
				Expect(bc.chains["c_2"].id).To(Equal(c2.id))
				Expect(bc.bestChain.id).To(Equal(genesisChain.id))
			})
		})

		Context("Metadata", func() {

			var meta = core.BlockchainMeta{}

			Describe(".UpdateMeta", func() {
				It("should successfully save metadata", func() {
					err = bc.updateMeta(&meta)
					Expect(err).To(BeNil())
				})
			})

			Describe(".GetMeta", func() {

				BeforeEach(func() {
					err = bc.updateMeta(&meta)
					Expect(err).To(BeNil())
				})

				It("should return metadata", func() {
					result := bc.GetMeta()
					Expect(result).To(Equal(&meta))
				})
			})
		})

		Describe(".HybridMode", func() {

			BeforeEach(func() {
				bc.bestChain = genesisChain
			})

			It("should return false if we have not reached hybrid mode block height", func() {
				reached, err := bc.HybridMode()
				Expect(err).To(BeNil())
				Expect(reached).To(BeFalse())
			})

			It("should return true if we have reached hybrid mode block height", func() {
				cfg.Chain.TargetHybridModeBlock = 1 // set to 1 (genesis block height)
				reached, err := bc.HybridMode()
				Expect(err).To(BeNil())
				Expect(reached).To(BeTrue())
			})
		})

		Describe(".findChainByLastBlockHash", func() {

			var b1 core.Block
			var chain2 *Chain

			BeforeEach(func() {
				chain2 = NewChain("chain2", db, cfg, log)
				err = bc.addChain(chain2)
				Expect(err).To(BeNil())

				Expect(bc.putAccount(1, chain2, &wire.Account{
					Type:    wire.AccountTypeBalance,
					Address: util.String(sender.Addr()),
					Balance: "1000",
				})).To(BeNil())

				b1 = MakeTestBlock(bc, chain2, &core.GenerateBlockParams{
					Transactions: []core.Transaction{
						wire.NewTx(wire.TxTypeBalance, 123, util.String(receiver.Addr()), sender, "1", "0.1", 1532730723),
					},
					Creator:    sender,
					Nonce:      core.EncodeNonce(1),
					Difficulty: new(big.Int).SetInt64(131072),
				})

				Expect(err).To(BeNil())
				err = chain2.append(b1)
			})

			It("should return chain=nil, header=nil, err=nil if no block on the chain matches the hash", func() {
				_block, chain, header, err := bc.findChainByBlockHash(util.Hash{1, 2, 3})
				Expect(err).To(BeNil())
				Expect(_block).To(BeNil())
				Expect(header).To(BeNil())
				Expect(chain).To(BeNil())
			})

			Context("when the hash belongs to the highest block in chain2", func() {
				It("should return chain2 and header must match the header of the recently added block", func() {
					_block, chain, header, err := bc.findChainByBlockHash(b1.GetHash())
					Expect(err).To(BeNil())
					Expect(b1.Bytes()).To(Equal(_block.Bytes()))
					Expect(chain.GetID()).To(Equal(chain2.id))
					Expect(header.ComputeHash()).To(Equal(b1.GetHeader().ComputeHash()))
				})
			})

			Context("when the hash belongs to a block that is not the highest block", func() {

				var b2 core.Block

				BeforeEach(func() {
					b2 = MakeTestBlock(bc, genesisChain, &core.GenerateBlockParams{
						Transactions: []core.Transaction{
							wire.NewTx(wire.TxTypeBalance, 123, util.String(receiver.Addr()), sender, "1", "0.1", 1532730724),
						},
						Creator:    sender,
						Nonce:      core.EncodeNonce(1),
						Difficulty: new(big.Int).SetInt64(131072),
					})
					err = genesisChain.append(b2)
					Expect(err).To(BeNil())
				})

				It("should return chain and header matching the header of block 1", func() {
					_block, chain, tipHeader, err := bc.findChainByBlockHash(genesisBlock.GetHash())
					Expect(err).To(BeNil())
					Expect(genesisBlock.Bytes()).To(Equal(_block.Bytes()))
					Expect(genesisBlock.GetNumber()).To(Equal(uint64(1)))
					Expect(chain.GetID()).To(Equal(chain.id))
					Expect(tipHeader.ComputeHash()).To(Equal(b2.GetHeader().ComputeHash()))
				})
			})
		})

		Describe(".loadChain", func() {
			var block core.Block
			var chain, subChain *Chain

			BeforeEach(func() {
				bc.chains = make(map[util.String]*Chain)
			})

			BeforeEach(func() {
				chain = NewChain("chain_2", db, cfg, log)
				err := bc.saveChain(chain, "", 0)
				Expect(err).To(BeNil())

				subChain = NewChain("sub_chain", db, cfg, log)

				block = MakeTestBlock(bc, genesisChain, &core.GenerateBlockParams{
					Transactions: []core.Transaction{
						wire.NewTx(wire.TxTypeBalance, 123, util.String(receiver.Addr()), sender, "1", "0.1", 1532730724),
					},
					Creator:    sender,
					Nonce:      core.EncodeNonce(1),
					Difficulty: new(big.Int).SetInt64(131072),
				})

				err = chain.append(block)
				Expect(err).To(BeNil())

				err = bc.saveChain(subChain, chain.GetID(), block.GetNumber())
				Expect(err).To(BeNil())
			})

			It("should return error when only ParentBlockNumber is set but ParentChainID is unset", func() {
				err = bc.loadChain(&core.ChainInfo{ID: "some_id", ParentBlockNumber: 1})
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("chain load failed: parent chain id and parent block id are both required"))
			})

			It("should return error when only ParentChainID is set but ParentBlockNumber is unset", func() {
				err = bc.loadChain(&core.ChainInfo{ID: chain.GetID(), ParentChainID: "some_id"})
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("chain load failed: parent chain id and parent block id are both required"))
			})

			It("should return error when parent block does not exist", func() {
				err = bc.loadChain(&core.ChainInfo{ID: chain.GetID(), ParentChainID: "some_id", ParentBlockNumber: 100})
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("chain load failed: parent block {100} of chain {chain_2} not found"))
			})

			It("should successfully load chain with no parent into the cache", func() {
				err = bc.loadChain(&core.ChainInfo{ID: chain.GetID()})
				Expect(err).To(BeNil())
				Expect(bc.chains).To(HaveLen(2))
				Expect(bc.chains[chain.GetID()]).ToNot(BeNil())
			})

			It("should successfully load chain into the cache with parent block and chain info set", func() {
				err = bc.loadChain(&core.ChainInfo{ID: subChain.GetID(), ParentChainID: chain.GetID(), ParentBlockNumber: block.GetNumber()})
				Expect(err).To(BeNil())
				Expect(bc.chains).To(HaveLen(2))
				Expect(bc.chains[subChain.GetID()]).ToNot(BeNil())
			})
		})

		Describe(".findChainInfo", func() {

			var chain *Chain

			BeforeEach(func() {
				chain = NewChain("chain_a", db, cfg, log)
				err := bc.saveChain(chain, "", 0)
				Expect(err).To(BeNil())
			})

			It("should find chain with id = chain_a", func() {
				chInfo, err := bc.findChainInfo("chain_a")
				Expect(err).To(BeNil())
				Expect(chInfo.ID).To(Equal(chain.GetID()))
				Expect(chInfo.Timestamp).ToNot(Equal(0))
			})

			It("should return err = 'chain not found' if chain does not exist", func() {
				_, err := bc.findChainInfo("chain_b")
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("chain not found"))
			})
		})

		Describe(".newChain", func() {

			var parentBlock, block, unknownParent core.Block

			BeforeEach(func() {
				parentBlock = MakeTestBlock(bc, genesisChain, &core.GenerateBlockParams{
					Transactions: []core.Transaction{
						wire.NewTx(wire.TxTypeBalance, 123, util.String(receiver.Addr()), sender, "1", "0.1", 1532730724),
					},
					Creator:    sender,
					Nonce:      core.EncodeNonce(1),
					Difficulty: new(big.Int).SetInt64(131072),
				})

				block = MakeTestBlock(bc, genesisChain, &core.GenerateBlockParams{
					Transactions: []core.Transaction{
						wire.NewTx(wire.TxTypeBalance, 123, util.String(receiver.Addr()), sender, "1", "0.1", 1532730724),
					},
					OverrideParentHash: parentBlock.GetHash(),
					Creator:            sender,
					Nonce:              core.EncodeNonce(1),
					Difficulty:         new(big.Int).SetInt64(131072),
					OverrideTimestamp:  time.Now().Add(2 * time.Second).Unix(),
				})

				unknownParent = MakeTestBlock(bc, genesisChain, &core.GenerateBlockParams{
					Transactions: []core.Transaction{
						wire.NewTx(wire.TxTypeBalance, 123, util.String(receiver.Addr()), sender, "1", "0.1", 1532730724),
					},
					Creator:            sender,
					OverrideParentHash: util.StrToHash("unknown"),
					Nonce:              core.EncodeNonce(1),
					Difficulty:         new(big.Int).SetInt64(131072),
					OverrideTimestamp:  time.Now().Add(3 * time.Second).Unix(),
				})
			})

			It("should return error if block is nil", func() {
				_, err := bc.newChain(nil, nil, nil, nil)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("initial block cannot be nil"))
			})

			It("should return error if initial block parent is nil", func() {
				_, err := bc.newChain(nil, block, nil, nil)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("initial block parent cannot be nil"))
			})

			It("should return error if initial block and parent are not related", func() {
				_, err = bc.newChain(nil, block, unknownParent, nil)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("initial block and parent are not related"))
			})

			It("should return error if parent chain is nil", func() {
				_, err = bc.newChain(nil, block, parentBlock, nil)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("parent chain cannot be nil"))
			})

			It("should successfully return a new chain", func() {
				chain, err := bc.newChain(nil, block, parentBlock, genesisChain)
				Expect(err).To(BeNil())
				Expect(chain).ToNot(BeNil())
				Expect(chain.parentBlock).To(Equal(parentBlock))
			})
		})

		Describe(".GetTransaction", func() {
			var block core.Block
			var chain *Chain

			BeforeEach(func() {
				chain = NewChain("chain_a", db, cfg, log)

				block = MakeTestBlock(bc, genesisChain, &core.GenerateBlockParams{
					Transactions: []core.Transaction{
						wire.NewTx(wire.TxTypeBalance, 123, util.String(receiver.Addr()), sender, "1", "0.1", 1532730724),
					},
					Creator:    sender,
					Nonce:      core.EncodeNonce(1),
					Difficulty: new(big.Int).SetInt64(131072),
				})

				chain.append(block)
				err = chain.PutTransactions(block.GetTransactions(), block.GetNumber())
				Expect(err).To(BeNil())
			})

			It("should return err = 'best chain unknown' if the best chain has not been decided", func() {
				bc.bestChain = nil
				_, err := bc.GetTransaction(block.GetTransactions()[0].GetHash())
				Expect(err).ToNot(BeNil())
				Expect(err).To(Equal(core.ErrBestChainUnknown))
			})

			It("should return transaction and no error", func() {
				bc.bestChain = chain
				tx, err := bc.GetTransaction(block.GetTransactions()[0].GetHash())
				Expect(err).To(BeNil())
				Expect(tx).To(Equal(block.GetTransactions()[0]))
			})

			It("should return err = 'transaction not found' when main chain does not have the transaction", func() {
				bc.bestChain = chain
				tx, err := bc.GetTransaction(util.Hash{1, 2, 3})
				Expect(err).ToNot(BeNil())
				Expect(err).To(Equal(core.ErrTxNotFound))
				Expect(tx).To(BeNil())
			})
		})

		Describe(".chooseBestChain", func() {

			var chainA, chainB *Chain

			BeforeEach(func() {
				genesisChainBlock2 := MakeTestBlock(bc, genesisChain, &core.GenerateBlockParams{
					Transactions: []core.Transaction{
						wire.NewTx(wire.TxTypeBalance, 123, util.String(receiver.Addr()), sender, "1", "0.1", 1532730724),
					},
					Creator:                 sender,
					Nonce:                   core.EncodeNonce(1),
					Difficulty:              new(big.Int).SetInt64(1),
					OverrideTotalDifficulty: new(big.Int).SetInt64(10),
				})
				err := genesisChain.append(genesisChainBlock2)
				Expect(err).To(BeNil())
			})

			Context("test difficulty rule", func() {

				When("chainA has the most total difficulty", func() {

					BeforeEach(func() {
						chainA = NewChain("chain_a", db, cfg, log)
						err := bc.saveChain(chainA, "", 0)
						Expect(err).To(BeNil())

						chainABlock1 := MakeTestBlock(bc, genesisChain, &core.GenerateBlockParams{
							Transactions: []core.Transaction{
								wire.NewTx(wire.TxTypeAlloc, 123, util.String(sender.Addr()), sender, "1", "0.1", 1532730724),
							},
							Creator:                 sender,
							Nonce:                   core.EncodeNonce(1),
							Difficulty:              new(big.Int).SetInt64(1),
							OverrideTotalDifficulty: new(big.Int).SetInt64(100),
						})

						err = chainA.append(chainABlock1)
						Expect(err).To(BeNil())
					})

					It("should return chainA as the best chain since it has a higher total difficulty than the genesis chain", func() {
						bc.bestChain = nil
						Expect(bc.chains).To(HaveLen(2))
						bestChain, err := bc.chooseBestChain()
						Expect(err).To(BeNil())
						Expect(bestChain.id).To(Equal(chainA.id))
					})
				})

				When("chainB has the lowest total difficulty", func() {
					BeforeEach(func() {
						chainB = NewChain("chain_b", db, cfg, log)
						err := bc.saveChain(chainB, "", 0)
						Expect(err).To(BeNil())

						chainBBlock1 := MakeTestBlock(bc, genesisChain, &core.GenerateBlockParams{
							Transactions: []core.Transaction{
								wire.NewTx(wire.TxTypeAlloc, 123, util.String(sender.Addr()), sender, "1", "0.1", 1532730726),
							},
							Creator:                 sender,
							Nonce:                   core.EncodeNonce(1),
							Difficulty:              new(big.Int).SetInt64(1),
							OverrideTotalDifficulty: new(big.Int).SetInt64(5),
						})

						err = chainB.append(chainBBlock1)
						Expect(err).To(BeNil())
					})

					It("should return genesis chain as the best chain since it has a higher total difficulty than chainB", func() {
						bc.bestChain = nil
						Expect(bc.chains).To(HaveLen(2))
						bestChain, err := bc.chooseBestChain()
						Expect(err).To(BeNil())
						Expect(bestChain.id).To(Equal(genesisChain.id))
					})
				})
			})

			Context("test oldest chain rule", func() {

				When("chainA and genesis chain have the same total difficulty but the genesis chain is older", func() {

					BeforeEach(func() {
						chainA = NewChain("chain_a", db, cfg, log)
						err := bc.saveChain(chainA, "", 0)
						Expect(err).To(BeNil())

						chainABlock1 := MakeTestBlock(bc, genesisChain, &core.GenerateBlockParams{
							Transactions: []core.Transaction{
								wire.NewTx(wire.TxTypeAlloc, 123, util.String(sender.Addr()), sender, "1", "0.1", 1532730724),
							},
							Creator:                 sender,
							Nonce:                   core.EncodeNonce(1),
							Difficulty:              new(big.Int).SetInt64(1),
							OverrideTotalDifficulty: new(big.Int).SetInt64(10),
						})

						err = chainA.append(chainABlock1)
						Expect(err).To(BeNil())
					})

					It("should return genesis chain as the best chain since it has an older chain timestamp", func() {
						bc.bestChain = nil
						Expect(bc.chains).To(HaveLen(2))
						bestChain, err := bc.chooseBestChain()
						Expect(err).To(BeNil())
						Expect(bestChain.id).To(Equal(genesisChain.id))
					})
				})

			})

			Context("test largest point address rule", func() {
				When("chainA and genesis chain have the same total difficulty and chain age", func() {

					BeforeEach(func() {
						chainA = NewChain("chain_a", db, cfg, log)
						chainA.info.Timestamp = genesisChain.info.Timestamp
						err := bc.saveChain(chainA, "", 0)
						Expect(err).To(BeNil())

						chainABlock1 := MakeTestBlock(bc, genesisChain, &core.GenerateBlockParams{
							Transactions: []core.Transaction{
								wire.NewTx(wire.TxTypeAlloc, 123, util.String(sender.Addr()), sender, "1", "0.1", 1532730724),
							},
							Creator:                 sender,
							Nonce:                   core.EncodeNonce(1),
							Difficulty:              new(big.Int).SetInt64(1),
							OverrideTotalDifficulty: new(big.Int).SetInt64(10),
						})

						err = chainA.append(chainABlock1)
						Expect(err).To(BeNil())
					})

					It("should return the chain with the largest pointer address", func() {
						bc.bestChain = nil
						Expect(bc.chains).To(HaveLen(2))
						bestChain, err := bc.chooseBestChain()
						Expect(err).To(BeNil())
						delete(bc.chains, bestChain.id)
						for _, leastChain := range bc.chains {
							Expect(util.GetPtrAddr(leastChain).Cmp(util.GetPtrAddr(bestChain))).To(Equal(-1))
						}
					})
				})
			})
		})

		Describe(".reOrg", func() {
			// var mainChainBlocks []core.Block
			// var forkBlock core.Block

			BeforeEach(func() {

				// forkBlock = MakeTestBlock(bc, genesisChain, &core.GenerateBlockParams{
				// 	Transactions: []core.Transaction{
				// 		wire.NewTx(wire.TxTypeBalance, 191, util.String(receiver.Addr()), sender, "1", "0.1", 1532730723),
				// 	},
				// 	Creator:    sender,
				// 	Nonce:      core.EncodeNonce(1),
				// 	Difficulty: new(big.Int).SetInt64(131072),
				// })

				// forkBlock2 := MakeTestBlock(bc, genesisChain, &core.GenerateBlockParams{
				// 	Transactions: []core.Transaction{
				// 		wire.NewTx(wire.TxTypeBalance, 191, util.String(receiver.Addr()), sender, "1", "0.1", 1532730724),
				// 	},
				// 	Creator:    sender,
				// 	Nonce:      core.EncodeNonce(1),
				// 	Difficulty: new(big.Int).SetInt64(131072),
				// })

				// _, err = bc.ProcessBlock(forkBlock)
				// Expect(err).To(BeNil())

				// _, err = bc.ProcessBlock(forkBlock2)
				// Expect(err).To(BeNil())

				// for i := 0; i < 2; i++ {
				// 	mainChainBlocks = append(mainChainBlocks, MakeTestBlock(bc, genesisChain, &core.GenerateBlockParams{
				// 		Transactions: []core.Transaction{
				// 			wire.NewTx(wire.TxTypeBalance, 123+int64(i), util.String(receiver.Addr()), sender, "1", "0.1", 1532730724),
				// 		},
				// 		Creator:    sender,
				// 		Nonce:      core.EncodeNonce(1),
				// 		Difficulty: new(big.Int).SetInt64(131072),
				// 	}))
				// 	_, err = bc.ProcessBlock(mainChainBlocks[i])
				// 	Expect(err).To(BeNil())
				// }
			})

			BeforeEach(func() {
				// sdBlock2 := MakeTestBlock(bc, genesisChain, &core.GenerateBlockParams{
				// 	Transactions: []core.Transaction{
				// 		wire.NewTx(wire.TxTypeAlloc, 127, util.String(sender.Addr()), sender, "1", "0.1", 1532730727),
				// 	},
				// 	OverrideChainTip:   1,
				// 	OverrideParentHash: mainChainBlocks[0].GetHash(),
				// 	Creator:            sender,
				// 	Nonce:              core.EncodeNonce(1),
				// 	Difficulty:         new(big.Int).SetInt64(131072),
				// })

				// _, err = bc.ProcessBlock(forkBlock)
				// Expect(err).To(BeNil())
			})

			It("should return error if side chain does not have a parent block set", func() {
				// c, _ := bc.bestChain.Current()
				// fmt.Println(c.GetNumber())
				// err := bc.reOrg(forkedBlockchain.bestChain)
				// Expect(err).ToNot(BeNil())
				// Expect(err.Error()).To(Equal("parent block not set on sidechain"))
			})

			It("should be successful", func() {
				// forkedBlockchain.bestChain.parentBlock = genesisBlock
				// err := bc.reOrg(forkedBlockchain.bestChain)
				// Expect(err).To(BeNil())

				// remove blocks that aren't in the forked chain
			})
		})
	})

}
