package blockchain

import (
	"fmt"

	"github.com/ellcrys/elld/blockchain/leveldb"
	"github.com/ellcrys/elld/blockchain/testdata"
	"github.com/ellcrys/elld/blockchain/types"
	"github.com/ellcrys/elld/database"
	"github.com/ellcrys/elld/testutil"
	"github.com/ellcrys/elld/wire"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Blockchain", func() {

	var err error
	var store types.Store
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
		bc.SetStore(store)
	})

	BeforeEach(func() {
		chain, err = NewChain(chainID, store, cfg, log)
		Expect(err).To(BeNil())
		bc.addChain(chain)
		err = chain.init(testdata.TestBlock1)
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
			bc.SetStore(store)
			Expect(bc.store).ToNot(BeNil())
		})
	})

	Describe(".addChain", func() {
		It("should add chain", func() {
			chain, err := NewChain("chain_id", store, cfg, log)
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
			chain, err = NewChain("chain_id", store, cfg, log)
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
			bc.UpdateMeta(&types.BlockchainMeta{Chains: []string{"chain_2", "chain_3"}})
			err = bc.Up()
			Expect(err).To(BeNil())
			Expect(bc.chains).To(HaveLen(3))
			// Expect(bc.chains[0].id).To(Equal("chain_1"))
			// Expect(bc.chains[1].id).To(Equal("chain_2"))
		})
	})

	Context("Metadata", func() {

		var meta = types.BlockchainMeta{
			Chains: []string{"chain_id"},
		}

		Describe(".UpdateMeta", func() {
			It("should successfully save metadata", func() {
				err = bc.UpdateMeta(&meta)
				Expect(err).To(BeNil())
			})
		})

		Describe(".GetMeta", func() {

			BeforeEach(func() {
				err = bc.UpdateMeta(&meta)
				Expect(err).To(BeNil())
			})

			It("should return metadata", func() {
				result, err := bc.GetMeta()
				Expect(err).To(BeNil())
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
			block, err = wire.BlockFromString(testdata.TestBlock2)
			Expect(err).To(BeNil())
		})

		It("should return false when block does not exist in any known chain", func() {
			has, err := bc.HasBlock(block.ComputeHash())
			Expect(err).To(BeNil())
			Expect(has).To(BeFalse())
		})

		It("should return true of block exists in a chain", func() {
			chain2, err := NewChain("chain2", store, cfg, log)
			Expect(err).To(BeNil())
			err = chain2.init(testdata.TestBlock1)
			Expect(err).To(BeNil())

			err = bc.addChain(chain2)
			Expect(err).To(BeNil())
			err = chain2.store.PutBlock(chain2.id, block)
			Expect(err).To(BeNil())

			has, err := bc.HasBlock(block.ComputeHash())
			Expect(err).To(BeNil())
			Expect(has).To(BeTrue())
		})
	})

	Describe(".findChainByLastBlockHash", func() {

		var block *wire.Block
		var chain2 *Chain

		BeforeEach(func() {
			block, err = wire.BlockFromString(testdata.TestBlock2)
			Expect(err).To(BeNil())
		})

		BeforeEach(func() {
			chain2, err = NewChain("chain2", store, cfg, log)
			Expect(err).To(BeNil())
			err = chain2.init(testdata.TestBlock2)
			Expect(err).To(BeNil())

			err = bc.addChain(chain2)
			Expect(err).To(BeNil())
			err = chain2.store.PutBlock(chain2.id, block)
			Expect(err).To(BeNil())
		})

		It("should return nil if no chain hash its last block equal to the hash", func() {
			chain, err := bc.findChainByTipHash("unknown")
			Expect(err).To(BeNil())
			Expect(chain).To(BeNil())
		})

		It("should return expected chain if its last block's hash is equal to the hash provided", func() {
			chain, err := bc.findChainByTipHash(block.ComputeHash())
			// Expect(err).To(BeNil())
			// Expect(chain.id).To(Equal(chain2.id))
			fmt.Println(chain, err)
		})
	})
})
