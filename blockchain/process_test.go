package blockchain

import (
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
	var chainID = "chain1"
	var chain *Chain
	var bc *Blockchain

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
	})

	BeforeEach(func() {
		bc = New(cfg, log)
		bc.SetStore(store)
		chain, err = NewChain(chainID, store, cfg, log)
		Expect(err).To(BeNil())
		bc.bestChain = chain
	})

	AfterEach(func() {
		db.Close()
	})

	AfterEach(func() {
		Expect(testutil.RemoveTestCfgDir()).To(BeNil())
	})

	Describe(".addOp", func() {

		var curOps = []types.Transition{
			&types.OpNewAccountBalance{OpBase: &types.OpBase{Addr: "addr1"}, Amount: "10"},
		}

		It("should add an additional op successfully", func() {
			op := &types.OpNewAccountBalance{OpBase: &types.OpBase{Addr: "addr2"}, Amount: "10"}
			newOps := addOp(curOps, op)
			Expect(newOps).To(HaveLen(2))
		})

		It("should replace any equal op found", func() {
			op := &types.OpNewAccountBalance{OpBase: &types.OpBase{Addr: "addr1"}, Amount: "30"}
			newOps := addOp(curOps, op)
			Expect(newOps).To(HaveLen(1))
			Expect(newOps[0]).To(Equal(op))
			Expect(newOps[0]).ToNot(Equal(curOps[0]))
		})
	})

	Describe(".processTransactions", func() {

		var block *wire.Block

		BeforeEach(func() {
			block, err = wire.BlockFromString(testdata.TestBlock1)
			Expect(err).To(BeNil())
		})

		It("should return error if sender does not exist in the best chain", func() {
			err = bc.processTransactions(block.Transactions)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("failed to get sender's account: account not found"))
		})

		// BeforeEach(func() {
		// 	block, err = wire.BlockFromString(testdata.TestBlock1)
		// 	Expect(err).To(BeNil())
		// })

		// It("", func() {

		// 	// err = bc.ProcessBlock(block)
		// 	// Expect(err).To(BeNil())
		// })
	})
})
