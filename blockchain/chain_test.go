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
	var chain *Chain
	var db database.DB
	var chainID = "main"
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
	})

	BeforeEach(func() {
		chain = NewChain(chainID, store, cfg, log)
		Expect(err).To(BeNil())
		err = chain.init(testdata.ChainDotJSON[0])
		Expect(err).To(BeNil())
		genesisBlock, _ = wire.BlockFromString(testdata.ChainDotJSON[0])
	})

	AfterEach(func() {
		db.Close()
		Expect(testutil.RemoveTestCfgDir()).To(BeNil())
	})

	Describe(".appendBlock", func() {

		var block, block2 *wire.Block

		BeforeEach(func() {
			block, _ = wire.BlockFromString(testdata.ChainDotJSON[1])
			block2, _ = wire.BlockFromString(testdata.ChainDotJSON[2])
		})

		It("should return err when the block's parent hash does not match the hash of the current tail block", func() {
			err = chain.appendBlock(block)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("unable to append block: parent hash does not match the hash of the current block"))
		})

		It("should return no error", func() {
			err = chain.appendBlock(block2)
			Expect(err).To(BeNil())
		})
	})

	Describe(".hashBlock", func() {

		var block *wire.Block

		BeforeEach(func() {
			block, err = wire.BlockFromString(testdata.ChainDotJSON[1])
			Expect(err).To(BeNil())
		})

		It("should return false if block does not exist in the chain", func() {
			exist, err := chain.hasBlock(block.GetHash())
			Expect(err).To(BeNil())
			Expect(exist).To(BeFalse())
		})

		It("should return true if block exist in the chain", func() {
			var r []*database.KVObject
			chain.store.Get([]byte("block"), &r)
			exist, err := chain.hasBlock(genesisBlock.GetHash())
			Expect(err).To(BeNil())
			Expect(exist).To(BeTrue())
		})
	})

	Describe(".getBlockHeaderByHash", func() {

		It("should return err if block was not found", func() {
			header, err := chain.getBlockHeaderByHash("unknown")
			Expect(err).To(Equal(common.ErrBlockNotFound))
			Expect(header).To(BeNil())
		})

		It("should successfully get block header by hash", func() {
			header, err := chain.getBlockHeaderByHash(genesisBlock.GetHash())
			Expect(err).To(BeNil())
			Expect(header).ToNot(BeNil())
		})
	})

	Describe(".getBlockByHash", func() {
		It("should return error if block is not found", func() {
			block, err := chain.getBlockByHash("unknown")
			Expect(err).To(Equal(common.ErrBlockNotFound))
			Expect(block).To(BeNil())
		})

		It("should successfully get block by hash", func() {
			block, err := chain.getBlockByHash(genesisBlock.GetHash())
			Expect(err).To(BeNil())
			Expect(block).ToNot(BeNil())
		})
	})

	// Describe(".getMatureTickets", func() {

	// 	BeforeEach(func() {
	// 		for _, b := range testdata.TestBlocks {
	// 			var block wire.Block
	// 			err := json.Unmarshal([]byte(b), &block)
	// 			Expect(err).To(BeNil())
	// 			err = store.PutBlock(chainID, &block)
	// 			Expect(err).To(BeNil())
	// 		}
	// 	})

	// 	It("", func() {
	// 		ticketTxs, err := chain.getMatureTickets(4)
	// 		Expect(err).To(BeNil())
	// 		fmt.Println(ticketTxs)
	// 	})
	// })
})
