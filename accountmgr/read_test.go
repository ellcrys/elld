package accountmgr

import (
	"os"
	"path/filepath"
	"time"

	"github.com/ellcrys/druid/crypto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Read", func() {

	path := filepath.Join("./", "test_cfg")
	accountPath := filepath.Join(path, "accounts")

	BeforeEach(func() {
		err := os.MkdirAll(accountPath, 0700)
		Expect(err).To(BeNil())
	})

	Describe("AccountManager", func() {

		Describe(".AccountExist", func() {

			am := New(accountPath)

			It("should return true and err = nil when account exists", func() {
				seed := int64(1)
				address, _ := crypto.NewKey(&seed)
				passphrase := "edge123"
				err := am.CreateAccount(address, passphrase)
				Expect(err).To(BeNil())

				exist, err := am.AccountExist(address.Addr())
				Expect(err).To(BeNil())
				Expect(exist).To(BeTrue())
			})

			It("should return false and err = nil when account does not exist", func() {
				seed := int64(1)
				address, _ := crypto.NewKey(&seed)

				exist, err := am.AccountExist(address.Addr())
				Expect(err).To(BeNil())
				Expect(exist).To(BeFalse())
			})

		})

		Describe(".GetDefault", func() {

			am := New(accountPath)

			It("should return oldest account with address", func() {
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
				Expect(account.Address).To(Equal(address.Addr()))
			})

			It("should return nil if no address was found", func() {
				account, err := am.GetDefault()
				Expect(err).To(BeNil())
				Expect(account).To(BeNil())
			})
		})

		Describe(".GetByIndex", func() {

			var address, address2 *crypto.Key
			am := New(accountPath)

			BeforeEach(func() {
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

			It("should get accounts at index 0 and 1", func() {
				act, err := am.GetByIndex(0)
				Expect(err).To(BeNil())
				Expect(act.Address).To(Equal(address.Addr()))
				act, err = am.GetByIndex(1)
				Expect(act.Address).To(Equal(address2.Addr()))
			})

			It("should return err = 'account not found' when no account is found", func() {
				_, err := am.GetByIndex(2)
				Expect(err).ToNot(BeNil())
				Expect(err).To(Equal(ErrAccountNotFound))
			})
		})

		Describe(".GetByAddress", func() {

			var address *crypto.Key
			am := New(accountPath)

			BeforeEach(func() {
				seed := int64(1)
				address, _ = crypto.NewKey(&seed)
				passphrase := "edge123"
				err := am.CreateAccount(address, passphrase)
				Expect(err).To(BeNil())
			})

			It("should successfully get account with address", func() {
				act, err := am.GetByAddress(address.Addr())
				Expect(err).To(BeNil())
				Expect(act.Address).To(Equal(address.Addr()))
			})

			It("should return err = 'account not found' when address does not exist", func() {
				_, err := am.GetByAddress("unknown_address")
				Expect(err).ToNot(BeNil())
				Expect(err).To(Equal(ErrAccountNotFound))
			})
		})
	})

	Describe("StoredAccount", func() {

		Describe(".Decrypt", func() {

			var account *StoredAccount
			var passphrase string
			am := New(accountPath)

			BeforeEach(func() {
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

			It("should return err = 'invalid password. invalid format: version and/or checksum bytes missing' when password is invalid", func() {
				err := account.Decrypt("invalid")
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("invalid password. invalid format: version and/or checksum bytes missing"))
			})

			It("should return nil when decryption is successful. account.address must not be nil.", func() {
				err := account.Decrypt(passphrase)
				Expect(err).To(BeNil())
				Expect(account.address).ToNot(BeNil())
			})
		})
	})

	AfterEach(func() {
		err := os.RemoveAll(path)
		Expect(err).To(BeNil())
	})
})
