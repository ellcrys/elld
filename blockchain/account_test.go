package blockchain

import (
	"os"

	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/elldb"
	"github.com/ellcrys/elld/testutil"
	"github.com/ellcrys/elld/txpool"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/types/core/objects"
	"github.com/ellcrys/elld/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Account", func() {

	var err error
	var bc *Blockchain
	var cfg *config.EngineConfig
	var db elldb.DB
	var genesisBlock core.Block
	var genesisChain *Chain

	BeforeEach(func() {
		cfg, err = testutil.SetTestCfg()
		Expect(err).To(BeNil())

		db = elldb.NewDB(cfg.ConfigDir())
		err = db.Open(util.RandString(5))
		Expect(err).To(BeNil())

		bc = New(txpool.New(100), cfg, log)
		bc.SetDB(db)
	})

	BeforeEach(func() {
		genesisBlock, err = LoadBlockFromFile("genesis-test.json")
		Expect(err).To(BeNil())
		bc.SetGenesisBlock(genesisBlock)
		err = bc.Up()
		Expect(err).To(BeNil())
		genesisChain = bc.bestChain
	})

	AfterEach(func() {
		db.Close()
		err = os.RemoveAll(cfg.ConfigDir())
		Expect(err).To(BeNil())
	})

	Describe(".putAccount", func() {

		var err error

		BeforeEach(func() {
			err = bc.CreateAccount(1, genesisChain, &objects.Account{
				Type:    objects.AccountTypeBalance,
				Address: "abc",
				Nonce:   1,
			})
		})

		It("should successfully create account", func() {
			Expect(err).To(BeNil())
		})
	})

	Describe(".GetNonce", func() {

		var account *objects.Account

		BeforeEach(func() {
			account = &objects.Account{Type: objects.AccountTypeBalance, Address: "abc", Nonce: 2}
			err = bc.CreateAccount(1, genesisChain, account)
			Expect(err).To(BeNil())
		})

		It("should return expected nonce = 2", func() {
			nonce, err := bc.GetAccountNonce("abc")
			Expect(err).To(BeNil())
			Expect(nonce).To(Equal(account.Nonce))
		})

		It("should return ErrAccountNotFound if account does not exist on the main chain", func() {
			nonce, err := bc.GetAccountNonce("xyz")
			Expect(err).ToNot(BeNil())
			Expect(err).To(Equal(core.ErrAccountNotFound))
			Expect(nonce).To(Equal(uint64(0)))
		})
	})
})
