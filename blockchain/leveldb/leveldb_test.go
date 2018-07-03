package leveldb

import (
	"encoding/json"

	"github.com/ellcrys/elld/blockchain/types"
	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/database"
	"github.com/ellcrys/elld/testutil"
	"github.com/ellcrys/elld/wire"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Leveldb", func() {

	var err error
	var db database.DB
	var cfg *config.EngineConfig
	var store *Store

	BeforeEach(func() {
		cfg, err = testutil.SetTestCfg()
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		Expect(testutil.RemoveTestCfgDir()).To(BeNil())
	})

	BeforeEach(func() {
		db = database.NewLevelDB(cfg.ConfigDir())
		err = db.Open("")
		Expect(err).To(BeNil())
	})

	BeforeEach(func() {
		store, err = New(db)
		Expect(err).To(BeNil())
	})

	Describe(".PutBlock", func() {

		var chainID = "main"
		var block = &wire.Block{
			Header: &wire.Header{Number: 1},
			Hash:   "hash",
		}

		It("should put block without error", func() {
			err = store.PutBlock(chainID, block)
			Expect(err).To(BeNil())
			result := store.db.GetByPrefix(database.MakePrefix([]string{"block", chainID, "number"}))
			Expect(result).To(HaveLen(1))

			var storedBlock wire.Block
			err = json.Unmarshal(result[0].Value, &storedBlock)
			Expect(err).To(BeNil())
			Expect(&storedBlock).To(Equal(block))
		})

		Context("on successful put", func() {

			BeforeEach(func() {
				err = store.PutBlock(chainID, block)
				Expect(err).To(BeNil())
				result := store.db.GetByPrefix(database.MakePrefix([]string{"block", chainID, "number"}))
				Expect(result).To(HaveLen(1))
			})

			It("should update metadata object", func() {
				result := store.db.GetByPrefix(database.MakePrefix([]string{"meta", chainID}))
				Expect(result).ToNot(BeEmpty())
				var meta types.Meta
				err := json.Unmarshal(result[0].Value, &meta)
				Expect(err).To(BeNil())
				Expect(meta.CurrentBlockNumber).To(Equal(block.Header.Number))
			})
		})

		It("should return nil and not add block when another block with same number exists", func() {
			err = store.PutBlock(chainID, block)
			Expect(err).To(BeNil())
			result := store.db.GetByPrefix(database.MakePrefix([]string{"block", chainID, "number"}))
			Expect(result).To(HaveLen(1))

			var storedBlock wire.Block
			err = json.Unmarshal(result[0].Value, &storedBlock)
			Expect(err).To(BeNil())
			Expect(&storedBlock).To(Equal(block))

			var block2 = &wire.Block{
				Header: &wire.Header{Number: 1},
				Hash:   "some_hash",
			}

			err = store.PutBlock(chainID, block2)
			Expect(err).To(BeNil())
			result = store.db.GetByPrefix(database.MakePrefix([]string{"block", chainID, "number"}))
			Expect(result).To(HaveLen(1))

			err = json.Unmarshal(result[0].Value, &storedBlock)
			Expect(err).To(BeNil())
			Expect(&storedBlock).ToNot(Equal(block2))
		})
	})

	Describe(".GetBlock", func() {

		var chainID = "main"
		var block = &wire.Block{
			Header: &wire.Header{Number: 1},
			Hash:   "hash",
		}

		BeforeEach(func() {
			err = store.PutBlock(chainID, block)
			Expect(err).To(BeNil())
			result := store.db.GetByPrefix(database.MakePrefix([]string{"block", chainID, "number"}))
			Expect(result).To(HaveLen(1))
		})

		It("should get block by number", func() {
			var storedBlock = &wire.Block{}
			err = store.GetBlock(chainID, block.Header.Number, storedBlock)
			Expect(err).To(BeNil())
			Expect(storedBlock).ToNot(BeNil())
		})

		It("should get block by hash", func() {
			var storedBlock = &wire.Block{}
			err = store.GetBlockByHash(chainID, block.Hash, storedBlock)
			Expect(err).To(BeNil())
			Expect(storedBlock).ToNot(BeNil())
		})

		It("with block hash; return error if block does not exist", func() {
			err = store.GetBlockByHash(chainID, "unknown", nil)
			Expect(err).ToNot(BeNil())
			Expect(err).To(Equal(types.ErrBlockNotFound))
		})

		It("with block number; return error if block does not exist", func() {
			err = store.GetBlock(chainID, 10000, nil)
			Expect(err).ToNot(BeNil())
			Expect(err).To(Equal(types.ErrBlockNotFound))
		})

		It("should return the block with the hightest number if 0 is passed", func() {
			var block2 = &wire.Block{
				Header: &wire.Header{Number: 2},
				Hash:   "hash",
			}
			err = store.PutBlock(chainID, block2)
			Expect(err).To(BeNil())

			var storedBlock = &wire.Block{}
			err = store.GetBlock(chainID, 0, storedBlock)
			Expect(err).To(BeNil())
			Expect(storedBlock).ToNot(BeNil())
			Expect(storedBlock).To(Equal(block2))
		})
	})

	Describe("GetBlockHeader", func() {

		var chainID = "main"
		var block = &wire.Block{
			Header: &wire.Header{Number: 1},
			Hash:   "hash",
		}

		BeforeEach(func() {
			err = store.PutBlock(chainID, block)
			Expect(err).To(BeNil())
			result := store.db.GetByPrefix(database.MakePrefix([]string{"block", chainID, "number"}))
			Expect(result).To(HaveLen(1))
		})

		It("should get block by number", func() {
			var storedBlockHeader = &wire.Header{}
			err = store.GetBlockHeader(chainID, block.Header.Number, storedBlockHeader)
			Expect(err).To(BeNil())
			Expect(storedBlockHeader).ToNot(BeNil())
			Expect(storedBlockHeader).To(Equal(block.Header))
		})
	})

	Describe(".UpdateMetadata", func() {

		var chainID = "main"
		var meta types.Meta

		It("should successfully update meta", func() {
			meta = types.Meta{CurrentBlockNumber: 10000}
			err := store.UpdateMetadata(chainID, &meta)
			Expect(err).To(BeNil())
		})
	})

	Describe(".GetMetadata", func() {

		var chainID = "main"
		var meta types.Meta

		BeforeEach(func() {
			meta = types.Meta{CurrentBlockNumber: 10000}
			err := store.UpdateMetadata(chainID, &meta)
			Expect(err).To(BeNil())
		})

		It("should successfully get meta", func() {
			var result types.Meta
			err := store.GetMetadata(chainID, &result)
			Expect(err).To(BeNil())
			Expect(result.CurrentBlockNumber).To(Equal(meta.CurrentBlockNumber))
		})
	})
})
