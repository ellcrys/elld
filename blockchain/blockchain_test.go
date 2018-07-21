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
	var genesisBlock *wire.Block

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
		bc.SetStore(store)
	})

	BeforeEach(func() {
		chain = NewChain(chainID, store, cfg, log)
		Expect(err).To(BeNil())
		bc.addChain(chain)
		err = chain.init(testdata.BlockchainDotGoJSON[0])
		Expect(err).To(BeNil())
		genesisBlock, err = wire.BlockFromString(testdata.BlockchainDotGoJSON[0])
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
			bc.SetStore(store)
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
			bc.chains = nil
		})

		It("should return error if store is not set", func() {
			bc := New(cfg, log)
			err = bc.Up()
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("store not set"))
		})

		It("should assign new chain as the best chain if no chain is known", func() {
			err = bc.Up()
			Expect(err).To(BeNil())
			Expect(bc.bestChain).ToNot(BeNil())
			Expect(bc.bestChain.id).To(Equal(MainChainID))
		})

		It("should load all chains", func() {
			// bc.updateMeta(&common.BlockchainMeta{Chains: []string{"chain_2", "chain_3"}})
			bc.updateMeta(&common.BlockchainMeta{
				Chains: []*common.ChainInfo{
					&common.ChainInfo{ID: "chain_2"},
					&common.ChainInfo{ID: "chain_3"},
				},
			})
			err = bc.Up()
			Expect(err).To(BeNil())
			Expect(bc.chains).To(HaveLen(2))
			Expect(bc.chains[0].id).To(Equal("chain_2"))
			Expect(bc.chains[1].id).To(Equal("chain_3"))
		})
	})

	Context("Metadata", func() {

		var meta = common.BlockchainMeta{
			Chains: []*common.ChainInfo{
				&common.ChainInfo{ID: "chain_id"},
			},
		}

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

	Describe(".newChain", func() {

		It("should return error if block is nil", func() {
			_, err := bc.newChain(nil, nil)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("stale block cannot be nil"))
		})

		It("should return error if stale block parent is nil", func() {
			staleBlock, _ := wire.BlockFromString(testdata.BlockchainDotGoJSON[0])
			_, err := bc.newChain(staleBlock, nil)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("stale block parent cannot be nil"))
		})

		It("should return error if stale block and parent are not related", func() {
			staleBlockParent, err := wire.BlockFromString(testdata.BlockchainDotGoJSON[0])
			staleBlock, _ := wire.BlockFromString(testdata.BlockchainDotGoJSON[1])
			_, err = bc.newChain(staleBlock, staleBlockParent)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("stale block and parent are not related"))
		})

		It("should successfully return a new chain", func() {
			staleBlockParent, err := wire.BlockFromString(testdata.BlockchainDotGoJSON[1])
			staleBlock, _ := wire.BlockFromString(testdata.BlockchainDotGoJSON[2])
			chain, err := bc.newChain(staleBlock, staleBlockParent)
			Expect(err).To(BeNil())
			Expect(chain).ToNot(BeNil())
			Expect(chain.parentBlock).To(Equal(staleBlockParent))

			Describe("metadata must have the new chain stored in the list of known chains", func() {
				meta := bc.GetMeta()
				Expect(meta.Chains).To(HaveLen(1))
				Expect(meta.Chains[0].ID).To(Equal(chain.id))
			})
		})
	})

})
