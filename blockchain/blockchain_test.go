package blockchain

import (
	"github.com/ellcrys/elld/blockchain/common"
	"github.com/ellcrys/elld/blockchain/testdata"
	"github.com/ellcrys/elld/wire"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var BlockchainTest = func() bool {

	return Describe("Blockchain", func() {

		Describe(".SetStore", func() {
			It("should set store", func() {
				bc := New(nil, cfg, log)
				bc.SetStore(testStore)
				Expect(bc.store).ToNot(BeNil())
			})
		})

		Describe(".addChain", func() {
			It("should add chain", func() {
				chain := NewChain("chain_id", testStore, cfg, log)
				Expect(err).To(BeNil())
				Expect(bc.chains).To(HaveLen(1))
				err = bc.addChain(chain)
				Expect(err).To(BeNil())
				Expect(bc.chains).To(HaveLen(2))
			})
		})

		Describe(".removeChain", func() {

			var chain *Chain

			BeforeEach(func() {
				chain = NewChain("chain_id", testStore, cfg, log)
				Expect(err).To(BeNil())
				err = bc.addChain(chain)
				Expect(err).To(BeNil())
				Expect(bc.chains).To(HaveLen(2))
			})

			It("should remove chain", func() {
				bc.removeChain(chain)
				Expect(bc.chains).To(HaveLen(1))
				Expect(bc.chains[chain.GetID()]).To(BeNil())
			})
		})

		Describe(".hasChain", func() {

			var chain *Chain

			BeforeEach(func() {
				chain = NewChain("chain_id", testStore, cfg, log)
				Expect(err).To(BeNil())
			})

			It("should return true if chain exists", func() {
				Expect(bc.hasChain(chain)).To(BeFalse())
				err = bc.addChain(chain)
				Expect(err).To(BeNil())
				Expect(bc.hasChain(chain)).To(BeTrue())
			})
		})

		Describe(".Up", func() {

			BeforeEach(func() {
				bc.chains = make(map[string]*Chain)
				GenesisBlock = testdata.BlockWithSingleAllocTx
			})

			// BeforeEach(func() {
			// 	blk, err := bc.GenerateBlock(&GenerateBlockParams{
			// 		Transactions: []*wire.Transaction{
			// 			wire.NewTx(wire.TxTypeAllocCoin, 123, sender.Addr(), sender, "1", "0.1", 1532730722),
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

			It("should return error if store is not set", func() {
				bc := New(nil, cfg, log)
				err = bc.Up()
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("store has not been initialized"))
			})

			When("genesis block is not valid", func() {
				It("should return error if genesis block number is not equal to 1", func() {
					GenesisBlock = testdata.GenesisBlockWithAllocTxAndNumber2
					err = bc.Up()
					Expect(err).ToNot(BeNil())
					Expect(err.Error()).To(Equal("genesis block error: expected block number 1"))
				})

				It("should return error if a transaction's sender does not exists", func() {
					GenesisBlock = testdata.GenesisBlockSenderNotFound
					err = bc.Up()
					Expect(err).ToNot(BeNil())
					Expect(err.Error()).To(Equal("genesis block error: transaction error: index{0}: failed to get sender's account: account not found"))
				})
			})

			When("genesis block is valid", func() {
				It("should assign new chain as the best chain if no chain is known", func() {
					GenesisBlock = testdata.BlockWithSingleAllocTx
					err = bc.Up()
					Expect(err).To(BeNil())
					Expect(bc.bestChain).ToNot(BeNil())
				})
			})

			It("should load all chains", func() {
				c1 := NewChain("c1", testStore, cfg, log)
				c2 := NewChain("c2", testStore, cfg, log)

				err = bc.saveChain(c1, "", 0)
				Expect(err).To(BeNil())

				err = bc.saveChain(c2, "", 0)
				Expect(err).To(BeNil())

				err = bc.Up()
				Expect(err).To(BeNil())
				Expect(bc.chains).To(HaveLen(2))
				Expect(bc.chains["c1"].id).To(Equal(c1.id))
				Expect(bc.chains["c2"].id).To(Equal(c2.id))
			})
		})

		Context("Metadata", func() {

			var meta = common.BlockchainMeta{}

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
				bc.bestChain = chain
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

			var block *wire.Block
			var chain2 *Chain

			BeforeEach(func() {
				block = testdata.BlockSet1[0]

				chain2 = NewChain("chain2", testStore, cfg, log)
				Expect(err).To(BeNil())
				err = chain2.append(block)
				Expect(err).To(BeNil())

				err = bc.addChain(chain2)
				Expect(err).To(BeNil())

				err = chain2.store.PutBlock(chain2.id, block)
				Expect(err).To(BeNil())
			})

			It("should return chain=nil, header=nil, err=nil if no block on the chain matches the hash", func() {
				_block, chain, header, err := bc.findBlockChainByHash("unknown")
				Expect(err).To(BeNil())
				Expect(_block).To(BeNil())
				Expect(header).To(BeNil())
				Expect(chain).To(BeNil())
			})

			Context("when the hash maps to the highest block in chain2", func() {
				It("should return chain2 and header matching the header of the recently added block", func() {
					_block, chain, header, err := bc.findBlockChainByHash(block.GetHash().HexStr())
					Expect(err).To(BeNil())
					Expect(block).To(Equal(_block))
					Expect(chain.GetID()).To(Equal(chain2.id))
					Expect(header.ComputeHash()).To(Equal(block.Header.ComputeHash()))
				})
			})

			Context("when the hash maps to a block that is not the highest block", func() {

				var block2 *wire.Block

				BeforeEach(func() {
					block2 = testdata.BlockSet1[0]
					Expect(err).To(BeNil())

					err = chain2.store.PutBlock(chain2.id, block2)
					Expect(err).To(BeNil())
				})

				It("should return chain (not chain2) and header matching the header of the recently aded block", func() {
					_block, chain, tipHeader, err := bc.findBlockChainByHash(block.GetHash().HexStr())
					Expect(err).To(BeNil())
					Expect(block).To(Equal(_block))
					Expect(chain.GetID()).To(Equal(chain2.id))
					Expect(tipHeader.ComputeHash()).To(Equal(block2.Header.ComputeHash()))
				})
			})
		})

		Describe(".loadChain", func() {
			var block *wire.Block
			var chain, subChain *Chain

			BeforeEach(func() {
				bc.chains = make(map[string]*Chain)
			})

			BeforeEach(func() {
				chain = NewChain("chain_2", testStore, cfg, log)
				subChain = NewChain("sub_chain", testStore, cfg, log)

				block = testdata.BlockSet1[0]
				Expect(err).To(BeNil())

				err := bc.saveChain(chain, "", 0)
				Expect(err).To(BeNil())
				err = chain.append(block)
				Expect(err).To(BeNil())

				err = bc.saveChain(subChain, chain.GetID(), block.GetNumber())
				Expect(err).To(BeNil())
			})

			It("should return error when only ParentBlockNumber is set but ParentChainID is unset", func() {
				err = bc.loadChain(&common.ChainInfo{ID: chain.GetID(), ParentBlockNumber: 1})
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("chain load failed: parent chain id and parent block id are both required"))
			})

			It("should return error when only ParentChainID is set but ParentBlockNumber is unset", func() {
				err = bc.loadChain(&common.ChainInfo{ID: chain.GetID(), ParentChainID: "some_id"})
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("chain load failed: parent chain id and parent block id are both required"))
			})

			It("should return error when parent block does not exist", func() {
				err = bc.loadChain(&common.ChainInfo{ID: chain.GetID(), ParentChainID: "some_id", ParentBlockNumber: 100})
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("chain load failed: parent block {100} of chain {chain_2} not found"))
			})

			It("should successfully load chain with no parent into the cache", func() {
				err = bc.loadChain(&common.ChainInfo{ID: chain.GetID()})
				Expect(err).To(BeNil())
				Expect(bc.chains).To(HaveLen(2))
				Expect(bc.chains[chain.GetID()]).ToNot(BeNil())
			})

			It("should successfully load chain into the cache with parent block and chain info set", func() {
				err = bc.loadChain(&common.ChainInfo{ID: subChain.GetID(), ParentChainID: chain.GetID(), ParentBlockNumber: block.GetNumber()})
				Expect(err).To(BeNil())
				Expect(bc.chains).To(HaveLen(2))
				Expect(bc.chains[subChain.GetID()]).ToNot(BeNil())
			})
		})

		Describe(".findChainInfo", func() {

			var chain *Chain

			BeforeEach(func() {
				chain = NewChain("chain_a", testStore, cfg, log)
				err := bc.saveChain(chain, "", 0)
				Expect(err).To(BeNil())
			})

			It("should find chain with id = chain_a", func() {
				chInfo, err := bc.findChainInfo("chain_a")
				Expect(err).To(BeNil())
				Expect(chInfo.ID).To(Equal(chain.GetID()))
			})

			It("should return err = 'chain not found' if chain does not exist", func() {
				_, err := bc.findChainInfo("chain_b")
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("chain not found"))
			})
		})

		Describe(".newChain", func() {

			It("should return error if block is nil", func() {
				_, err := bc.newChain(nil, nil, nil, nil)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("initial block cannot be nil"))
			})

			It("should return error if initial block parent is nil", func() {
				initialBlock := testdata.BlockSet1[0]
				_, err := bc.newChain(nil, initialBlock, nil, nil)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("initial block parent cannot be nil"))
			})

			It("should return error if initial block and parent are not related", func() {
				initialBlock := testdata.BlockSet1[1]
				initialBlockParent := testdata.BlockSet1[2]
				_, err = bc.newChain(nil, initialBlock, initialBlockParent, nil)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("initial block and parent are not related"))
			})

			It("should successfully return a new chain", func() {
				initialBlock := testdata.BlockSet1[1]
				initialBlockParent := testdata.BlockSet1[0]

				tx, _ := chain.store.NewTx()
				chain, err := bc.newChain(tx, initialBlock, initialBlockParent, nil)
				Expect(err).To(BeNil())
				Expect(chain).ToNot(BeNil())
				Expect(chain.parentBlock).To(Equal(initialBlockParent))
				tx.Commit()
			})
		})

		Describe(".GetTransaction", func() {
			var block *wire.Block
			var chain *Chain

			BeforeEach(func() {
				chain = NewChain("chain_a", testStore, cfg, log)
				block = testdata.BlockSet1[0]
				chain.append(block)
				err = chain.putTransactions(block.Transactions)
				Expect(err).To(BeNil())
			})

			It("should return err = 'best chain unknown' if the best chain has not been decided", func() {
				bc.bestChain = nil
				_, err := bc.GetTransaction(block.Transactions[0].ID())
				Expect(err).ToNot(BeNil())
				Expect(err).To(Equal(common.ErrBestChainUnknown))
			})

			It("should return transaction and no error", func() {
				bc.bestChain = chain
				tx, err := bc.GetTransaction(block.Transactions[0].ID())
				Expect(err).To(BeNil())
				Expect(tx).To(Equal(block.Transactions[0]))
			})

			It("should return err = 'transaction not found' when main chain does not have the transaction", func() {
				bc.bestChain = chain
				tx, err := bc.GetTransaction("unknown")
				Expect(err).ToNot(BeNil())
				Expect(err).To(Equal(common.ErrTxNotFound))
				Expect(tx).To(BeNil())
			})
		})

		Describe(".chooseBestChain", func() {

			// Context("with one highest block", func() {

			// 	var chainA, chainB *Chain

			// 	BeforeEach(func() {
			// 		bc.chains = make(map[string]*Chain)
			// 	})

			// 	BeforeEach(func() {
			// 		chainA = NewChain("chain_a", testStore, cfg, log)
			// 		err := bc.saveChain(chainA, "", 0)
			// 		Expect(err).To(BeNil())
			// 		block, err := wire.BlockFromString(testdata.ChooseBestChainData[0])
			// 		err = chainA.append(block)
			// 		Expect(err).To(BeNil())
			// 	})

			// 	BeforeEach(func() {
			// 		chainB = NewChain("chain_b", testStore, cfg, log)
			// 		err := bc.saveChain(chainB, "", 0)
			// 		Expect(err).To(BeNil())
			// 		block, err := wire.BlockFromString(testdata.ChooseBestChainData[0])
			// 		err = chainB.append(block)
			// 		Expect(err).To(BeNil())
			// 	})

			// 	It("should", func() {
			// 		bc.chooseBestChain()
			// 		_ = 2
			// 	})

			// })
			Context("with one highest block", func() {

				var chainA, chainB, chainC *Chain

				BeforeEach(func() {
					bc.chains = make(map[string]*Chain)
				})

				BeforeEach(func() {
					chainA = NewChain("chain_a", testStore, cfg, log)
					err := bc.saveChain(chainA, "", 0)
					Expect(err).To(BeNil())
					block = testdata.BlockSet1[0]
					err = chainA.append(block)
					Expect(err).To(BeNil())
				})

				BeforeEach(func() {
					chainB = NewChain("chain_b", testStore, cfg, log)
					err := bc.saveChain(chainB, "", 0)
					Expect(err).To(BeNil())
					block = testdata.BlockSet1[0]
					err = chainB.append(block)
					Expect(err).To(BeNil())
				})

				BeforeEach(func() {
					chainC = NewChain("chain_c", testStore, cfg, log)
					err := bc.saveChain(chainC, "", 0)
					Expect(err).To(BeNil())
					block = testdata.BlockSet1[0]
					block2 := testdata.BlockSet1[1]
					err = chainC.append(block)
					Expect(err).To(BeNil())
					err = chainC.append(block2)
					Expect(err).To(BeNil())
				})

				It("should return 'chain_c' as the highest block", func() {
					bestChain, err := bc.chooseBestChain()
					Expect(err).To(BeNil())
					Expect(bestChain).To(Equal(chainC))
				})
			})
		})
	})

}
