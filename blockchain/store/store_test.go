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

var _ = Describe("Store", func() {

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
		db.Close()
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

		It("should store 2 transaction block pointers", func() {
			err = store.PutTransactions(txs, 211)
			Expect(err).To(BeNil())

			r := store.db.GetByPrefix(common.MakeQueryKeyTransactions(store.chainID.Bytes()))
			Expect(r).To(HaveLen(2))

			tx1Value := util.DecodeNumber(r[0].Value)
			tx2Value := util.DecodeNumber(r[1].Value)
			Expect(tx1Value).To(Equal(uint64(211)))
			Expect(tx2Value).To(Equal(uint64(211)))
		})
	})

	Describe(".GetTransaction", func() {

		var txs = []core.Transaction{
			&objects.Transaction{To: "to_addr", From: "from_addr", Hash: util.StrToHash("hash1")},
			&objects.Transaction{To: "to_addr_2", From: "from_addr_2", Hash: util.StrToHash("hash2")},
		}

		Context("when the transaction's value holds an non-existing block number", func() {
			It("should return err", func() {
				err = store.PutTransactions(txs, 211)
				Expect(err).To(BeNil())

				_, err := store.GetTransaction(txs[0].GetHash())
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("transaction's block not found"))
			})
		})

		Context("when the transaction's value holds an existing block number", func() {

			var block core.Block

			BeforeEach(func() {
				block = &objects.Block{
					Header: &objects.Header{Number: 211},
					Transactions: []*objects.Transaction{
						{Hash: util.StrToHash("hash1")},
					},
				}
				err := store.PutBlock(block)
				Expect(err).To(BeNil())
			})

			It("should return ErrTxNotFound when the transaction is not in the block", func() {
				var txs = []core.Transaction{
					&objects.Transaction{Hash: util.StrToHash("hash2")},
				}
				err = store.PutTransactions(txs, 211)
				Expect(err).To(BeNil())
				_, err := store.GetTransaction(util.StrToHash("hash2"))
				Expect(err).ToNot(BeNil())
				Expect(err).To(Equal(core.ErrTxNotFound))
			})

			It("should successfully get transaction when transaction exist in the block", func() {
				err = store.PutTransactions(block.GetTransactions(), 211)
				Expect(err).To(BeNil())
				tx, err := store.GetTransaction(util.StrToHash("hash1"))
				Expect(err).To(BeNil())
				Expect(tx).To(Equal(block.GetTransactions()[0]))
			})
		})
	})

	Describe(".Delete", func() {
		var txs = []core.Transaction{
			&objects.Transaction{To: "to_addr", From: "from_addr", Hash: util.StrToHash("hash1")},
		}

		It("should successfully delete", func() {
			err = store.PutTransactions(txs, 211)
			Expect(err).To(BeNil())

			err := store.Delete(common.MakeQueryKeyTransaction(store.chainID.Bytes(), txs[0].GetHash().Hex()))
			Expect(err).To(BeNil())

			tx, err := store.GetTransaction(txs[0].GetHash())
			Expect(err).ToNot(BeNil())
			Expect(err).To(Equal(core.ErrTxNotFound))
			Expect(tx).To(BeNil())
		})
	})

	Describe(".CreateAccount", func() {
		It("should successfully create an account", func() {
			var acct = &objects.Account{Type: objects.AccountTypeBalance, Address: "addr"}
			err = store.CreateAccount(1, acct)
			Expect(err).To(BeNil())

			r := store.db.GetByPrefix(common.MakeKeyAccount(1, store.chainID.Bytes(), []byte("addr")))
			var found objects.Account
			r[0].Scan(&found)
			Expect(&found).To(Equal(acct))
		})

		Context("with two accounts created on different blocks", func() {
			var accts []*objects.Account
			BeforeEach(func() {
				for i := uint64(1); i <= 2; i++ {
					var acct = &objects.Account{Type: objects.AccountTypeBalance, Address: "addr"}
					err = store.CreateAccount(i, acct)
					Expect(err).To(BeNil())
					accts = append(accts, acct)
				}
				Expect(accts).To(HaveLen(2))
			})

			Specify("when all account is fetched, the account with the highest block number (2) must be last", func() {
				fetchedAccts := store.db.GetByPrefix(common.MakeQueryKeyAccounts(store.chainID.Bytes()))
				Expect(fetchedAccts).To(HaveLen(2))
				Expect(util.DecodeNumber(fetchedAccts[1].Key)).To(Equal(uint64(2)))
			})
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

			var acct, acct2 *objects.Account

			BeforeEach(func() {
				acct = &objects.Account{Type: objects.AccountTypeBalance, Balance: "0.1", Address: "addr"}
				err = store.CreateAccount(1, acct)
				Expect(err).To(BeNil())

				acct2 = &objects.Account{Type: objects.AccountTypeBalance, Balance: "1.2", Address: "addr"}
				err = store.CreateAccount(2, acct2)
				Expect(err).To(BeNil())

				addedAccts := store.db.GetByPrefix(common.MakeQueryKeyAccounts(store.chainID.Bytes()))
				Expect(addedAccts).To(HaveLen(2))
			})

			It("should return account with the highest block number", func() {
				latestAccount, err := store.GetAccount(acct2.Address)
				Expect(err).To(BeNil())
				Expect(latestAccount).To(Equal(acct2))
			})

			Context("when latest account is at block 2 and block range option is provided", func() {
				Context("with minimum block = 3", func() {
					It("should return ErrAccountNotFound", func() {
						latestAccount, err := store.GetAccount(acct2.Address, &common.BlockQueryRange{Min: 3})
						Expect(err).ToNot(BeNil())
						Expect(err).To(Equal(core.ErrAccountNotFound))
						Expect(latestAccount).To(BeNil())
					})
				})

				Context("with minimum block = 2", func() {
					It("should return account with the highest block number", func() {
						latestAccount, err := store.GetAccount(acct2.Address, &common.BlockQueryRange{Min: 2})
						Expect(err).To(BeNil())
						Expect(latestAccount).To(Equal(acct2))
					})
				})

				Context("with maximum block = 2", func() {
					It("should return the account with the highest block number less than or equal to the maximum block range", func() {
						latestAccount, err := store.GetAccount(acct2.Address, &common.BlockQueryRange{Max: 2})
						Expect(err).To(BeNil())
						Expect(latestAccount).To(Equal(acct2))
					})
				})

				Context("with maximum block = 1", func() {
					It("should return the account with the highest block number less than or equal to the maximum block range", func() {
						latestAccount, err := store.GetAccount(acct2.Address, &common.BlockQueryRange{Max: 1})
						Expect(err).To(BeNil())
						Expect(latestAccount).To(Equal(acct))
					})
				})
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

		Context("on successful save", func() {

			var result []*elldb.KVObject

			BeforeEach(func() {
				err := store.PutBlock(block)
				Expect(err).To(BeNil())
				result = store.db.GetByPrefix(common.MakeQueryKeyBlocks(chainID.Bytes()))
				Expect(result).To(HaveLen(1))
			})

			Specify("the return block is same as the added saved block", func() {
				var storedBlock objects.Block
				err = result[0].Scan(&storedBlock)
				Expect(err).To(BeNil())
				Expect(&storedBlock).To(Equal(block))
			})

			Specify("a block number pointer should be added", func() {
				pointerKey := common.MakeKeyBlockHash(store.chainID.Bytes(), block.GetHash().Hex())
				result = store.db.GetByPrefix(pointerKey)
				Expect(result).To(HaveLen(1))
				Expect(util.DecodeNumber(result[0].Value)).To(Equal(block.GetNumber()))
			})
		})

		It("should return nil and not add block when another block with same number exists", func() {
			err = store.PutBlock(block)
			Expect(err).To(BeNil())
			result := store.db.GetByPrefix(common.MakeQueryKeyBlocks(chainID.Bytes()))
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
			result = store.db.GetByPrefix(common.MakeQueryKeyBlocks(chainID.Bytes()))
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
			result := store.db.GetByPrefix(common.MakeQueryKeyBlocks(chainID.Bytes()))
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
			result := store.db.GetByPrefix(common.MakeQueryKeyBlocks(chainID.Bytes()))
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
			result := store.db.GetByPrefix(common.MakeQueryKeyBlocks(chainID.Bytes()))
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
			result := store.db.GetByPrefix(common.MakeQueryKeyBlocks(chainID.Bytes()))
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
