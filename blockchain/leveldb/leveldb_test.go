package leveldb

import (
	"encoding/json"

	"github.com/ellcrys/elld/blockchain/common"
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

	Describe(".PutBlockWithTx", func() {
		var chainID = "main"
		var block = &wire.Block{
			Header: &wire.Header{Number: 1},
			Hash:   "hash",
		}

		It("should put block without error", func() {
			tx := store.NewTx()
			err = store.PutBlockWithTx(tx, chainID, block)
			Expect(err).To(BeNil())
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
			err = store.GetBlockByHash(chainID, block.GetHash(), storedBlock)
			Expect(err).To(BeNil())
			Expect(storedBlock).ToNot(BeNil())
		})

		It("with block hash; return error if block does not exist", func() {
			err = store.GetBlockByHash(chainID, "unknown", nil)
			Expect(err).ToNot(BeNil())
			Expect(err).To(Equal(common.ErrBlockNotFound))
		})

		It("with block number; return error if block does not exist", func() {
			err = store.GetBlock(chainID, 10000, nil)
			Expect(err).ToNot(BeNil())
			Expect(err).To(Equal(common.ErrBlockNotFound))
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

	Describe(".GetBlockHeaderByHash", func() {
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

		It("should get block by hash", func() {
			var storedBlockHeader = &wire.Header{}
			err = store.GetBlockHeaderByHash(chainID, block.GetHash(), storedBlockHeader)
			Expect(err).To(BeNil())
			Expect(storedBlockHeader).ToNot(BeNil())
			Expect(storedBlockHeader).To(Equal(block.Header))
		})
	})

	Describe(".Put", func() {
		It("should successfully store object", func() {
			key := database.MakeKey([]byte("my_key"), []string{"block", "account"})
			err = store.Put(key, []byte("stuff"))
			Expect(err).To(BeNil())
		})
	})

	Describe(".PutTx", func() {

		It("should successfully store object", func() {
			tx := store.NewTx()
			key := database.MakeKey([]byte("my_key"), []string{"block", "account"})
			err = store.PutWithTx(tx, key, []byte("stuff"))
			Expect(err).To(BeNil())
		})
	})

	Describe(".Get", func() {

		It("should successfully get object by prefix", func() {
			key := database.MakeKey([]byte("my_key"), []string{"an_obj", "account"})
			err = store.Put(key, []byte("stuff"))
			Expect(err).To(BeNil())

			var result = []*database.KVObject{}
			store.Get([]byte("an_obj"), &result)
			Expect(result).To(HaveLen(1))

			result = []*database.KVObject{}
			store.Get(database.MakePrefix([]string{"an_obj", "account"}), &result)
			Expect(result).To(HaveLen(1))
		})

		It("should successfully get object by key", func() {
			key := database.MakeKey([]byte("my_key"), []string{"block", "account"})
			err = store.Put(key, []byte("stuff"))
			Expect(err).To(BeNil())

			var result = []*database.KVObject{}
			store.Get(key, &result)
			Expect(result).To(HaveLen(1))
		})
	})

	Describe(".GetFirstOrLast", func() {

		var key, key2 []byte

		BeforeEach(func() {
			key = database.MakeKey([]byte("my_key"), []string{"an_obj", "account", "1"})
			err = store.Put(key, []byte("stuff"))
			Expect(err).To(BeNil())

			key2 = database.MakeKey([]byte("my_key"), []string{"an_obj", "account", "2"})
			err = store.Put(key2, []byte("stuff2"))
			Expect(err).To(BeNil())
		})

		It("should return the first object when first arg. is true", func() {
			var result = &database.KVObject{}
			store.GetFirstOrLast(true, []byte("an_obj"), result)
			Expect(result).ToNot(BeNil())
			Expect(result.GetKey()).To(Equal(key))
		})

		It("should return the last object when last arg. is false", func() {
			var result = &database.KVObject{}
			store.GetFirstOrLast(false, []byte("an_obj"), result)
			Expect(result).ToNot(BeNil())
			Expect(result.GetKey()).To(Equal(key2))
		})
	})
})
