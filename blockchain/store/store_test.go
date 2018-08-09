package store

import (
	"github.com/ellcrys/elld/blockchain/common"
	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/elldb"
	"github.com/ellcrys/elld/testutil"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/wire"
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
		Expect(testutil.RemoveTestCfgDir()).To(BeNil())
	})

	BeforeEach(func() {
		db = elldb.NewDB(cfg.ConfigDir())
		err = db.Open("")
		Expect(err).To(BeNil())
	})

	BeforeEach(func() {
		store = New(db, chainID)
		Expect(err).To(BeNil())
	})

	Describe(".PutTransactions", func() {

		var txs = []*wire.Transaction{
			&wire.Transaction{To: "to_addr", From: "from_addr", Hash: util.StrToHash("hash1")},
			&wire.Transaction{To: "to_addr_2", From: "from_addr_2", Hash: util.StrToHash("hash2")},
		}

		It("should successfully put transactions", func() {
			err = store.PutTransactions(txs)
			Expect(err).To(BeNil())

			r := store.db.GetByPrefix(common.MakeTxKey(store.chainID.Bytes(), txs[0].Hash.Bytes()))
			var tx wire.Transaction
			r[0].Scan(&tx)
			Expect(&tx).To(Equal(txs[0]))

			r = store.db.GetByPrefix(common.MakeTxKey(store.chainID.Bytes(), txs[1].Hash.Bytes()))
			r[0].Scan(&tx)
			Expect(&tx).To(Equal(txs[1]))
		})
	})

	Describe(".GetTransaction", func() {

		var txs = []*wire.Transaction{
			&wire.Transaction{To: "to_addr", From: "from_addr", Hash: util.StrToHash("hash1")},
			&wire.Transaction{To: "to_addr_2", From: "from_addr_2", Hash: util.StrToHash("hash2")},
		}

		It("should successfully put transactions", func() {
			err = store.PutTransactions(txs)
			Expect(err).To(BeNil())

			tx := store.GetTransaction(txs[0].Hash)
			Expect(tx).ToNot(BeNil())
			Expect(tx).To(Equal(txs[0]))
			tx = store.GetTransaction(txs[1].Hash)
			Expect(tx).ToNot(BeNil())
			Expect(tx).To(Equal(txs[1]))
		})
	})

	Describe(".CreateAccount", func() {
		It("should successfully create an account", func() {
			var acct = &wire.Account{Type: wire.AccountTypeBalance, Address: "addr"}
			err = store.CreateAccount(1, acct)
			Expect(err).To(BeNil())

			r := store.db.GetByPrefix(common.MakeAccountKey(1, store.chainID.Bytes(), []byte("addr")))
			var found wire.Account
			r[0].Scan(&found)
			Expect(&found).To(Equal(acct))
		})
	})

	Describe(".GetAccount", func() {
		Context("no existing account in store", func() {
			It("should return the only account", func() {
				var acct = &wire.Account{Type: wire.AccountTypeBalance, Address: "addr"}
				err = store.CreateAccount(1, acct)
				Expect(err).To(BeNil())

				found, err := store.GetAccount(util.String(acct.Address))
				Expect(err).To(BeNil())
				Expect(found).To(Equal(acct))
			})
		})

		Context("with multiple account of same address", func() {
			It("should contain account with the highest block number", func() {
				var acct = &wire.Account{Type: wire.AccountTypeBalance, Address: "addr"}
				err = store.CreateAccount(1, acct)
				Expect(err).To(BeNil())

				var acct2 = &wire.Account{Type: wire.AccountTypeBalance, Address: "addr2"}
				err = store.CreateAccount(2, acct2)
				Expect(err).To(BeNil())

				found, err := store.GetAccount((util.String(acct2.Address)))
				Expect(err).To(BeNil())
				Expect(found).To(Equal(acct2))
			})
		})
	})

	Describe(".PutBlock", func() {

		var block = &wire.Block{
			Header: &wire.Header{Number: 1},
			Hash:   util.StrToHash("hash"),
		}

		It("should put block without error", func() {
			err = store.PutBlock(block)
			Expect(err).To(BeNil())
			result := store.db.GetByPrefix(common.MakeBlocksQueryKey(chainID.Bytes()))
			Expect(result).To(HaveLen(1))

			var storedBlock wire.Block
			err = result[0].Scan(&storedBlock)
			Expect(err).To(BeNil())
			Expect(&storedBlock).To(Equal(block))
		})

		It("should return nil and not add block when another block with same number exists", func() {
			err = store.PutBlock(block)
			Expect(err).To(BeNil())
			result := store.db.GetByPrefix(common.MakeBlocksQueryKey(chainID.Bytes()))
			Expect(result).To(HaveLen(1))

			var storedBlock wire.Block
			err = result[0].Scan(&storedBlock)
			Expect(err).To(BeNil())
			Expect(&storedBlock).To(Equal(block))

			var block2 = &wire.Block{
				Header: &wire.Header{Number: 1},
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
			var block = &wire.Block{Header: &wire.Header{Number: 1}, Hash: util.StrToHash("hash")}
			var block2 = &wire.Block{Header: &wire.Header{Number: 2}, Hash: util.StrToHash("hash2")}

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
				Expect(err).To(Equal(common.ErrBlockNotFound))
			})
		})
	})

	Describe(".GetBlock", func() {

		var block = &wire.Block{
			Header: &wire.Header{Number: 1},
			Hash:   util.StrToHash("hash"),
		}

		BeforeEach(func() {
			err = store.PutBlock(block)
			Expect(err).To(BeNil())
			result := store.db.GetByPrefix(common.MakeBlocksQueryKey(chainID.Bytes()))
			Expect(result).To(HaveLen(1))
		})

		It("should get block by number", func() {
			storedBlock, err := store.GetBlock(block.Header.Number)
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
			Expect(err).To(Equal(common.ErrBlockNotFound))
			Expect(b).To(BeNil())
		})

		It("with block number; return error if block does not exist", func() {
			_, err = store.GetBlock(10000)
			Expect(err).ToNot(BeNil())
			Expect(err).To(Equal(common.ErrBlockNotFound))
		})

		It("should return the block with the hightest number if 0 is passed", func() {
			var block2 = &wire.Block{
				Header: &wire.Header{Number: 2},
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

		var block = &wire.Block{
			Header: &wire.Header{Number: 1},
			Hash:   util.StrToHash("hash"),
		}

		BeforeEach(func() {
			err = store.PutBlock(block)
			Expect(err).To(BeNil())
			result := store.db.GetByPrefix(common.MakeBlocksQueryKey(chainID.Bytes()))
			Expect(result).To(HaveLen(1))
		})

		It("should get block header by number", func() {
			storedBlockHeader, err := store.GetHeader(block.Header.Number)
			Expect(err).To(BeNil())
			Expect(storedBlockHeader).ToNot(BeNil())
			Expect(storedBlockHeader).To(Equal(block.Header))
		})
	})

	Describe(".GetBlockHeaderByHash", func() {
		var block = &wire.Block{
			Header: &wire.Header{Number: 1},
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
