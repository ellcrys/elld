package blockchain

import (
	"math/big"
	"time"

	"github.com/jinzhu/copier"

	"github.com/ellcrys/elld/blockchain/common"
	"github.com/ellcrys/elld/elldb"
	"github.com/ellcrys/elld/txpool"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/types/core/objects"
	"github.com/ellcrys/elld/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var BlockchainTests = func() bool {
	BlockchainIntegrationTests()
	BlockchainUnitTests()
	return true
}

var BlockchainIntegrationTests = func() {
	Describe("Blockchain", func() {

		Describe(".Up", func() {

			txp := txpool.New(1)

			BeforeEach(func() {
				bc = New(txp, cfg, log)
			})

			It("should return error if db is not set", func() {
				err = bc.Up()
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("db has not been initialized"))
			})

			It("should succeed", func() {
				bc.SetDB(db)
				err = bc.Up()
				Expect(err).To(BeNil())
			})

			When("genesis block is invalid", func() {

				BeforeEach(func() {
					db = elldb.NewDB(cfg.ConfigDir())
					err = db.Open(util.RandString(5))
					Expect(err).To(BeNil())
					bc = New(txp, cfg, log)
					bc.SetDB(db)
					bc.SetGenesisBlock(GenesisBlock)
				})

				When("block number is > 1", func() {

					var invalidBlock core.Block

					BeforeEach(func() {
						bc.bestChain = genesisChain
						invalidBlock = makeBlock(genesisChain)
						copier.Copy(&invalidBlock, GenesisBlock)
						invalidBlock.GetHeader().SetNumber(2)
					})

					It("should return error if genesis block number is not equal to 1", func() {
						bc.SetGenesisBlock(invalidBlock)
						err = bc.Up()
						Expect(err).ToNot(BeNil())
						Expect(err.Error()).To(Equal("genesis block error: expected block number 1"))
					})
				})

				When("sender account does not exist", func() {

					BeforeEach(func() {
						bc.bestChain = genesisChain
						chain := NewChain("c1", db, cfg, log)
						Expect(bc.CreateAccount(1, chain, &objects.Account{
							Type:    objects.AccountTypeBalance,
							Address: util.String(sender.Addr()),
							Balance: "1000",
						})).To(BeNil())

						block := makeBlockWithBalanceTx(chain)

						// modify a transaction's sender to one that
						// does not exist
						unknownSenderTx := block.GetTransactions()[0]
						unknownSenderTx.SetSenderPubKey(util.String(receiver.PubKey().Base58()))
						unknownSenderTx.SetFrom(util.String(receiver.Addr()))
						unknownSenderTx.SetHash(unknownSenderTx.ComputeHash())
						txSig, _ := objects.TxSign(unknownSenderTx, receiver.PrivKey().Base58())
						unknownSenderTx.SetSignature(txSig)

						// recompute block, transactions root, hash and signature
						block.GetHeader().SetTransactionsRoot(common.ComputeTxsRoot(block.GetTransactions()))
						block.SetHash(block.ComputeHash())
						blockSig, _ := objects.BlockSign(block, sender.PrivKey().Base58())
						block.SetSignature(blockSig)
						bc.SetGenesisBlock(block)
						block.SetChainReader(nil)
					})

					It("should return error if a transaction's sender does not exists", func() {
						err = bc.Up()
						Expect(err).ToNot(BeNil())
						Expect(err.Error()).To(Equal("genesis block error: tx:0, field:from, error:sender account not found"))
					})
				})
			})

			It("should load all chains", func() {
				bc.SetDB(db)
				c1 := NewChain("c_1", db, cfg, log)
				c2 := NewChain("c_2", db, cfg, log)

				err = bc.saveChain(c1, "", 0)
				Expect(err).To(BeNil())

				err = bc.saveChain(c2, "", 0)
				Expect(err).To(BeNil())

				err = c1.append(GenesisBlock)
				Expect(err).To(BeNil())

				bc.SetGenesisBlock(GenesisBlock)
				err = bc.Up()
				Expect(err).To(BeNil())

				Expect(bc.chains).To(HaveLen(3))
				Expect(bc.chains).To(HaveKey(c1.id))
				Expect(bc.chains).To(HaveKey(c2.id))
				Expect(bc.chains).To(HaveKey(genesisChain.id))
			})
		})

		Describe(".findChainByLastBlockHash", func() {

			var b1 core.Block
			var chain2 *Chain

			BeforeEach(func() {
				chain2 = NewChain("chain2", db, cfg, log)
				bc.addChain(chain2)

				Expect(bc.CreateAccount(1, chain2, &objects.Account{
					Type:    objects.AccountTypeBalance,
					Address: util.String(sender.Addr()),
					Balance: "1000",
				})).To(BeNil())

				b1 = MakeTestBlock(bc, chain2, &core.GenerateBlockParams{
					Transactions: []core.Transaction{
						objects.NewTx(objects.TxTypeBalance, 1, util.String(receiver.Addr()), sender, "1", "2.36", 1532730723),
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
							objects.NewTx(objects.TxTypeBalance, 1, util.String(receiver.Addr()), sender, "1", "2.36", 1532730724),
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
						objects.NewTx(objects.TxTypeBalance, 1, util.String(receiver.Addr()), sender, "1", "2.36", 1532730724),
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
				Expect(err.Error()).To(Equal("chain load failed: chain parent chain ID and block are required"))
			})

			It("should return error when only ParentChainID is set but ParentBlockNumber is unset", func() {
				err = bc.loadChain(&core.ChainInfo{ID: chain.GetID(), ParentChainID: "some_id"})
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("chain load failed: chain parent chain ID and block are required"))
			})

			It("should return error when parent chain does not exist", func() {
				err = bc.loadChain(&core.ChainInfo{ID: chain.GetID(), ParentChainID: "some_id", ParentBlockNumber: 100})
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("chain load failed: chain parent not found"))
			})

			It("should return error when parent block does not exist", func() {
				err = bc.loadChain(&core.ChainInfo{ID: chain.GetID(), ParentChainID: "chain_2", ParentBlockNumber: 100})
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("chain load failed: chain parent block not found"))
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
						objects.NewTx(objects.TxTypeBalance, 1, util.String(receiver.Addr()), sender, "1", "2.36", 1532730724),
					},
					Creator:    sender,
					Nonce:      core.EncodeNonce(1),
					Difficulty: new(big.Int).SetInt64(131072),
				})

				block = MakeTestBlock(bc, genesisChain, &core.GenerateBlockParams{
					Transactions: []core.Transaction{
						objects.NewTx(objects.TxTypeBalance, 1, util.String(receiver.Addr()), sender, "1", "2.36", 1532730724),
					},
					OverrideParentHash: parentBlock.GetHash(),
					Creator:            sender,
					Nonce:              core.EncodeNonce(1),
					Difficulty:         new(big.Int).SetInt64(131072),
					OverrideTimestamp:  time.Now().Add(2 * time.Second).Unix(),
				})

				unknownParent = MakeTestBlock(bc, genesisChain, &core.GenerateBlockParams{
					Transactions: []core.Transaction{
						objects.NewTx(objects.TxTypeBalance, 1, util.String(receiver.Addr()), sender, "1", "2.36", 1532730724),
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
						objects.NewTx(objects.TxTypeBalance, 1, util.String(receiver.Addr()), sender, "1", "2.36", 1532730724),
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
	})
}

var BlockchainUnitTests = func() {
	Describe("UnitBlockchainTest", func() {

		Describe(".LoadBlockFromFile", func() {
			It("should return block", func() {
				b, err := LoadBlockFromFile("genesis-small.json")
				Expect(err).To(BeNil())
				Expect(b).ToNot(BeNil())
			})

			It("should return err='block file not found' when file does not exist", func() {
				b, err := LoadBlockFromFile("unknown.json")
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("block file not found"))
				Expect(b).To(BeNil())
			})

			It("should return err when file is malformed", func() {
				b, err := LoadBlockFromFile("genesis-small.malformed.json")
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("invalid character ',' looking for beginning of value"))
				Expect(b).To(BeNil())
			})
		})

		Describe(".IsMainChain", func() {
			It("should return false when the given chain is not the main chain", func() {
				ch := NewChain("c1", db, cfg, log)
				Expect(bc.IsMainChain(ch.ChainReader())).To(BeFalse())
			})

			It("should return true when the given chain is the main chain", func() {
				ch := NewChain("c1", db, cfg, log)
				bc.bestChain = ch
				Expect(bc.IsMainChain(ch.ChainReader())).To(BeTrue())
			})
		})

		Describe(".ChainReader", func() {
			It("should return a chain reader with same ID as the chain", func() {
				Expect(bc.ChainReader().GetID()).To(Equal(genesisChain.id))
			})
		})

		Describe(".GetChainReaderByHash", func() {
			It("should get chain of the genesis block", func() {
				reader := bc.GetChainReaderByHash(genesisBlock.GetHash())
				Expect(reader.GetID()).To(Equal(genesisChain.GetID()))
			})

			It("should nil when chain reader could not be found", func() {
				reader := bc.GetChainReaderByHash(util.StrToHash("invalid_unknown"))
				Expect(reader).To(BeNil())
			})
		})

		Describe(".GetChainsReader", func() {
			It("should get chain readers for all known chains", func() {
				ch := NewChain("c1", db, cfg, log)
				bc.chains[ch.id] = ch
				Expect(bc.chains).To(HaveLen(2))
				readers := bc.GetChainsReader()
				Expect(readers).To(HaveLen(2))
				expectedChains := []string{genesisChain.id.String(), ch.id.String()}
				Expect(expectedChains).To(ContainElement(genesisChain.GetID().String()))
				Expect(expectedChains).To(ContainElement(ch.GetID().String()))
			})
		})

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
				bc.addChain(chain)
				Expect(bc.chains).To(HaveLen(2))
			})
		})

		Describe(".removeChain", func() {

			var chain *Chain

			BeforeEach(func() {
				chain = NewChain("chain_id", db, cfg, log)
				Expect(err).To(BeNil())
				bc.addChain(chain)
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
				bc.addChain(chain)
				Expect(bc.hasChain(chain)).To(BeTrue())
			})
		})
	})
}
