package store

import (
	"os"

	"github.com/ellcrys/elld/blockchain/common"
	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/elldb"
	"github.com/ellcrys/elld/testutil"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/types/core/objects"
	"github.com/ellcrys/elld/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Leveldb", func() {

	var err error
	var db elldb.DB
	var cfg *config.EngineConfig
	var store *ChainStore
	var chainID = util.String("main")

	BeforeEach(func() {
		cfg, err = testutil.SetTestCfg()
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		err = os.RemoveAll(cfg.ConfigDir())
		Expect(err).To(BeNil())
	})

	BeforeEach(func() {
		db = elldb.NewDB(cfg.ConfigDir())
		err = db.Open(util.RandString(5))
		Expect(err).To(BeNil())
	})

	BeforeEach(func() {
		store = New(db, chainID)
		Expect(err).To(BeNil())
	})

	Describe(".PutTransactions", func() {

		var txs = []core.Transaction{
			&objects.Transaction{To: "to_addr", From: "from_addr", Hash: util.StrToHash("hash1")},
			&objects.Transaction{To: "to_addr_2", From: "from_addr_2", Hash: util.StrToHash("hash2")},
		}

		It("should successfully put transactions", func() {
			err = store.PutTransactions(txs, 211)
			Expect(err).To(BeNil())

			r := store.db.GetByPrefix(common.MakeTxQueryKey(store.chainID.Bytes(), txs[0].GetHash().Bytes()))
			var tx objects.Transaction
			r[0].Scan(&tx)
			Expect(&tx).To(Equal(txs[0]))

			r = store.db.GetByPrefix(common.MakeTxQueryKey(store.chainID.Bytes(), txs[1].GetHash().Bytes()))
			r[0].Scan(&tx)
			Expect(&tx).To(Equal(txs[1]))
		})
	})

	Describe(".GetTransaction", func() {

		var txs = []core.Transaction{
			&objects.Transaction{To: "to_addr", From: "from_addr", Hash: util.StrToHash("hash1")},
			&objects.Transaction{To: "to_addr_2", From: "from_addr_2", Hash: util.StrToHash("hash2")},
		}

		It("should successfully put transactions", func() {
			err = store.PutTransactions(txs, 211)
			Expect(err).To(BeNil())

			tx := store.GetTransaction(txs[0].GetHash())
			Expect(tx).ToNot(BeNil())
			Expect(tx).To(Equal(txs[0]))
			tx = store.GetTransaction(txs[1].GetHash())
			Expect(tx).ToNot(BeNil())
			Expect(tx).To(Equal(txs[1]))
		})
	})

	Describe(".Delete", func() {
		var txs = []core.Transaction{
			&objects.Transaction{To: "to_addr", From: "from_addr", Hash: util.StrToHash("hash1")},
		}

		It("should successfully delete", func() {
			err = store.PutTransactions(txs, 211)
			Expect(err).To(BeNil())

			err := store.Delete(common.MakeTxQueryKey(store.chainID.Bytes(), txs[0].GetHash().Bytes()))
			Expect(err).To(BeNil())

			tx := store.GetTransaction(txs[0].GetHash())
			Expect(tx).To(BeNil())
		})
	})

	Describe(".CreateAccount", func() {
		It("should successfully create an account", func() {
			var acct = &objects.Account{Type: objects.AccountTypeBalance, Address: "addr"}
			err = store.CreateAccount(1, acct)
			Expect(err).To(BeNil())

			r := store.db.GetByPrefix(common.MakeAccountKey(1, store.chainID.Bytes(), []byte("addr")))
			var found objects.Account
			r[0].Scan(&found)
			Expect(&found).To(Equal(acct))
		})
	})

	Describe(".GetAccount", func() {
		Context("no existing account in store", func() {
			It("should return the only account", func() {
				var acct = &objects.Account{Type: objects.AccountTypeBalance, Address: "addr"}
				err = store.CreateAccount(1, acct)
				Expect(err).To(BeNil())

				found, err := store.GetAccount(util.String(acct.Address))
				Expect(err).To(BeNil())
				Expect(found).To(Equal(acct))
			})
		})

		Context("with multiple account of same address", func() {
			It("should contain account with the highest block number", func() {
				var acct = &objects.Account{Type: objects.AccountTypeBalance, Address: "addr"}
				err = store.CreateAccount(1, acct)
				Expect(err).To(BeNil())

				var acct2 = &objects.Account{Type: objects.AccountTypeBalance, Address: "addr2"}
				err = store.CreateAccount(2, acct2)
				Expect(err).To(BeNil())

				found, err := store.GetAccount((util.String(acct2.Address)))
				Expect(err).To(BeNil())
				Expect(found).To(Equal(acct2))
			})
		})
	})

	Describe(".PutBlock", func() {

		var block *objects.Block

		BeforeEach(func() {
			block = &objects.Block{
				Header: &objects.Header{Number: 1},
				Hash:   util.StrToHash("hash"),
				Sig:    []byte("stuff"),
			}
		})

		It("should put block without error", func() {
			err = store.PutBlock(block)
			Expect(err).To(BeNil())
			result := store.db.GetByPrefix(common.MakeBlocksQueryKey(chainID.Bytes()))
			Expect(result).To(HaveLen(1))

			var storedBlock objects.Block
			err = result[0].Scan(&storedBlock)
			Expect(err).To(BeNil())
			Expect(&storedBlock).To(Equal(block))
		})

		It("should return nil and not add block when another block with same number exists", func() {
			err = store.PutBlock(block)
			Expect(err).To(BeNil())
			result := store.db.GetByPrefix(common.MakeBlocksQueryKey(chainID.Bytes()))
			Expect(result).To(HaveLen(1))

			var storedBlock objects.Block
			err = result[0].Scan(&storedBlock)
			Expect(err).To(BeNil())
			Expect(&storedBlock).To(Equal(block))

			var block2 = &objects.Block{
				Header: &objects.Header{Number: 1},
				Hash:   util.StrToHash("some_hash"),
			}

			err = store.PutBlock(block2)
			Expect(err).To(BeNil())
			result = store.db.GetByPrefix(common.MakeBlocksQueryKey(chainID.Bytes()))
			Expect(result).To(HaveLen(1))

			err = result[0].Scan(&storedBlock)
			Expect(err).To(BeNil())
			Expect(&storedBlock).ToNot(Equal(block2))
		})
	})

	Describe(".Current", func() {
		When("two blocks are in the chain", func() {
			var block = &objects.Block{Header: &objects.Header{Number: 1}, Hash: util.StrToHash("hash")}
			var block2 = &objects.Block{Header: &objects.Header{Number: 2}, Hash: util.StrToHash("hash2")}

			BeforeEach(func() {
				err = store.PutBlock(block)
				Expect(err).To(BeNil())
				err = store.PutBlock(block2)
				Expect(err).To(BeNil())
			})

			It("should return most recently added block", func() {
				cb, err := store.Current()
				Expect(err).To(BeNil())
				Expect(cb).To(Equal(block2))
			})
		})

		When("when no block is in the chain", func() {
			It("should return ErrBlockNotFound", func() {
				_, err := store.Current()
				Expect(err).ToNot(BeNil())
				Expect(err).To(Equal(core.ErrBlockNotFound))
			})
		})
	})

	Describe(".GetBlockByNumberAndHash", func() {

		var block = &objects.Block{
			Header: &objects.Header{Number: 100},
			Hash:   util.StrToHash("hash"),
		}

		BeforeEach(func() {
			err = store.PutBlock(block)
			Expect(err).To(BeNil())
			result := store.db.GetByPrefix(common.MakeBlocksQueryKey(chainID.Bytes()))
			Expect(result).To(HaveLen(1))
		})

		It("should return ErrBlockNotFound if block does not exist", func() {
			_, err := store.GetBlockByNumberAndHash(1, block.Hash)
			Expect(err).ToNot(BeNil())
			Expect(err).To(Equal(core.ErrBlockNotFound))
		})

		It("should return ErrBlockNotFound if block with number exist but hash does not match", func() {
			_, err := store.GetBlockByNumberAndHash(block.GetNumber(), util.StrToHash("invalid_hash"))
			Expect(err).ToNot(BeNil())
			Expect(err).To(Equal(core.ErrBlockNotFound))
		})

		It("should successfully get block", func() {
			result, err := store.GetBlockByNumberAndHash(block.GetNumber(), block.Hash)
			Expect(err).To(BeNil())
			Expect(result).To(Equal(block))
		})
	})

	Describe(".GetBlock", func() {

		var block = &objects.Block{
			Header: &objects.Header{Number: 1},
			Hash:   util.StrToHash("hash"),
		}

		BeforeEach(func() {
			err = store.PutBlock(block)
			Expect(err).To(BeNil())
			result := store.db.GetByPrefix(common.MakeBlocksQueryKey(chainID.Bytes()))
			Expect(result).To(HaveLen(1))
		})

		It("should get block by number", func() {
			storedBlock, err := store.GetBlock(block.Header.GetNumber())
			Expect(err).To(BeNil())
			Expect(storedBlock).ToNot(BeNil())
		})

		It("should get block by hash", func() {
			storedBlock, err := store.GetBlockByHash(block.GetHash())
			Expect(err).To(BeNil())
			Expect(storedBlock).ToNot(BeNil())
		})

		It("with block hash; return error if block does not exist", func() {
			b, err := store.GetBlockByHash(util.Hash{1, 3, 4})
			Expect(err).ToNot(BeNil())
			Expect(err).To(Equal(core.ErrBlockNotFound))
			Expect(b).To(BeNil())
		})

		It("with block number; return error if block does not exist", func() {
			_, err = store.GetBlock(10000)
			Expect(err).ToNot(BeNil())
			Expect(err).To(Equal(core.ErrBlockNotFound))
		})

		It("should return the block with the hightest number if 0 is passed", func() {
			var block2 = &objects.Block{
				Header: &objects.Header{Number: 2},
				Hash:   util.StrToHash("hash"),
			}

			err = store.PutBlock(block2)
			Expect(err).To(BeNil())

			storedBlock, err := store.GetBlock(0)
			Expect(err).To(BeNil())
			Expect(storedBlock).ToNot(BeNil())
			Expect(storedBlock).To(Equal(block2))
		})
	})

	Describe("GetBlockHeader", func() {

		var block = &objects.Block{
			Header: &objects.Header{Number: 1},
			Hash:   util.StrToHash("hash"),
		}

		BeforeEach(func() {
			err = store.PutBlock(block)
			Expect(err).To(BeNil())
			result := store.db.GetByPrefix(common.MakeBlocksQueryKey(chainID.Bytes()))
			Expect(result).To(HaveLen(1))
		})

		It("should get block header by number", func() {
			storedBlockHeader, err := store.GetHeader(block.Header.GetNumber())
			Expect(err).To(BeNil())
			Expect(storedBlockHeader).ToNot(BeNil())
			Expect(storedBlockHeader).To(Equal(block.Header))
		})
	})

	Describe(".GetBlockHeaderByHash", func() {
		var block = &objects.Block{
			Header: &objects.Header{Number: 1},
			Hash:   util.StrToHash("hash"),
		}

		BeforeEach(func() {
			err = store.PutBlock(block)
			Expect(err).To(BeNil())
			result := store.db.GetByPrefix(common.MakeBlocksQueryKey(chainID.Bytes()))
			Expect(result).To(HaveLen(1))
		})

		It("should get block by hash", func() {
			storedBlockHeader, err := store.GetHeaderByHash(block.GetHash())
			Expect(err).To(BeNil())
			Expect(storedBlockHeader).ToNot(BeNil())
			Expect(storedBlockHeader).To(Equal(block.Header))
		})
	})

	Describe(".put", func() {
		It("should successfully store object", func() {
			key := elldb.MakeKey([]byte("my_key"), []byte("block"), []byte("account"))
			err = store.put(key, []byte("stuff"))
			Expect(err).To(BeNil())
		})
	})

	Describe(".get", func() {
		It("should successfully get object by prefix", func() {
			key := elldb.MakeKey([]byte("my_key"), []byte("an_obj"), []byte("account"))
			err = store.put(key, []byte("stuff"))
			Expect(err).To(BeNil())

			var result = []*elldb.KVObject{}
			store.get([]byte("an_obj"), &result)
			Expect(result).To(HaveLen(1))

			result = []*elldb.KVObject{}
			store.get(elldb.MakePrefix([]byte("an_obj"), []byte("account")), &result)
			Expect(result).To(HaveLen(1))
		})

		It("should successfully get object by key", func() {
			key := elldb.MakeKey([]byte("my_key"), []byte("block"), []byte("account"))
			err = store.put(key, []byte("stuff"))
			Expect(err).To(BeNil())

			var result = []*elldb.KVObject{}
			store.get(key, &result)
			Expect(result).To(HaveLen(1))
		})
	})
})
