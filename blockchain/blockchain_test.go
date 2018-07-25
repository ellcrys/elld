package blockchain

import (
	"github.com/ellcrys/elld/blockchain/common"
	"github.com/ellcrys/elld/blockchain/leveldb"
	"github.com/ellcrys/elld/blockchain/testdata"
	"github.com/ellcrys/elld/database"
	"github.com/ellcrys/elld/testutil"
	"github.com/ellcrys/elld/wire"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Blockchain", func() {

	var err error
	var store common.Store
	var db database.DB
	var bc *Blockchain
	var chainID = "chain1"
	var chain *Chain

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
		bc = New(cfg, log)
		bc.setStore(store)
	})

	BeforeEach(func() {
		chain = NewChain(chainID, store, cfg, log)
		Expect(err).To(BeNil())
		bc.addChain(chain)

		block, err := wire.BlockFromString(testdata.BlockchainDotGoJSON[0])
		Expect(err).To(BeNil())

		err = chain.append(block)
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		db.Close()
	})

	AfterEach(func() {
		Expect(testutil.RemoveTestCfgDir()).To(BeNil())
	})

	Describe(".SetStore", func() {
		It("should set store", func() {
			bc := New(cfg, log)
			bc.setStore(store)
			Expect(bc.store).ToNot(BeNil())
		})
	})

	Describe(".addChain", func() {
		It("should add chain", func() {
			chain := NewChain("chain_id", store, cfg, log)
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
			chain = NewChain("chain_id", store, cfg, log)
			Expect(err).To(BeNil())
			err = bc.addChain(chain)
			Expect(err).To(BeNil())
			Expect(bc.chains).To(HaveLen(2))
		})

		It("should remove chain", func() {
			bc.removeChain(chain)
			Expect(bc.chains).To(HaveLen(1))
			Expect(bc.chains[chain.id]).To(BeNil())
		})
	})

	Describe(".hasChain", func() {

		var chain *Chain

		BeforeEach(func() {
			chain = NewChain("chain_id", store, cfg, log)
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
		})

		It("should return error if store is not set", func() {
			bc := New(cfg, log)
			err = bc.Up()
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("store has not been initialized"))
		})

		It("should assign new chain as the best chain if no chain is known", func() {
			err = bc.Up()
			Expect(err).To(BeNil())
			Expect(bc.bestChain).ToNot(BeNil())
		})

		It("should load all chains", func() {
			c1 := NewChain("c1", store, cfg, log)
			c2 := NewChain("c2", store, cfg, log)

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

	Describe(".HasBlock", func() {

		var block *wire.Block

		BeforeEach(func() {
			block, err = wire.BlockFromString(testdata.BlockchainDotGoJSON[1])
			Expect(err).To(BeNil())
		})

		It("should return false when block does not exist in any known chain", func() {
			has, err := bc.HaveBlock(block.GetHash())
			Expect(err).To(BeNil())
			Expect(has).To(BeFalse())
		})

		It("should return true of block exists in a chain", func() {
			chain2 := NewChain("chain2", store, cfg, log)
			Expect(err).To(BeNil())
			err = chain2.init(testdata.BlockchainDotGoJSON[1])
			Expect(err).To(BeNil())

			err = bc.addChain(chain2)
			Expect(err).To(BeNil())
			err = chain2.store.PutBlock(chain2.id, block)
			Expect(err).To(BeNil())

			has, err := bc.HaveBlock(block.GetHash())
			Expect(err).To(BeNil())
			Expect(has).To(BeTrue())
		})
	})

	Describe(".findChainByLastBlockHash", func() {

		var block *wire.Block
		var chain2 *Chain

		BeforeEach(func() {
			block, err = wire.BlockFromString(testdata.BlockchainDotGoJSON[1])
			Expect(err).To(BeNil())

			chain2 = NewChain("chain2", store, cfg, log)
			Expect(err).To(BeNil())

			err = chain2.init(testdata.BlockchainDotGoJSON[1])
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
				_block, chain, header, err := bc.findBlockChainByHash(block.GetHash())
				Expect(err).To(BeNil())
				Expect(block).To(Equal(_block))
				Expect(chain.id).To(Equal(chain2.id))
				Expect(header.ComputeHash()).To(Equal(block.Header.ComputeHash()))
			})
		})

		Context("when the hash maps to a block that is not the highest block", func() {

			var block2 *wire.Block

			BeforeEach(func() {
				block2, err = wire.BlockFromString(testdata.BlockchainDotGoJSON[1])
				Expect(err).To(BeNil())

				err = chain2.store.PutBlock(chain2.id, block2)
				Expect(err).To(BeNil())
			})

			It("should return chain (not chain2) and header matching the header of the recently aded block", func() {
				_block, chain, tipHeader, err := bc.findBlockChainByHash(block.GetHash())
				Expect(err).To(BeNil())
				Expect(block).To(Equal(_block))
				Expect(chain.id).To(Equal(chain2.id))
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
			chain = NewChain("chain_2", store, cfg, log)
			subChain = NewChain("sub_chain", store, cfg, log)

			block, err = wire.BlockFromString(testdata.LoadChainData[0])
			Expect(err).To(BeNil())

			err := bc.saveChain(chain, "", 0)
			Expect(err).To(BeNil())
			err = chain.append(block)
			Expect(err).To(BeNil())

			err = bc.saveChain(subChain, chain.id, block.GetNumber())
			Expect(err).To(BeNil())
		})

		It("should return error when only ParentBlockNumber is set but ParentChainID is unset", func() {
			err = bc.loadChain(&common.ChainInfo{ID: chain.id, ParentBlockNumber: 1})
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("chain load failed: parent chain id and parent block id are both required"))
		})

		It("should return error when only ParentChainID is set but ParentBlockNumber is unset", func() {
			err = bc.loadChain(&common.ChainInfo{ID: chain.id, ParentChainID: "some_id"})
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("chain load failed: parent chain id and parent block id are both required"))
		})

		It("should return error when parent block does not exist", func() {
			err = bc.loadChain(&common.ChainInfo{ID: chain.id, ParentChainID: "some_id", ParentBlockNumber: 100})
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("chain load failed: parent block {100} of chain {chain_2} not found"))
		})

		It("should successfully load chain with no parent into the cache", func() {
			err = bc.loadChain(&common.ChainInfo{ID: chain.id})
			Expect(err).To(BeNil())
			Expect(bc.chains).To(HaveLen(2))
			Expect(bc.chains[chain.id]).ToNot(BeNil())
		})

		It("should successfully load chain into the cache with parent block and chain info set", func() {
			err = bc.loadChain(&common.ChainInfo{ID: subChain.id, ParentChainID: chain.id, ParentBlockNumber: block.GetNumber()})
			Expect(err).To(BeNil())
			Expect(bc.chains).To(HaveLen(2))
			Expect(bc.chains[subChain.id]).ToNot(BeNil())
		})
	})

	Describe(".findChainInfo", func() {

		var chain *Chain

		BeforeEach(func() {
			chain = NewChain("chain_a", store, cfg, log)
			err := bc.saveChain(chain, "", 0)
			Expect(err).To(BeNil())
		})

		It("should find chain with id = chain_a", func() {
			chInfo, err := bc.findChainInfo("chain_a")
			Expect(err).To(BeNil())
			Expect(chInfo.ID).To(Equal(chain.id))
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
			initialBlock, _ := wire.BlockFromString(testdata.BlockchainDotGoJSON[0])
			_, err := bc.newChain(nil, initialBlock, nil, nil)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("initial block parent cannot be nil"))
		})

		It("should return error if initial block and parent are not related", func() {
			initialBlockParent, err := wire.BlockFromString(testdata.BlockchainDotGoJSON[0])
			initialBlock, _ := wire.BlockFromString(testdata.BlockchainDotGoJSON[1])
			_, err = bc.newChain(nil, initialBlock, initialBlockParent, nil)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("initial block and parent are not related"))
		})

		It("should successfully return a new chain", func() {
			initialBlockParent, err := wire.BlockFromString(testdata.BlockchainDotGoJSON[1])
			initialBlock, _ := wire.BlockFromString(testdata.BlockchainDotGoJSON[2])

			tx, _ := chain.store.NewTx()
			chain, err := bc.newChain(tx, initialBlock, initialBlockParent, nil)
			Expect(err).To(BeNil())
			Expect(chain).ToNot(BeNil())
			Expect(chain.parentBlock).To(Equal(initialBlockParent))
			tx.Commit()
		})
	})

	Describe(".chooseBestChain", func() {

		// Context("with one highest block", func() {

		// 	var chainA, chainB *Chain

		// 	BeforeEach(func() {
		// 		bc.chains = make(map[string]*Chain)
		// 	})

		// 	BeforeEach(func() {
		// 		chainA = NewChain("chain_a", store, cfg, log)
		// 		err := bc.saveChain(chainA, "", 0)
		// 		Expect(err).To(BeNil())
		// 		block, err := wire.BlockFromString(testdata.ChooseBestChainData[0])
		// 		err = chainA.append(block)
		// 		Expect(err).To(BeNil())
		// 	})

		// 	BeforeEach(func() {
		// 		chainB = NewChain("chain_b", store, cfg, log)
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
				chainA = NewChain("chain_a", store, cfg, log)
				err := bc.saveChain(chainA, "", 0)
				Expect(err).To(BeNil())
				block, err := wire.BlockFromString(testdata.ChooseBestChainData[0])
				err = chainA.append(block)
				Expect(err).To(BeNil())
			})

			BeforeEach(func() {
				chainB = NewChain("chain_b", store, cfg, log)
				err := bc.saveChain(chainB, "", 0)
				Expect(err).To(BeNil())
				block, err := wire.BlockFromString(testdata.ChooseBestChainData[0])
				err = chainB.append(block)
				Expect(err).To(BeNil())
			})

			BeforeEach(func() {
				chainC = NewChain("chain_c", store, cfg, log)
				err := bc.saveChain(chainC, "", 0)
				Expect(err).To(BeNil())
				block, err := wire.BlockFromString(testdata.ChooseBestChainData[0])
				block2, err := wire.BlockFromString(testdata.ChooseBestChainData[1])
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
