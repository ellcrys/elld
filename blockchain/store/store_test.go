package store

import (
	"os"
	"time"

	"github.com/ellcrys/elld/crypto"

	"github.com/ellcrys/elld/blockchain/common"
	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/elldb"
	"github.com/ellcrys/elld/testutil"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/types/core"

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
		err = os.RemoveAll(cfg.DataDir())
		Expect(err).To(BeNil())
	})

	BeforeEach(func() {
		db = elldb.NewDB(cfg.NetDataDir())
		err = db.Open(util.RandString(5))
		Expect(err).To(BeNil())
	})

	BeforeEach(func() {
		store = New(db, chainID)
		Expect(err).To(BeNil())
	})

	Describe(".PutTransactions", func() {

		var txs = []types.Transaction{
			&core.Transaction{To: "to_addr", From: "from_addr", Hash: util.StrToHash("hash1")},
			&core.Transaction{To: "to_addr_2", From: "from_addr_2", Hash: util.StrToHash("hash2")},
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

		var txs = []types.Transaction{
			&core.Transaction{To: "to_addr", From: "from_addr", Hash: util.StrToHash("hash1")},
			&core.Transaction{To: "to_addr_2", From: "from_addr_2", Hash: util.StrToHash("hash2")},
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

			var block types.Block

			BeforeEach(func() {
				block = &core.Block{
					Header: &core.Header{Number: 211},
					Transactions: []*core.Transaction{
						{Hash: util.StrToHash("hash1")},
					},
				}
				err := store.PutBlock(block)
				Expect(err).To(BeNil())
			})

			It("should return ErrTxNotFound when the transaction is not in the block", func() {
				var txs = []types.Transaction{
					&core.Transaction{Hash: util.StrToHash("hash2")},
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
		var txs = []types.Transaction{
			&core.Transaction{To: "to_addr", From: "from_addr", Hash: util.StrToHash("hash1")},
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
			var acct = &core.Account{Type: core.AccountTypeBalance, Address: "addr"}
			err = store.CreateAccount(1, acct)
			Expect(err).To(BeNil())

			r := store.db.GetByPrefix(common.MakeKeyAccount(1, store.chainID.Bytes(), []byte("addr")))
			var found core.Account
			r[0].Scan(&found)
			Expect(&found).To(Equal(acct))
		})

		Context("with two accounts created on different blocks", func() {
			var accts []*core.Account
			BeforeEach(func() {
				for i := uint64(1); i <= 2; i++ {
					var acct = &core.Account{Type: core.AccountTypeBalance, Address: "addr"}
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
				var acct = &core.Account{Type: core.AccountTypeBalance, Address: "addr"}
				err = store.CreateAccount(1, acct)
				Expect(err).To(BeNil())

				found, err := store.GetAccount(util.String(acct.Address))
				Expect(err).To(BeNil())
				Expect(found).To(Equal(acct))
			})
		})

		Context("with multiple account of same address but different block number", func() {

			var acct, acct2 *core.Account

			BeforeEach(func() {
				acct = &core.Account{Type: core.AccountTypeBalance, Balance: "0.1", Address: "addr"}
				err = store.CreateAccount(1, acct)
				Expect(err).To(BeNil())

				acct2 = &core.Account{Type: core.AccountTypeBalance, Balance: "1.2", Address: "addr"}
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
						latestAccount, err := store.GetAccount(acct2.Address, &common.OpBlockQueryRange{Min: 3})
						Expect(err).ToNot(BeNil())
						Expect(err).To(Equal(core.ErrAccountNotFound))
						Expect(latestAccount).To(BeNil())
					})
				})

				Context("with minimum block = 2", func() {
					It("should return account with the highest block number", func() {
						latestAccount, err := store.GetAccount(acct2.Address, &common.OpBlockQueryRange{Min: 2})
						Expect(err).To(BeNil())
						Expect(latestAccount).To(Equal(acct2))
					})
				})

				Context("with maximum block = 2", func() {
					It("should return the account with the highest block number less than or equal to the maximum block range", func() {
						latestAccount, err := store.GetAccount(acct2.Address, &common.OpBlockQueryRange{Max: 2})
						Expect(err).To(BeNil())
						Expect(latestAccount).To(Equal(acct2))
					})
				})

				Context("with maximum block = 1", func() {
					It("should return the account with the highest block number less than or equal to the maximum block range", func() {
						latestAccount, err := store.GetAccount(acct2.Address, &common.OpBlockQueryRange{Max: 1})
						Expect(err).To(BeNil())
						Expect(latestAccount).To(Equal(acct))
					})
				})
			})
		})
	})

	Describe(".GetAccounts", func() {
		Context("when two accounts are stored", func() {

			var acct, acct2 *core.Account

			BeforeEach(func() {
				acct = &core.Account{Type: core.AccountTypeBalance, Address: "addr"}
				err = store.CreateAccount(1, acct)
				Expect(err).To(BeNil())

				acct2 = &core.Account{Type: core.AccountTypeBalance, Address: "addr2"}
				err = store.CreateAccount(2, acct2)
				Expect(err).To(BeNil())
			})

			It("should return two accounts", func() {
				result, err := store.GetAccounts()
				Expect(err).To(BeNil())
				Expect(result).To(HaveLen(2))
			})
		})

		Context("with two accounts of same address at different blocks", func() {
			var acct, acct2 *core.Account

			BeforeEach(func() {
				acct = &core.Account{Type: core.AccountTypeBalance,
					Balance: "1", Address: "addr"}
				err = store.CreateAccount(1, acct)
				Expect(err).To(BeNil())

				acct2 = &core.Account{Type: core.AccountTypeBalance,
					Balance: "2", Address: "addr"}
				err = store.CreateAccount(2, acct2)
				Expect(err).To(BeNil())
			})

			It("should return one account", func() {
				result, err := store.GetAccounts()
				Expect(err).To(BeNil())
				Expect(result).To(HaveLen(1))
			})

			It("should return the account at the highest block", func() {
				result, err := store.GetAccounts()
				Expect(err).To(BeNil())
				Expect(result[0].GetBalance().String()).To(Equal("2"))
			})
		})

		Context("Using BlockRange", func() {
			Context("with minimum = 2", func() {
				Context("with two accounts of same address at different blocks", func() {
					var acct, acct2 *core.Account

					BeforeEach(func() {
						acct = &core.Account{Type: core.AccountTypeBalance,
							Balance: "1", Address: "addr"}
						err = store.CreateAccount(1, acct)
						Expect(err).To(BeNil())

						acct2 = &core.Account{Type: core.AccountTypeBalance,
							Balance: "2", Address: "addr"}
						err = store.CreateAccount(2, acct2)
						Expect(err).To(BeNil())
					})

					It("should return one account", func() {
						result, err := store.GetAccounts(&common.OpBlockQueryRange{Min: 2})
						Expect(err).To(BeNil())
						Expect(result).To(HaveLen(1))
						Expect(result[0].GetBalance().String()).To(Equal("2"))
					})
				})
			})

			Context("with maximum = 3", func() {
				Context("with two accounts of different address at block 1 and 3 respectively", func() {
					var acct, acct2 *core.Account

					BeforeEach(func() {
						acct = &core.Account{Type: core.AccountTypeBalance,
							Balance: "1", Address: "addr"}
						err = store.CreateAccount(1, acct)
						Expect(err).To(BeNil())

						acct2 = &core.Account{Type: core.AccountTypeBalance,
							Balance: "30", Address: "addr2"}
						err = store.CreateAccount(3, acct2)
						Expect(err).To(BeNil())
					})

					It("should return one account", func() {
						result, err := store.GetAccounts(&common.OpBlockQueryRange{Max: 3})
						Expect(err).To(BeNil())
						Expect(result).To(HaveLen(2))
					})
				})
			})

			Context("with maximum = 4", func() {
				Context("with two accounts of different address at block 1 and 5 respectively", func() {
					var acct, acct2 *core.Account

					BeforeEach(func() {
						acct = &core.Account{Type: core.AccountTypeBalance,
							Balance: "1", Address: "addr"}
						err = store.CreateAccount(1, acct)
						Expect(err).To(BeNil())

						acct2 = &core.Account{Type: core.AccountTypeBalance,
							Balance: "3", Address: "addr2"}
						err = store.CreateAccount(5, acct2)
						Expect(err).To(BeNil())
					})

					It("should return 1 account", func() {
						result, err := store.GetAccounts(&common.OpBlockQueryRange{Max: 4})
						Expect(err).To(BeNil())
						Expect(result).To(HaveLen(1))
						Expect(result[0].GetBalance().String()).To(Equal("1"))
					})
				})
			})
		})

	})

	Describe(".PutBlock", func() {

		var block *core.Block

		BeforeEach(func() {
			block = &core.Block{
				Header: &core.Header{Number: 1},
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
				var storedBlock core.Block
				err = result[0].Scan(&storedBlock)
				Expect(err).To(BeNil())
				Expect(&storedBlock).To(Equal(block))
			})

			Specify("that a block hash key that points to the block number must be added", func() {
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

			var storedBlock core.Block
			err = result[0].Scan(&storedBlock)
			Expect(err).To(BeNil())
			Expect(&storedBlock).To(Equal(block))

			var block2 = &core.Block{
				Header: &core.Header{Number: 1},
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

	Describe(".PutMinedBlock", func() {
		var block *core.Block

		BeforeEach(func() {
			key, _ := crypto.NewKey(nil)
			block = &core.Block{
				Header: &core.Header{
					Number:        1,
					CreatorPubKey: "xabc",
				},
				Hash: util.StrToHash("hash"),
				Sig:  []byte("stuff"),
				Transactions: []*core.Transaction{
					core.NewTx(core.TxTypeBalance, 1, "to_addr", key, "23", "2", time.Now().Unix()),
				},
			}
		})

		Context("on successful save", func() {
			var result []*elldb.KVObject

			BeforeEach(func() {
				err := store.PutMinedBlock(block)
				Expect(err).To(BeNil())
				result = store.db.GetByPrefix(common.MakeQueryKeyMinedBlocks(chainID.Bytes()))
				Expect(result).To(HaveLen(1))
			})

			Specify("the return block is same as the added saved block", func() {
				var minedBlock core.MinedBlock
				err = result[0].Scan(&minedBlock)
				Expect(err).To(BeNil())
				Expect(minedBlock.CreatorPubKey).To(Equal(block.Header.CreatorPubKey))
				Expect(minedBlock.Hash).To(Equal(block.Hash))
				Expect(minedBlock.Number).To(Equal(block.Header.Number))
				Expect(minedBlock.TxCount).To(Equal(uint(1)))
			})
		})
	})

	Describe(".Current", func() {
		When("two blocks are in the chain", func() {
			var block = &core.Block{Header: &core.Header{Number: 1}, Hash: util.StrToHash("hash")}
			var block2 = &core.Block{Header: &core.Header{Number: 2}, Hash: util.StrToHash("hash2")}

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

		var block = &core.Block{
			Header: &core.Header{Number: 100},
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

		var block = &core.Block{
			Header: &core.Header{Number: 1},
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
			var block2 = &core.Block{
				Header: &core.Header{Number: 2},
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

		var block = &core.Block{
			Header: &core.Header{Number: 1},
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
		var block = &core.Block{
			Header: &core.Header{Number: 1},
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
