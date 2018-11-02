package accountmgr

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ellcrys/elld/crypto"
	. "github.com/ncodes/goblin"
	. "github.com/onsi/gomega"
)

func TestRead(t *testing.T) {
	g := Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("Read", func() {

		path := filepath.Join("./", "test_cfg")
		accountPath := filepath.Join(path, "accounts")

		g.BeforeEach(func() {
			err := os.MkdirAll(accountPath, 0700)
			Expect(err).To(BeNil())
		})

		g.AfterEach(func() {
			err := os.RemoveAll(path)
			Expect(err).To(BeNil())
		})

		g.Describe("AccountManager", func() {

			g.Describe(".AccountExist", func() {

				am := New(accountPath)

				g.It("should return true and err = nil when account exists", func() {
					seed := int64(1)
					address, _ := crypto.NewKey(&seed)
					passphrase := "edge123"
					err := am.CreateAccount(address, passphrase)
					Expect(err).To(BeNil())

					exist, err := am.AccountExist(address.Addr().String())
					Expect(err).To(BeNil())
					Expect(exist).To(BeTrue())
				})

				g.It("should return false and err = nil when account does not exist", func() {
					seed := int64(1)
					address, _ := crypto.NewKey(&seed)

					exist, err := am.AccountExist(address.Addr().String())
					Expect(err).To(BeNil())
					Expect(exist).To(BeFalse())
				})

			})

			g.Describe(".GetDefault", func() {

				am := New(accountPath)

				g.It("should return oldest account with address", func() {
					seed := int64(1)
					address, _ := crypto.NewKey(&seed)
					passphrase := "edge123"
					err := am.CreateAccount(address, passphrase)
					Expect(err).To(BeNil())
					time.Sleep(1 * time.Second)

					seed = int64(2)
					address2, _ := crypto.NewKey(&seed)
					passphrase = "edge123"
					err = am.CreateAccount(address2, passphrase)
					Expect(err).To(BeNil())

					account, err := am.GetDefault()
					Expect(err).To(BeNil())
					Expect(account).ToNot(BeNil())
					Expect(account.Address).To(Equal(address.Addr().String()))
				})

				g.It("should return nil if no address was found", func() {
					account, err := am.GetDefault()
					Expect(err).To(BeNil())
					Expect(account).To(BeNil())
				})
			})

			g.Describe(".GetByIndex", func() {

				var address, address2 *crypto.Key
				am := New(accountPath)

				g.BeforeEach(func() {
					seed := int64(1)
					address, _ = crypto.NewKey(&seed)
					passphrase := "edge123"
					err := am.CreateAccount(address, passphrase)
					Expect(err).To(BeNil())
					time.Sleep(1 * time.Second)

					seed = int64(2)
					address2, _ = crypto.NewKey(&seed)
					passphrase = "edge123"
					err = am.CreateAccount(address2, passphrase)
					Expect(err).To(BeNil())
				})

				g.It("should get accounts at index 0 and 1", func() {
					act, err := am.GetByIndex(0)
					Expect(err).To(BeNil())
					Expect(act.Address).To(Equal(address.Addr().String()))
					act, err = am.GetByIndex(1)
					Expect(act.Address).To(Equal(address2.Addr().String()))
				})

				g.It("should return err = 'account not found' when no account is found", func() {
					_, err := am.GetByIndex(2)
					Expect(err).ToNot(BeNil())
					Expect(err).To(Equal(ErrAccountNotFound))
				})
			})

			g.Describe(".GetByAddress", func() {

				var address *crypto.Key
				am := New(accountPath)

				g.BeforeEach(func() {
					seed := int64(1)
					address, _ = crypto.NewKey(&seed)
					passphrase := "edge123"
					err := am.CreateAccount(address, passphrase)
					Expect(err).To(BeNil())
				})

				g.It("should successfully get account with address", func() {
					act, err := am.GetByAddress(address.Addr().String())
					Expect(err).To(BeNil())
					Expect(act.Address).To(Equal(address.Addr().String()))
				})

				g.It("should return err = 'account not found' when address does not exist", func() {
					_, err := am.GetByAddress("unknown_address")
					Expect(err).ToNot(BeNil())
					Expect(err).To(Equal(ErrAccountNotFound))
				})
			})
		})

		g.Describe("StoredAccount", func() {

			g.Describe(".Decrypt", func() {

				var account *StoredAccount
				var passphrase string
				am := New(accountPath)

				g.BeforeEach(func() {
					var err error
					seed := int64(1)

					address, _ := crypto.NewKey(&seed)
					passphrase = "edge123"
					err = am.CreateAccount(address, passphrase)
					Expect(err).To(BeNil())

					account, err = am.GetDefault()
					Expect(err).To(BeNil())
					Expect(account).ToNot(BeNil())
				})

				g.It("should return err = 'invalid password' when password is invalid", func() {
					err := account.Decrypt("invalid")
					Expect(err).ToNot(BeNil())
					Expect(err.Error()).To(Equal("invalid password"))
				})

				g.It("should return nil when decryption is successful. account.address must not be nil.", func() {
					err := account.Decrypt(passphrase)
					Expect(err).To(BeNil())
					Expect(account.key).ToNot(BeNil())
				})
			})
		})

	})
}
