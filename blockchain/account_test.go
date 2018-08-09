package blockchain

import (
	"github.com/ellcrys/elld/blockchain/common"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/wire"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var AccountTest = func() bool {
	return Describe("Account", func() {
		Describe(".putAccount", func() {

			var key *crypto.Key
			var account *wire.Account

			BeforeEach(func() {
				key = crypto.NewKeyFromIntSeed(1)
				account = &wire.Account{
					Type:    wire.AccountTypeBalance,
					Address: util.String(key.Addr()),
				}
			})

			It("should successfully create account with no err", func() {
				err = bc.putAccount(1, genesisChain, account)
				Expect(err).To(BeNil())
			})
		})

		Describe(".getAccount", func() {

			var key *crypto.Key
			var account *wire.Account

			BeforeEach(func() {
				key = crypto.NewKeyFromIntSeed(1)
				account = &wire.Account{
					Type:    wire.AccountTypeBalance,
					Address: util.String(key.Addr()),
				}
			})

			It("should return error if account is not supplied", func() {
				_, err := bc.getAccount(genesisChain, "")
				Expect(err).ToNot(BeNil())
				Expect(err).To(Equal(common.ErrAccountNotFound))
			})

			It("should return error if account does not exist", func() {
				_, err := bc.getAccount(genesisChain, "does_not_exist")
				Expect(err).ToNot(BeNil())
				Expect(err).To(Equal(common.ErrAccountNotFound))
			})

			Context("with one object matching the account prefix", func() {

				BeforeEach(func() {
					err = bc.putAccount(1, genesisChain, account)
					Expect(err).To(BeNil())
				})

				It("should return the only object as the account", func() {
					a, err := bc.getAccount(genesisChain, account.Address)
					Expect(err).To(BeNil())
					Expect(a).ToNot(BeNil())
					Expect(a).To(Equal(account))
				})
			})

			Context("with more that one object matching the account prefix but differ by block number", func() {

				BeforeEach(func() {
					err = bc.putAccount(1, genesisChain, account)
					Expect(err).To(BeNil())

					// update account
					account.Balance = "100"
					err = bc.putAccount(2, genesisChain, account)
					Expect(err).To(BeNil())
				})

				It("should return the account with the highest block number", func() {
					a, err := bc.getAccount(genesisChain, account.Address)
					Expect(err).To(BeNil())
					Expect(a).ToNot(BeNil())
					Expect(a).To(Equal(account))
					Expect(a.Balance).To(Equal(util.String("100")))
				})
			})

			Context("account object on the parent chain", func() {

				var child *Chain

				BeforeEach(func() {
					chainParent := NewChain("papa", db, cfg, log)
					bc.addChain(chainParent)

					child = NewChain("child", db, cfg, log)
					child.info.ParentChainID = chainParent.id

					err = bc.putAccount(1, chainParent, account)
					Expect(err).To(BeNil())
				})

				It("should return the account object", func() {
					a, err := bc.getAccount(child, account.Address)
					Expect(err).To(BeNil())
					Expect(a).ToNot(BeNil())
					Expect(a).To(Equal(account))
				})
			})

			Context("account object on the grand-parent chain", func() {

				var child *Chain

				BeforeEach(func() {
					chainGrandParent := NewChain("grand-papa", db, cfg, log)
					bc.addChain(chainGrandParent)

					chainParent := NewChain("papa", db, cfg, log)
					bc.addChain(chainParent)
					chainParent.info.ParentChainID = chainGrandParent.id

					child = NewChain("child", db, cfg, log)
					child.info.ParentChainID = chainParent.id

					err = bc.putAccount(1, chainGrandParent, account)
					Expect(err).To(BeNil())
				})

				It("should return the account object", func() {
					a, err := bc.getAccount(child, account.Address)
					Expect(err).To(BeNil())
					Expect(a).ToNot(BeNil())
					Expect(a).To(Equal(account))
				})
			})
		})
	})
}
