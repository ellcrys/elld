package blockchain

import (
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/types/core/objects"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var AccountTest = func() bool {
	return Describe("Account", func() {

		Describe(".putAccount", func() {
			It("should successfully create account", func() {
				err = bc.CreateAccount(1, genesisChain, &objects.Account{
					Type:    objects.AccountTypeBalance,
					Address: "abc",
					Nonce:   1,
				})
				Expect(err).To(BeNil())
			})
		})

		Describe(".GetNonce", func() {

			var account = &objects.Account{
				Type:    objects.AccountTypeBalance,
				Address: "abc",
				Nonce:   2,
			}

			BeforeEach(func() {
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
}
