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
		chain, err = NewChain(chainID, store, cfg, log)
		Expect(err).To(BeNil())
		bc.addChain(chain)
		err = chain.init(testdata.TestBlock1)
		Expect(err).To(BeNil())
		genesisBlock, err = wire.BlockFromString(testdata.TestBlock1)
	})

	AfterEach(func() {
		db.Close()
	})

	AfterEach(func() {
		Expect(testutil.RemoveTestCfgDir()).To(BeNil())
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

})
