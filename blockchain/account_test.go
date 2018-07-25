package blockchain

import (
	"github.com/ellcrys/elld/blockchain/common"
	"github.com/ellcrys/elld/blockchain/leveldb"
	"github.com/ellcrys/elld/blockchain/testdata"
	"github.com/ellcrys/elld/crypto"
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
	var chain *Chain
	var bc *Blockchain
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
		GenesisBlock = testdata.TestGenesisBlock
		genesisBlock, _ = wire.BlockFromString(GenesisBlock)
		err = SeedTestGenesisState(store, genesisBlock)
		Expect(err).To(BeNil())
	})

	BeforeEach(func() {
		bc = New(cfg, log)
		bc.SetStore(store)
		err = bc.Up()
		Expect(err).To(BeNil())
		chain = bc.bestChain
	})

	AfterEach(func() {
		db.Close()
	})

	AfterEach(func() {
		Expect(testutil.RemoveTestCfgDir()).To(BeNil())
	})

	Describe(".putAccount", func() {

		var key *crypto.Key
		var account *wire.Account

		BeforeEach(func() {
			key = crypto.NewKeyFromIntSeed(1)
			account = &wire.Account{
				Type:    wire.AccountTypeBalance,
				Address: key.Addr(),
			}
		})

		It("should successfully create account with no err", func() {
			err = bc.putAccount(1, chain, account)
			Expect(err).To(BeNil())
		})
	})

	Describe(".GetAccount", func() {

		var key *crypto.Key
		var account *wire.Account

		BeforeEach(func() {
			key = crypto.NewKeyFromIntSeed(1)
			account = &wire.Account{
				Type:    wire.AccountTypeBalance,
				Address: key.Addr(),
			}
		})

		It("should return error if account is not supplied", func() {
			_, err := bc.GetAccount(chain, "")
			Expect(err).ToNot(BeNil())
			Expect(err).To(Equal(common.ErrAccountNotFound))
		})

		It("should return error if account does not exist", func() {
			_, err := bc.GetAccount(chain, "does_not_exist")
			Expect(err).ToNot(BeNil())
			Expect(err).To(Equal(common.ErrAccountNotFound))
		})

		Context("with one object matching the account prefix", func() {

			BeforeEach(func() {
				err = bc.putAccount(1, chain, account)
				Expect(err).To(BeNil())
			})

			It("should return the only object as the account", func() {
				a, err := bc.GetAccount(chain, account.Address)
				Expect(err).To(BeNil())
				Expect(a).ToNot(BeNil())
				Expect(a).To(Equal(account))
			})
		})

		Context("with more that one object matching the account prefix but differ by block number", func() {

			BeforeEach(func() {
				err = bc.putAccount(1, chain, account)
				Expect(err).To(BeNil())

				// update account
				account.Balance = "100"
				err = bc.putAccount(2, chain, account)
				Expect(err).To(BeNil())
			})

			It("should return the account with the highest block number", func() {
				a, err := bc.GetAccount(chain, account.Address)
				Expect(err).To(BeNil())
				Expect(a).ToNot(BeNil())
				Expect(a).To(Equal(account))
				Expect(a.Balance).To(Equal("100"))
			})
		})
	})
})
