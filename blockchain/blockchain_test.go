package blockchain

import (
	"math/big"

	"github.com/ellcrys/elld/blockchain/common"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/wire"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var BlockchainTest = func() bool {

	return Describe("Blockchain", func() {

		Describe(".SetStore", func() {
			It("should set store", func() {
				bc := New(nil, cfg, log)
				bc.SetDB(db)
			})
		})

		Describe(".addChain", func() {
			It("should add chain", func() {
				chain := NewChain("chain_id", db, cfg, log)
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
				chain = NewChain("chain_id", db, cfg, log)
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
				chain = NewChain("chain_id", db, cfg, log)
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

			var block2 *wire.Block

			BeforeEach(func() {
				bc.chains = make(map[util.String]*Chain)
				chain := NewChain("c1", db, cfg, log)
				block2 = MakeTestBlock(bc, chain, &common.GenerateBlockParams{
					Transactions: []*wire.Transaction{
						wire.NewTx(wire.TxTypeAllocCoin, 123, util.String(sender.Addr()), sender, "1", "0.1", 1532730722),
					},
					Creator:    sender,
					Nonce:      wire.EncodeNonce(1),
					MixHash:    util.BytesToHash([]byte("mix hash")),
					Difficulty: new(big.Int).SetInt64(500),
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
					block := MakeTestBlock(bc, chain, &common.GenerateBlockParams{
						Transactions: []*wire.Transaction{
							wire.NewTx(wire.TxTypeAllocCoin, 123, util.String(sender.Addr()), sender, "1", "0.1", 1532730722),
						},
						Creator:    sender,
						Nonce:      wire.EncodeNonce(1),
						MixHash:    util.BytesToHash([]byte("mix hash")),
						Difficulty: new(big.Int).SetInt64(500),
					})
					block.Header.Number = 2
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

					block := MakeTestBlock(bc, chain, &common.GenerateBlockParams{
						Transactions: []*wire.Transaction{
							wire.NewTx(wire.TxTypeBalance, 123, util.String(sender.Addr()), sender, "1", "0.1", 1532730722),
						},
						Creator:    sender,
						Nonce:      wire.EncodeNonce(1),
						MixHash:    util.BytesToHash([]byte("mix hash")),
						Difficulty: new(big.Int).SetInt64(500),
					})
					block.Transactions[0].From = "unknown_account"
					block.Hash = block.ComputeHash()
					block.Sig, err = wire.BlockSign(block, sender.PrivKey().Base58())
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
					block := MakeTestBlock(bc, chain, &common.GenerateBlockParams{
						Transactions: []*wire.Transaction{
							wire.NewTx(wire.TxTypeAllocCoin, 123, util.String(sender.Addr()), sender, "1", "0.1", 1532730722),
						},
						Creator:    sender,
						Nonce:      wire.EncodeNonce(1),
						MixHash:    util.BytesToHash([]byte("mix hash")),
						Difficulty: new(big.Int).SetInt64(500),
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

				c1 := NewChain("c1", db, cfg, log)
				c2 := NewChain("c2", db, cfg, log)

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

			var b1 *wire.Block
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

				b1 = MakeTestBlock(bc, chain2, &common.GenerateBlockParams{
					Transactions: []*wire.Transaction{
						wire.NewTx(wire.TxTypeBalance, 123, util.String(receiver.Addr()), sender, "1", "0.1", 1532730723),
					},
					Creator:    sender,
					Nonce:      wire.EncodeNonce(1),
					MixHash:    util.BytesToHash([]byte("mix hash")),
					Difficulty: new(big.Int).SetInt64(500),
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
					Expect(header.ComputeHash()).To(Equal(b1.Header.ComputeHash()))
				})
			})

			Context("when the hash belongs to a block that is not the highest block", func() {

				var b2 *wire.Block

				BeforeEach(func() {
					b2 = MakeTestBlock(bc, genesisChain, &common.GenerateBlockParams{
						Transactions: []*wire.Transaction{
							wire.NewTx(wire.TxTypeBalance, 123, util.String(receiver.Addr()), sender, "1", "0.1", 1532730724),
						},
						Creator:    sender,
						Nonce:      wire.EncodeNonce(1),
						MixHash:    util.BytesToHash([]byte("mix hash")),
						Difficulty: new(big.Int).SetInt64(500),
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
					Expect(tipHeader.ComputeHash()).To(Equal(b2.Header.ComputeHash()))
				})
			})
		})

		Describe(".loadChain", func() {
			var block *wire.Block
			var chain, subChain *Chain

			BeforeEach(func() {
				bc.chains = make(map[util.String]*Chain)
			})

			BeforeEach(func() {
				chain = NewChain("chain_2", db, cfg, log)
				err := bc.saveChain(chain, "", 0)
				Expect(err).To(BeNil())

				subChain = NewChain("sub_chain", db, cfg, log)

				block = MakeTestBlock(bc, genesisChain, &common.GenerateBlockParams{
					Transactions: []*wire.Transaction{
						wire.NewTx(wire.TxTypeBalance, 123, util.String(receiver.Addr()), sender, "1", "0.1", 1532730724),
					},
					Creator:    sender,
					Nonce:      wire.EncodeNonce(1),
					MixHash:    util.BytesToHash([]byte("mix hash")),
					Difficulty: new(big.Int).SetInt64(500),
				})

				err = chain.append(block)
				Expect(err).To(BeNil())

				err = bc.saveChain(subChain, chain.GetID(), block.GetNumber())
				Expect(err).To(BeNil())
			})

			It("should return error when only ParentBlockNumber is set but ParentChainID is unset", func() {
				err = bc.loadChain(&common.ChainInfo{ID: "some_id", ParentBlockNumber: 1})
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
				chain = NewChain("chain_a", db, cfg, log)
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

			var parentBlock, block, unknownParent *wire.Block

			BeforeEach(func() {
				parentBlock = MakeTestBlock(bc, genesisChain, &common.GenerateBlockParams{
					Transactions: []*wire.Transaction{
						wire.NewTx(wire.TxTypeBalance, 123, util.String(receiver.Addr()), sender, "1", "0.1", 1532730724),
					},
					Creator:    sender,
					Nonce:      wire.EncodeNonce(1),
					MixHash:    util.BytesToHash([]byte("mix hash")),
					Difficulty: new(big.Int).SetInt64(500),
				})

				block = MakeTestBlock(bc, genesisChain, &common.GenerateBlockParams{
					Transactions: []*wire.Transaction{
						wire.NewTx(wire.TxTypeBalance, 123, util.String(receiver.Addr()), sender, "1", "0.1", 1532730724),
					},
					OverrideParentHash: parentBlock.Hash,
					Creator:            sender,
					Nonce:              wire.EncodeNonce(1),
					MixHash:            util.BytesToHash([]byte("mix hash")),
					Difficulty:         new(big.Int).SetInt64(500),
				})

				unknownParent = MakeTestBlock(bc, genesisChain, &common.GenerateBlockParams{
					Transactions: []*wire.Transaction{
						wire.NewTx(wire.TxTypeBalance, 123, util.String(receiver.Addr()), sender, "1", "0.1", 1532730724),
					},
					Creator:            sender,
					OverrideParentHash: util.StrToHash("unknown"),
					Nonce:              wire.EncodeNonce(1),
					MixHash:            util.BytesToHash([]byte("mix hash")),
					Difficulty:         new(big.Int).SetInt64(500),
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
			var block *wire.Block
			var chain *Chain

			BeforeEach(func() {
				chain = NewChain("chain_a", db, cfg, log)

				block = MakeTestBlock(bc, genesisChain, &common.GenerateBlockParams{
					Transactions: []*wire.Transaction{
						wire.NewTx(wire.TxTypeBalance, 123, util.String(receiver.Addr()), sender, "1", "0.1", 1532730724),
					},
					Creator:    sender,
					Nonce:      wire.EncodeNonce(1),
					MixHash:    util.BytesToHash([]byte("mix hash")),
					Difficulty: new(big.Int).SetInt64(500),
				})

				chain.append(block)
				err = chain.PutTransactions(block.Transactions)
				Expect(err).To(BeNil())
			})

			It("should return err = 'best chain unknown' if the best chain has not been decided", func() {
				bc.bestChain = nil
				_, err := bc.GetTransaction(block.Transactions[0].Hash)
				Expect(err).ToNot(BeNil())
				Expect(err).To(Equal(common.ErrBestChainUnknown))
			})

			It("should return transaction and no error", func() {
				bc.bestChain = chain
				tx, err := bc.GetTransaction(block.Transactions[0].Hash)
				Expect(err).To(BeNil())
				Expect(tx).To(Equal(block.Transactions[0]))
			})

			It("should return err = 'transaction not found' when main chain does not have the transaction", func() {
				bc.bestChain = chain
				tx, err := bc.GetTransaction(util.Hash{1, 2, 3})
				Expect(err).ToNot(BeNil())
				Expect(err).To(Equal(common.ErrTxNotFound))
				Expect(tx).To(BeNil())
			})
		})

		Describe(".chooseBestChain", func() {

			// Context("with one highest block", func() {

			// 	var chainA, chainB *Chain

			// 	BeforeEach(func() {
			// 		bc.chains = make(map[util.String]*Chain)
			// 	})

			// 	BeforeEach(func() {
			// 		chainA = NewChain("chain_a", db, cfg, log)
			// 		err := bc.saveChain(chainA, "", 0)
			// 		Expect(err).To(BeNil())
			// 		block, err := wire.BlockFromString(testdata.ChooseBestChainData[0])
			// 		err = chainA.append(block)
			// 		Expect(err).To(BeNil())
			// 	})

			// 	BeforeEach(func() {
			// 		chainB = NewChain("chain_b", db, cfg, log)
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
					bc.chains = make(map[util.String]*Chain)
				})

				BeforeEach(func() {
					chainA = NewChain("chain_a", db, cfg, log)
					err := bc.saveChain(chainA, "", 0)
					Expect(err).To(BeNil())

					block := MakeTestBlock(bc, chainA, &common.GenerateBlockParams{
						Transactions: []*wire.Transaction{
							wire.NewTx(wire.TxTypeAllocCoin, 123, util.String(sender.Addr()), sender, "1", "0.1", 1532730724),
						},
						Creator:    sender,
						Nonce:      wire.EncodeNonce(1),
						MixHash:    util.BytesToHash([]byte("mix hash")),
						Difficulty: new(big.Int).SetInt64(500),
					})

					err = chainA.append(block)
					Expect(err).To(BeNil())
				})

				BeforeEach(func() {
					chainB = NewChain("chain_b", db, cfg, log)
					err := bc.saveChain(chainB, "", 0)
					Expect(err).To(BeNil())

					block := MakeTestBlock(bc, chainB, &common.GenerateBlockParams{
						Transactions: []*wire.Transaction{
							wire.NewTx(wire.TxTypeAllocCoin, 123, util.String(sender.Addr()), sender, "1", "0.1", 1532730725),
						},
						Creator:    sender,
						Nonce:      wire.EncodeNonce(1),
						MixHash:    util.BytesToHash([]byte("mix hash")),
						Difficulty: new(big.Int).SetInt64(500),
					})

					err = chainB.append(block)
					Expect(err).To(BeNil())
				})

				BeforeEach(func() {
					chainC = NewChain("chain_c", db, cfg, log)
					err := bc.saveChain(chainC, "", 0)
					Expect(err).To(BeNil())

					block := MakeTestBlock(bc, chainC, &common.GenerateBlockParams{
						Transactions: []*wire.Transaction{
							wire.NewTx(wire.TxTypeAllocCoin, 123, util.String(sender.Addr()), sender, "1", "0.1", 1532730726),
						},
						Creator:    sender,
						Nonce:      wire.EncodeNonce(1),
						MixHash:    util.BytesToHash([]byte("mix hash")),
						Difficulty: new(big.Int).SetInt64(500),
					})
					err = chainC.append(block)
					Expect(err).To(BeNil())

					block2 := MakeTestBlock(bc, chainC, &common.GenerateBlockParams{
						Transactions: []*wire.Transaction{
							wire.NewTx(wire.TxTypeAllocCoin, 123, util.String(sender.Addr()), sender, "1", "0.1", 1532730727),
						},
						Creator:    sender,
						Nonce:      wire.EncodeNonce(1),
						MixHash:    util.BytesToHash([]byte("mix hash")),
						Difficulty: new(big.Int).SetInt64(500),
					})
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
