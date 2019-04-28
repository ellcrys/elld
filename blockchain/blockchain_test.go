package blockchain

import (
	"os"
	"time"

	. "github.com/ellcrys/elld/blockchain/testutil"
	"github.com/ellcrys/elld/blockchain/txpool"
	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/elldb"
	"github.com/ellcrys/elld/testutil"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/types/core"

	"github.com/ellcrys/elld/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("IntegrationBlockchain", func() {

	var err error
	var cfg *config.EngineConfig
	var db elldb.DB
	var genesisBlock types.Block
	var genesisChain *Chain
	var sender, receiver *crypto.Key

	BeforeEach(func() {
		cfg, err = testutil.SetTestCfg()
		Expect(err).To(BeNil())

		db = elldb.NewDB(cfg.NetDataDir())
		err = db.Open(util.RandString(5))
		Expect(err).To(BeNil())

		sender = crypto.NewKeyFromIntSeed(1)
		receiver = crypto.NewKeyFromIntSeed(2)
	})

	AfterEach(func() {
		db.Close()
		err = os.RemoveAll(cfg.DataDir())
		Expect(err).To(BeNil())
	})

	Describe(".Up", func() {
		var bc *Blockchain

		BeforeEach(func() {
			bc = New(txpool.New(100), cfg, log)
			bc.SetDB(db)
			bc.SetCoinbase(crypto.NewKeyFromIntSeed(1234))
		})

		AfterEach(func() {
			bc.chains = map[util.String]*Chain{}
			bc.bestChain = nil
		})

		When("db is not set", func() {
			It("should return err='db has not been initialized'", func() {
				bc.SetDB(nil)
				err = bc.Up()
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("db has not been initialized"))
			})
		})

		When("no genesis block file does not exist", func() {
			It("should return err='block file not found'", func() {
				GenesisBlockFileName = "unknown.json"
				err = bc.Up()
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("block file not found"))
			})
		})

		When("genesis block number is not equal to 1", func() {
			BeforeEach(func() {
				genesisBlock, err = LoadBlockFromFile("genesis-test.json")
				Expect(err).To(BeNil())
				genesisBlock.GetHeader().SetNumber(2)
				bc.SetGenesisBlock(genesisBlock)
			})

			It("should return err='genesis block error: expected block number 1'", func() {
				err = bc.Up()
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("genesis block error: expected block number 1"))
			})
		})

		When("successful (no error occurred)", func() {
			BeforeEach(func() {
				genesisBlock, err = LoadBlockFromFile("genesis-test.json")
				Expect(err).To(BeNil())
				bc.SetGenesisBlock(genesisBlock)
				err = bc.Up()
				Expect(err).To(BeNil())
			})

			Specify("that only one chain exist", func() {
				Expect(bc.chains).To(HaveLen(1))
			})

			Specify("that the best chain is the same with the only known chain", func() {
				Expect(bc.chains[bc.bestChain.GetID()]).ToNot(BeNil())
			})

			When("chain is reset and .Up is called again", func() {
				BeforeEach(func() {
					err = bc.Up()
					Expect(err).To(BeNil())
				})

				It("should populate chain index with one chain", func() {
					Expect(bc.chains).To(HaveLen(1))
				})

				Specify("that the best chain is the same with the only known chain", func() {
					Expect(bc.chains[bc.bestChain.GetID()]).ToNot(BeNil())
				})
			})
		})

		When("genesis block is invalid", func() {

			BeforeEach(func() {
				invalidBlock, err := LoadBlockFromFile("genesis-test.json")
				Expect(err).To(BeNil())
				invalidBlock.SetHash(util.Hash{0, 0, 1, 2, 3})
				bc.SetGenesisBlock(invalidBlock)
			})

			It("should return error", func() {
				err = bc.Up()
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("genesis block error: field:hash, error:hash is not correct"))
			})
		})
	})

	When("blockchain has been initialized", func() {

		var bc *Blockchain

		BeforeEach(func() {
			bc = New(txpool.New(100), cfg, log)
			bc.SetDB(db)
			bc.SetCoinbase(crypto.NewKeyFromIntSeed(1234))
		})

		BeforeEach(func() {
			genesisBlock, err = LoadBlockFromFile("genesis-test.json")
			Expect(err).To(BeNil())
			bc.SetGenesisBlock(genesisBlock)
			err = bc.Up()
			Expect(err).To(BeNil())
			genesisChain = bc.bestChain
		})

		Describe(".findChainByBlockHash", func() {

			It("should return chain=nil, header=nil, err=nil if no block on the chain matches the hash", func() {
				result, chain, header, err := bc.findChainByBlockHash(util.Hash{1, 2, 3})
				Expect(err).To(BeNil())
				Expect(result).To(BeNil())
				Expect(header).To(BeNil())
				Expect(chain).To(BeNil())
			})

			When("one chain includes the block", func() {
				var b1 types.Block
				var chain2 *Chain

				BeforeEach(func() {
					chain2 = NewChain("chain2", db, cfg, log)
					bc.addChain(chain2)

					err := bc.CreateAccount(1, chain2, &core.Account{
						Type:    core.AccountTypeBalance,
						Address: util.String(sender.Addr()),
						Balance: "100",
					})
					Expect(err).To(BeNil())

					b1 = MakeBlock(bc, chain2, sender, receiver)
					err = chain2.append(b1)
					Expect(err).To(BeNil())
				})

				It("should return the expected chain, chain tip header, queried block", func() {
					qb, chain, header, err := bc.findChainByBlockHash(b1.GetHash())
					Expect(err).To(BeNil())
					Expect(chain.GetID()).To(Equal(chain2.id))
					Expect(header.ComputeHash().Equal(b1.GetHeader().ComputeHash())).To(BeTrue())
					Expect(qb.GetHash().Equal(b1.GetHash())).To(BeTrue())
				})
			})

			When("more than one chain include the block", func() {
				var b1 types.Block
				var chain2, olderChain *Chain

				BeforeEach(func() {
					chain2 = NewChain("chain2", db, cfg, log)
					chain2.info.Timestamp = time.Now().Unix()
					bc.addChain(chain2)

					olderChain = NewChain("older_chain", db, cfg, log)
					olderChain.info.Timestamp = time.Now().Unix() - 1000
					bc.addChain(olderChain)

					err := bc.CreateAccount(1, chain2, &core.Account{
						Type:    core.AccountTypeBalance,
						Address: util.String(sender.Addr()),
						Balance: "100",
					})
					Expect(err).To(BeNil())

					b1 = MakeBlock(bc, chain2, sender, receiver)
					err = chain2.append(b1)
					Expect(err).To(BeNil())

					err = olderChain.append(b1)
					Expect(err).To(BeNil())
				})

				It("should return the expected chain, chain tip header, queried block", func() {
					qb, chain, header, err := bc.findChainByBlockHash(b1.GetHash())
					Expect(err).To(BeNil())
					Expect(chain.GetID()).To(Equal(olderChain.GetID()))
					Expect(header.ComputeHash().Equal(b1.GetHeader().ComputeHash())).To(BeTrue())
					Expect(qb.GetHash().Equal(b1.GetHash())).To(BeTrue())
				})
			})
		})

		Describe(".loadChain", func() {
			var block types.Block
			var chain, childChain *Chain

			BeforeEach(func() {
				bc.chains = make(map[util.String]*Chain)
				chain = NewChain("chain_2", db, cfg, log)
				err := bc.saveChain(chain, "", 0)
				Expect(err).To(BeNil())
				block = MakeBlock(bc, genesisChain, sender, receiver)
				err = chain.append(block)
				Expect(err).To(BeNil())

				childChain = NewChain("child_chain", db, cfg, log)
				err = bc.saveChain(childChain, chain.GetID(), block.GetNumber())
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
				err = bc.loadChain(&core.ChainInfo{ID: childChain.GetID(), ParentChainID: chain.GetID(), ParentBlockNumber: block.GetNumber()})
				Expect(err).To(BeNil())
				Expect(bc.chains).To(HaveLen(2))
				Expect(bc.chains[childChain.GetID()]).ToNot(BeNil())
			})
		})

		Describe(".copyChains", func() {
			var ch, ch2 *Chain

			BeforeEach(func() {
				ch = NewChain("c1", db, cfg, log)
				ch2 = NewChain("c2", db, cfg, log)
				ch3 := NewChain("c3", db, cfg, log)
				bc.addChain(ch)
				bc.addChain(ch2)
				bc.addChain(ch3)
			})

			It("should return 4 chains", func() {
				chains := bc.copyChains([]util.String{})
				Expect(chains).To(HaveLen(4))
			})

			When("chain `c1` is added to the exclusion list", func() {
				It("should return 3 chains but not `c1`", func() {
					chains := bc.copyChains([]util.String{ch.GetID()})
					Expect(chains).To(HaveLen(3))
					Expect(chains).ToNot(ContainElement(ch))
				})
			})

			When("chain `c1` and `c2` are added to the exclusion list", func() {
				It("should return 2 chains but not `c1` and `c2`", func() {
					chains := bc.copyChains([]util.String{ch.GetID(), ch2.GetID()})
					Expect(chains).To(HaveLen(2))
					Expect(chains).ToNot(ContainElement(ch))
					Expect(chains).ToNot(ContainElement(ch2))
				})
			})
		})

		Describe(".copyChainsMap", func() {
			var ch, ch2 *Chain

			BeforeEach(func() {
				ch = NewChain("c1", db, cfg, log)
				ch2 = NewChain("c2", db, cfg, log)
				ch3 := NewChain("c3", db, cfg, log)
				bc.addChain(ch)
				bc.addChain(ch2)
				bc.addChain(ch3)
			})

			It("should return 4 chains", func() {
				chains := bc.copyChainsMap([]util.String{})
				Expect(chains).To(HaveLen(4))
			})

			When("chain `c1` is added to the exclusion list", func() {
				It("should return 3 chains but not `c1`", func() {
					chains := bc.copyChainsMap([]util.String{ch.GetID()})
					Expect(chains).To(HaveLen(3))
					Expect(chains).ToNot(ContainElement(ch))
				})
			})

			When("chain `c1` and `c2` are added to the exclusion list", func() {
				It("should return 2 chains but not `c1` and `c2`", func() {
					chains := bc.copyChainsMap([]util.String{ch.GetID(), ch2.GetID()})
					Expect(chains).To(HaveLen(2))
					Expect(chains).ToNot(ContainElement(ch))
					Expect(chains).ToNot(ContainElement(ch2))
				})
			})
		})

		Describe(".getRootChain", func() {

			// Create a chain with a parent and add the chain
			BeforeEach(func() {
				ch := NewChain("abc", db, cfg, log)
				ch.parentChain = genesisChain
				bc.addChain(ch)
				Expect(bc.chains).To(HaveLen(2))
			})

			It("should get the chain with no branch", func() {
				root := bc.getRootChain()
				Expect(bc.chains).To(HaveLen(2))
				Expect(root.GetID().Equal(genesisChain.GetID())).To(BeTrue())
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

			var parentBlock, block, unknownParent types.Block

			BeforeEach(func() {
				parentBlock = MakeBlock(bc, genesisChain, sender, receiver)
				block = MakeBlockWithParentHash(bc, genesisChain, sender, parentBlock.GetHash())
				unknownParent = MakeBlockWithParentHash(bc, genesisChain, sender, util.StrToHash("unknown"))
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
			var block types.Block
			var chain *Chain

			BeforeEach(func() {
				chain = NewChain("chain_a", db, cfg, log)
				block = MakeBlock(bc, genesisChain, sender, receiver)
				err := chain.append(block)
				Expect(err).To(BeNil())
				err = chain.PutTransactions(block.GetTransactions(), block.GetNumber())
				Expect(err).To(BeNil())
			})

			Context("when best chain is not set", func() {
				It("should return err = 'best chain unknown' if the best chain has not been decided", func() {
					bc.bestChain = nil
					_, err := bc.GetTransaction(block.GetTransactions()[0].GetHash())
					Expect(err).ToNot(BeNil())
					Expect(err).To(Equal(core.ErrBestChainUnknown))
				})
			})

			Context("when best chain is set", func() {
				It("should return transaction and no error", func() {
					bc.bestChain = chain
					tx, err := bc.GetTransaction(block.GetTransactions()[0].GetHash())
					Expect(err).To(BeNil())
					Expect(tx).To(Equal(block.GetTransactions()[0]))
				})

				Context("when transaction does not exist", func() {
					It("should return ErrTxNotFound", func() {
						bc.bestChain = chain
						tx, err := bc.GetTransaction(util.Hash{1, 2, 3})
						Expect(err).ToNot(BeNil())
						Expect(err).To(Equal(core.ErrTxNotFound))
						Expect(tx).To(BeNil())
					})
				})
			})
		})

		Describe(".GetChainReaderByHash", func() {
			It("should get chain reader of the genesis block", func() {
				reader := bc.GetChainReaderByHash(genesisBlock.GetHash())
				Expect(reader.GetID()).To(Equal(genesisChain.GetID()))
			})

			It("should nil when chain reader could not be found", func() {
				reader := bc.GetChainReaderByHash(util.StrToHash("invalid_unknown"))
				Expect(reader).To(BeNil())
			})
		})

		Describe(".GetLocators", func() {

			Context("with ten blocks in the main chain", func() {

				var blocks []types.Block
				var locators []util.Hash

				BeforeEach(func() {
					blocks = []types.Block{}
					locators = []util.Hash{}
					blocks = append(blocks, genesisBlock)
				})

				BeforeEach(func() {
					for i := uint64(1); i <= 9; i++ {
						block := MakeBlockWithTx(bc, genesisChain, sender, i)
						_, err := bc.ProcessBlock(block)
						Expect(err).To(BeNil())
						blocks = append(blocks, block)
					}
					var err error
					locators, err = bc.GetLocators()
					Expect(err).To(BeNil())
				})

				It("should return 10 locators", func() {
					Expect(locators).To(HaveLen(10))
				})

				Specify("locators must match the initial blocks in reverse order", func() {
					Expect(len(locators)).To(Equal(len(blocks)))
					Expect(blocks[0].GetHash()).To(Equal(locators[9]))
					Expect(blocks[1].GetHash()).To(Equal(locators[8]))
					Expect(blocks[2].GetHash()).To(Equal(locators[7]))
					Expect(blocks[3].GetHash()).To(Equal(locators[6]))
					Expect(blocks[4].GetHash()).To(Equal(locators[5]))
					Expect(blocks[5].GetHash()).To(Equal(locators[4]))
					Expect(blocks[6].GetHash()).To(Equal(locators[3]))
					Expect(blocks[7].GetHash()).To(Equal(locators[2]))
					Expect(blocks[8].GetHash()).To(Equal(locators[1]))
					Expect(blocks[9].GetHash()).To(Equal(locators[0]))
				})
			})

			Context("with 20 blocks in the main chain", func() {
				var blocks []types.Block
				var locators []util.Hash

				BeforeEach(func() {
					blocks = []types.Block{}
					locators = []util.Hash{}
					blocks = append(blocks, genesisBlock)
				})

				BeforeEach(func() {
					for i := uint64(1); i <= 19; i++ {
						block := MakeBlockWithTx(bc, genesisChain, sender, i)
						_, err := bc.ProcessBlock(block)
						Expect(err).To(BeNil())
						blocks = append(blocks, block)
					}
					var err error
					locators, err = bc.GetLocators()
					Expect(err).To(BeNil())
				})

				It("should return 13 locators", func() {
					Expect(locators).To(HaveLen(13))
				})

				Specify("locators must be expected", func() {
					Expect(genesisBlock.GetHash()).To(Equal(locators[12]))
					Expect(blocks[4].GetHash()).To(Equal(locators[11]))
					Expect(blocks[8].GetHash()).To(Equal(locators[10]))
					Expect(blocks[10].GetHash()).To(Equal(locators[9]))
					Expect(blocks[11].GetHash()).To(Equal(locators[8]))
					Expect(blocks[12].GetHash()).To(Equal(locators[7]))
					Expect(blocks[13].GetHash()).To(Equal(locators[6]))
					Expect(blocks[14].GetHash()).To(Equal(locators[5]))
					Expect(blocks[15].GetHash()).To(Equal(locators[4]))
					Expect(blocks[16].GetHash()).To(Equal(locators[3]))
					Expect(blocks[17].GetHash()).To(Equal(locators[2]))
					Expect(blocks[18].GetHash()).To(Equal(locators[1]))
					Expect(blocks[19].GetHash()).To(Equal(locators[0]))
				})
			})
		})

		Describe(".SelectTransactions", func() {

			var tp types.TxPool
			var tx, tx2, tx3 *core.Transaction
			var txs []types.Transaction

			Context("pool has 1 transaction and account nonce = 0", func() {
				Context("transaction nonce is 2", func() {
					BeforeEach(func() {
						tp = bc.txPool
						tx = core.NewTx(core.TxTypeBalance, 2, util.String(sender.Addr()), sender, "0.1", "0.001", time.Now().Unix())
						tx.Hash = tx.ComputeHash()
						tp.Put(tx)
						maxSize := tx.GetSizeNoFee() + 100
						txs, err = bc.selectTransactions(maxSize)
						Expect(err).To(BeNil())
					})

					It("should return no transaction", func() {
						Expect(txs).To(HaveLen(0))
					})

					Specify("container should contain 1 transaction since selected txs go back in the pool", func() {
						Expect(tp.Container().Size()).To(Equal(int64(1)))
					})
				})
			})

			Context("pool has 2 transactions and account nonce = 0", func() {
				Context("tx(1) nonce = 1 and tx(2) nonce = 2", func() {
					var tp types.TxPool

					BeforeEach(func() {
						tp = bc.txPool
						tx = core.NewTx(core.TxTypeBalance, 1, util.String(sender.Addr()), sender, "0.1", "0.001", time.Now().Unix())
						tx.Hash = tx.ComputeHash()
						err := tp.Put(tx)
						Expect(err).To(BeNil())

						tx2 = core.NewTx(core.TxTypeBalance, 2, util.String(sender.Addr()), sender, "0.2", "0.001", time.Now().Unix())
						tx2.Hash = tx2.ComputeHash()
						err = tp.Put(tx2)
						Expect(tp.Container().Size()).To(Equal(int64(2)))
						Expect(err).To(BeNil())

						Expect(tp.Size()).To(Equal(int64(2)))
						maxSize := tx.GetSizeNoFee() + tx2.GetSizeNoFee()
						txs, err = bc.selectTransactions(maxSize)
						Expect(err).To(BeNil())
					})

					It("should return 2 transactions", func() {
						Expect(txs).To(HaveLen(2))
					})

					Specify("container should contain 2 transactions since selected txs go back in the pool", func() {
						Expect(tp.Size()).To(Equal(int64(2)))
					})
				})
			})

			Context("pool has 2 transactions and account nonce = 0", func() {
				Context("tx(1) nonce = 2 and tx(2) nonce = 1", func() {
					BeforeEach(func() {
						tp = bc.txPool
						tx = core.NewTx(core.TxTypeBalance, 2, util.String(sender.Addr()), sender, "0.1", "0.001", time.Now().Unix())
						tx.Hash = tx.ComputeHash()
						err := tp.Put(tx)
						Expect(err).To(BeNil())

						tx2 = core.NewTx(core.TxTypeBalance, 1, util.String(sender.Addr()), sender, "0.2", "0.001", time.Now().Unix())
						tx2.Hash = tx2.ComputeHash()
						err = tp.Put(tx2)
						Expect(tp.Container().Size()).To(Equal(int64(2)))
						Expect(err).To(BeNil())

						maxSize := tx.GetSizeNoFee() + tx2.GetSizeNoFee()
						txs, err = bc.selectTransactions(maxSize)
						Expect(err).To(BeNil())
					})

					It("should return 2 transactions with tx(2) selected before tx(1)", func() {
						Expect(txs).To(HaveLen(2))
						Expect(txs[0]).To(Equal(tx2))
						Expect(txs[1]).To(Equal(tx))
					})

					Specify("container should contain 2 transactions since selected txs go back in the pool", func() {
						Expect(tp.Container().Size()).To(Equal(int64(2)))
					})
				})
			})

			Context("pool has 2 transactions and account nonce = 0", func() {
				Context("tx(1) nonce = 1 and tx(2) nonce = 3", func() {
					BeforeEach(func() {
						tp = bc.txPool
						tx = core.NewTx(core.TxTypeBalance, 1, util.String(sender.Addr()), sender, "0.1", "0.001", time.Now().Unix())
						tx.Hash = tx.ComputeHash()
						err := tp.Put(tx)
						Expect(err).To(BeNil())

						tx2 = core.NewTx(core.TxTypeBalance, 3, util.String(sender.Addr()), sender, "0.2", "0.001", time.Now().Unix())
						tx2.Hash = tx2.ComputeHash()
						err = tp.Put(tx2)
						Expect(tp.Container().Size()).To(Equal(int64(2)))
						Expect(err).To(BeNil())

						maxSize := tx.GetSizeNoFee() + tx2.GetSizeNoFee()
						txs, err = bc.selectTransactions(maxSize)
						Expect(err).To(BeNil())
					})

					It("should return 1 transaction", func() {
						Expect(txs).To(HaveLen(1))
						Expect(txs[0]).To(Equal(tx))
					})

					Specify("container should contain 2 transactions since selected txs go back in the pool", func() {
						Expect(tp.Container().Size()).To(Equal(int64(2)))
					})
				})
			})

			Context("pool has 1 transaction and account nonce = 0", func() {
				Context("transaction nonce is 1", func() {
					BeforeEach(func() {
						tp = bc.txPool
						tx = core.NewTx(core.TxTypeBalance, 1, util.String(sender.Addr()), sender, "0.1", "0.001", time.Now().Unix())
						tx.Hash = tx.ComputeHash()
						tp.Put(tx)
						maxSize := tx.GetSizeNoFee() + 100
						txs, err = bc.selectTransactions(maxSize)
						Expect(err).To(BeNil())
					})

					It("should return 1 transaction", func() {
						Expect(txs).To(HaveLen(1))
					})

					Specify("container should contain 1 transaction since selected txs go back in the pool", func() {
						Expect(tp.Container().Size()).To(Equal(int64(1)))
					})
				})
			})

			Context("pool has 2 transactions and account nonce = 0", func() {
				Context("tx(1) nonce = 1 and tx(2) nonce = 1", func() {
					var tp types.TxPool

					BeforeEach(func() {
						tp = bc.txPool
						tx = core.NewTx(core.TxTypeBalance, 1, util.String(sender.Addr()), sender, "0.1", "0.001", time.Now().Unix())
						tx.Hash = tx.ComputeHash()
						err := tp.Put(tx)
						Expect(err).To(BeNil())

						tx2 = core.NewTx(core.TxTypeBalance, 1, util.String(sender.Addr()), sender, "0.2", "0.001", time.Now().Unix())
						tx2.Hash = tx2.ComputeHash()
						err = tp.Put(tx2)
						Expect(tp.Container().Size()).To(Equal(int64(2)))
						Expect(err).To(BeNil())

						Expect(tp.Size()).To(Equal(int64(2)))
						maxSize := tx.GetSizeNoFee() + tx2.GetSizeNoFee()
						txs, err = bc.selectTransactions(maxSize)
						Expect(err).To(BeNil())
					})

					It("should return 1 transactions", func() {
						Expect(txs).To(HaveLen(1))
					})

					Specify("container should contain 2 transactions since selected txs go back in the pool", func() {
						Expect(tp.Size()).To(Equal(int64(2)))
					})
				})
			})

			Context("test size validations", func() {
				BeforeEach(func() {
					tp = bc.txPool
					tx = core.NewTx(core.TxTypeBalance, 1, util.String(sender.Addr()), sender, "0.1", "0.001", time.Now().Unix())
					tx.Hash = tx.ComputeHash()
					tp.Put(tx)

					tx2 = core.NewTx(core.TxTypeBalance, 2, util.String(sender.Addr()), sender, "0.2", "0.001", time.Now().Unix())
					tx2.Hash = tx2.ComputeHash()
					tp.Put(tx2)

					tx3 = core.NewTx(core.TxTypeBalance, 3, util.String(sender.Addr()), sender, "0.3", "0.001", time.Now().Unix())
					tx3.Hash = tx3.ComputeHash()
					tp.Put(tx3)
				})

				It("should only include transactions up to the given max size", func() {
					maxSize := tx.GetSizeNoFee() + tx2.GetSizeNoFee()
					txs, err := bc.selectTransactions(maxSize)
					Expect(err).To(BeNil())
					Expect(txs).To(HaveLen(2))
				})

				It("should only include all transactions when max size exceeds pool size", func() {
					maxSize := tx.GetSizeNoFee() + tx2.GetSizeNoFee() + tx3.GetSizeNoFee() + 100
					txs, err := bc.selectTransactions(maxSize)
					Expect(err).To(BeNil())
					Expect(txs).To(HaveLen(3))
				})

				When("max size is too small", func() {
					It("should select nothing and put back all transactions back in the pool", func() {
						maxSize := int64(1)
						txs, err := bc.selectTransactions(maxSize)
						Expect(err).To(BeNil())
						Expect(txs).To(HaveLen(0))
						Expect(tp.Size()).To(Equal(int64(3)))
					})
				})
			})
		})
	})

})

var _ = Describe("UnitBlock", func() {

	var err error
	var bc *Blockchain
	var cfg *config.EngineConfig
	var db elldb.DB

	BeforeEach(func() {
		cfg, err = testutil.SetTestCfg()
		Expect(err).To(BeNil())

		db = elldb.NewDB(cfg.NetDataDir())
		err = db.Open(util.RandString(5))
		Expect(err).To(BeNil())

		bc = New(txpool.New(100), cfg, log)
		bc.SetDB(db)
		bc.SetCoinbase(crypto.NewKeyFromIntSeed(1234))
	})

	AfterEach(func() {
		db.Close()
		err = os.RemoveAll(cfg.DataDir())
		Expect(err).To(BeNil())
	})

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
			ch2 := NewChain("c2", db, cfg, log)
			bc.bestChain = ch
			Expect(bc.IsMainChain(ch2.ChainReader())).To(BeFalse())
		})

		It("should return true when the given chain is the main chain", func() {
			ch := NewChain("c1", db, cfg, log)
			bc.bestChain = ch
			Expect(bc.IsMainChain(ch.ChainReader())).To(BeTrue())
		})
	})

	Describe(".ChainReader", func() {
		It("should return a chain reader with same ID as the best chain", func() {
			ch := NewChain("c1", db, cfg, log)
			bc.addChain(ch)
			bc.bestChain = ch
			Expect(bc.ChainReader().GetID()).To(Equal(ch.id))
		})
	})

	Describe(".GetChainsReader", func() {
		It("should get chain readers for all known chains", func() {

			ch := NewChain("c1", db, cfg, log)
			bc.chains[ch.id] = ch
			ch2 := NewChain("c22", db, cfg, log)
			bc.chains[ch2.id] = ch2
			Expect(bc.chains).To(HaveLen(2))

			readers := bc.GetChainsReader()
			Expect(readers).To(HaveLen(2))
			expectedChains := []string{ch.id.String(), ch2.id.String()}
			Expect(expectedChains).To(ContainElement(ch.GetID().String()))
			Expect(expectedChains).To(ContainElement(ch2.GetID().String()))
		})
	})

	Describe(".addChain", func() {
		It("should add chain", func() {
			chain := NewChain("c1", db, cfg, log)
			Expect(err).To(BeNil())
			Expect(bc.chains).To(HaveLen(0))
			bc.addChain(chain)
			Expect(bc.chains).To(HaveLen(1))
		})
	})

	Describe(".removeChain", func() {

		var chain *Chain

		BeforeEach(func() {
			chain = NewChain("c1", db, cfg, log)
			Expect(err).To(BeNil())
			bc.addChain(chain)
			Expect(bc.chains).To(HaveLen(1))
		})

		It("should remove chain", func() {
			bc.removeChain(chain)
			Expect(bc.chains).To(HaveLen(0))
			Expect(bc.chains[chain.GetID()]).To(BeNil())
		})
	})

	Describe(".hasChain", func() {

		var chain *Chain

		BeforeEach(func() {
			chain = NewChain("c1", db, cfg, log)
			Expect(err).To(BeNil())
		})

		It("should return true if chain exists", func() {
			Expect(bc.hasChain(chain)).To(BeFalse())
			bc.addChain(chain)
			Expect(bc.hasChain(chain)).To(BeTrue())
		})
	})

})
