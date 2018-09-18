package blockchain

import (
	"os"

	. "github.com/ellcrys/elld/blockchain/testutil"
	"github.com/ellcrys/elld/blockchain/txpool"
	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/elldb"
	"github.com/ellcrys/elld/testutil"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/types/core/objects"
	"github.com/ellcrys/elld/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Blockchain.Up", func() {

	var err error
	var bc *Blockchain
	var cfg *config.EngineConfig
	var db elldb.DB
	var genesisBlock core.Block
	var genesisChain *Chain
	var sender, receiver *crypto.Key

	BeforeEach(func() {
		cfg, err = testutil.SetTestCfg()
		Expect(err).To(BeNil())

		db = elldb.NewDB(cfg.ConfigDir())
		err = db.Open(util.RandString(5))
		Expect(err).To(BeNil())

		sender = crypto.NewKeyFromIntSeed(1)
		receiver = crypto.NewKeyFromIntSeed(2)

		bc = New(txpool.New(100), cfg, log)
		bc.SetDB(db)
	})

	AfterEach(func() {
		db.Close()
		err = os.RemoveAll(cfg.ConfigDir())
		Expect(err).To(BeNil())
	})

	Describe(".Up", func() {

		It("should return error if db is not set", func() {
			bc.SetDB(nil)
			err = bc.Up()
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("db has not been initialized"))
		})

		When("well configured", func() {
			BeforeEach(func() {
				genesisBlock, err = LoadBlockFromFile("genesis-test.json")
				Expect(err).To(BeNil())
				bc.SetGenesisBlock(genesisBlock)
			})

			It("should return nil", func() {
				err = bc.Up()
				Expect(err).To(BeNil())
			})

			It("should load all chains", func() {
				c1 := NewChain("c_1", db, cfg, log)
				err = bc.saveChain(c1, "", 0)
				Expect(err).To(BeNil())
				err = c1.append(genesisBlock)
				Expect(err).To(BeNil())

				c2 := NewChain("c_2", db, cfg, log)
				err = bc.saveChain(c2, "", 0)
				Expect(err).To(BeNil())

				bc.SetGenesisBlock(genesisBlock)
				err = bc.Up()
				Expect(err).To(BeNil())

				Expect(bc.chains).To(HaveLen(2))
				Expect(bc.chains).To(HaveKey(c1.id))
				Expect(bc.chains).To(HaveKey(c2.id))
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

	Context("With a well initialized blockchain instance", func() {

		BeforeEach(func() {
			genesisBlock, err = LoadBlockFromFile("genesis-test.json")
			Expect(err).To(BeNil())
			bc.SetGenesisBlock(genesisBlock)
			err = bc.Up()
			Expect(err).To(BeNil())
			genesisChain = bc.bestChain
		})

		Describe(".findChainByLastBlockHash", func() {

			var b1 core.Block
			var chain2 *Chain

			BeforeEach(func() {
				chain2 = NewChain("chain2", db, cfg, log)
				bc.addChain(chain2)

				err := bc.CreateAccount(1, chain2, &objects.Account{
					Type:    objects.AccountTypeBalance,
					Address: util.String(sender.Addr()),
					Balance: "100",
				})
				Expect(err).To(BeNil())

				b1 = MakeBlock(bc, chain2, sender, receiver)
				err = chain2.append(b1)
				Expect(err).To(BeNil())
			})

			It("should return chain=nil, header=nil, err=nil if no block on the chain matches the hash", func() {
				result, chain, header, err := bc.findChainByBlockHash(util.Hash{1, 2, 3})
				Expect(err).To(BeNil())
				Expect(result).To(BeNil())
				Expect(header).To(BeNil())
				Expect(chain).To(BeNil())
			})

			Context("when the hash belongs to the highest block of a chain", func() {
				It("should return the chain", func() {
					_, chain, _, err := bc.findChainByBlockHash(b1.GetHash())
					Expect(err).To(BeNil())
					Expect(chain.GetID()).To(Equal(chain2.id))
				})

				Specify("header must match the header of the recently added block", func() {
					_, _, header, err := bc.findChainByBlockHash(b1.GetHash())
					Expect(err).To(BeNil())
					Expect(header.ComputeHash()).To(Equal(b1.GetHeader().ComputeHash()))
				})

				Specify("the returned block mush equal the block used for query", func() {
					result, _, _, err := bc.findChainByBlockHash(b1.GetHash())
					Expect(err).To(BeNil())
					Expect(b1.Bytes()).To(Equal(result.Bytes()))
				})
			})

			Context("when the hash belongs to a block that is not the highest block", func() {

				var b2 core.Block

				BeforeEach(func() {
					b2 = MakeBlock(bc, genesisChain, sender, receiver)
					err = genesisChain.append(b2)
					Expect(err).To(BeNil())
				})

				It("should return chain and header matching the header of block 1", func() {
					result, chain, tipHeader, err := bc.findChainByBlockHash(genesisBlock.GetHash())
					Expect(err).To(BeNil())
					Expect(genesisBlock.Bytes()).To(Equal(result.Bytes()))
					Expect(genesisBlock.GetNumber()).To(Equal(uint64(1)))
					Expect(chain.GetID()).To(Equal(chain.id))
					Expect(tipHeader.ComputeHash()).To(Equal(b2.GetHeader().ComputeHash()))
				})
			})
		})

		Describe(".loadChain", func() {
			var block core.Block
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
				parentBlock = MakeBlock(bc, genesisChain, sender, receiver)
				block = MakeBlockWithParentHash(bc, genesisChain, sender, receiver, parentBlock.GetHash())
				unknownParent = MakeBlockWithParentHash(bc, genesisChain, sender, receiver, util.StrToHash("unknown"))
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
	})
})

var _ = Describe("Blockchain Unit Test", func() {

	var err error
	var bc *Blockchain
	var cfg *config.EngineConfig
	var db elldb.DB

	BeforeEach(func() {
		cfg, err = testutil.SetTestCfg()
		Expect(err).To(BeNil())

		db = elldb.NewDB(cfg.ConfigDir())
		err = db.Open(util.RandString(5))
		Expect(err).To(BeNil())

		bc = New(txpool.New(100), cfg, log)
		bc.SetDB(db)
	})

	AfterEach(func() {
		db.Close()
		err = os.RemoveAll(cfg.ConfigDir())
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
