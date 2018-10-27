package blockchain

import (
	"os"
	"testing"

	. "github.com/ncodes/goblin"

	"github.com/ellcrys/elld/blockchain/txpool"
	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/elldb"
	"github.com/ellcrys/elld/testutil"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/types/core"

	"github.com/ellcrys/elld/util"
	. "github.com/onsi/gomega"
)

func TestAccount(t *testing.T) {
	g := Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("Account", func() {

		var err error
		var bc *Blockchain
		var cfg *config.EngineConfig
		var db elldb.DB
		var genesisBlock types.Block
		var genesisChain *Chain

		g.BeforeEach(func() {
			cfg, err = testutil.SetTestCfg()
			Expect(err).To(BeNil())

			db = elldb.NewDB(cfg.DataDir())
			err = db.Open(util.RandString(5))
			Expect(err).To(BeNil())

			bc = New(txpool.New(100), cfg, log)
			bc.SetDB(db)
		})

		g.BeforeEach(func() {
			genesisBlock, err = LoadBlockFromFile("genesis-test.json")
			Expect(err).To(BeNil())
			bc.SetGenesisBlock(genesisBlock)
			err = bc.Up()
			Expect(err).To(BeNil())
			genesisChain = bc.bestChain
		})

		g.AfterEach(func() {
			db.Close()
			err = os.RemoveAll(cfg.DataDir())
			Expect(err).To(BeNil())
		})

		g.Describe(".putAccount", func() {

			var err error

			g.BeforeEach(func() {
				err = bc.CreateAccount(1, genesisChain, &core.Account{
					Type:    core.AccountTypeBalance,
					Address: "abc",
					Nonce:   1,
				})
			})

			g.It("should successfully create account", func() {
				Expect(err).To(BeNil())
			})
		})

		g.Describe(".GetAccountNonce", func() {

			var account *core.Account

			g.BeforeEach(func() {
				account = &core.Account{Type: core.AccountTypeBalance, Address: "abc", Nonce: 2}
				err = bc.CreateAccount(1, genesisChain, account)
				Expect(err).To(BeNil())
			})

			g.It("should return expected nonce = 2", func() {
				nonce, err := bc.GetAccountNonce("abc")
				Expect(err).To(BeNil())
				Expect(nonce).To(Equal(account.Nonce))
			})

			g.It("should return ErrAccountNotFound if account does not exist on the main chain", func() {
				nonce, err := bc.GetAccountNonce("xyz")
				Expect(err).ToNot(BeNil())
				Expect(err).To(Equal(core.ErrAccountNotFound))
				Expect(nonce).To(Equal(uint64(0)))
			})
		})

		g.Describe(".ListAccounts", func() {
			g.It("should return error when best chain is unknown", func() {
				bc.bestChain = nil
				_, err := bc.ListAccounts()
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("best chain unknown"))
			})

			g.Context("with 2 accounts stored", func() {
				var account, account2 *core.Account

				g.BeforeEach(func() {
					account = &core.Account{Type: core.AccountTypeBalance, Address: "abc"}
					err = bc.CreateAccount(1, genesisChain, account)
					Expect(err).To(BeNil())

					account2 = &core.Account{Type: core.AccountTypeBalance, Address: "xyz"}
					err = bc.CreateAccount(2, genesisChain, account2)
					Expect(err).To(BeNil())
				})

				g.It("should return all accounts", func() {
					result, err := bc.ListAccounts()
					Expect(err).To(BeNil())
					Expect(result).To(HaveLen(3))
				})
			})
		})

		g.Describe(".ListTopAccounts", func() {
			var account, account2 *core.Account

			g.BeforeEach(func() {
				account = &core.Account{Type: core.AccountTypeBalance,
					Address: "abc", Balance: "10"}
				err = bc.CreateAccount(1, genesisChain, account)
				Expect(err).To(BeNil())

				account2 = &core.Account{Type: core.AccountTypeBalance,
					Address: "xyz", Balance: "300"}
				err = bc.CreateAccount(2, genesisChain, account2)
				Expect(err).To(BeNil())
			})

			g.It("should return top accounts", func() {
				result, err := bc.ListTopAccounts(100)
				Expect(err).To(BeNil())
				Expect(result).To(HaveLen(3))
				Expect(result[0].GetBalance().String()).To(Equal("300"))
				Expect(result[2].GetBalance().String()).To(Equal("10"))
			})
		})
	})
}
