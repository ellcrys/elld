package accountmgr

import (
	"os"
	"path/filepath"
	"time"

	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/crypto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Read", func() {

	path := filepath.Join("./", "test_cfg")
	accountPath := filepath.Join(path, "accounts")
	burnerAccountPath := filepath.Join(accountPath, config.BurnerAccountDirName)

	BeforeEach(func() {
		err := os.MkdirAll(accountPath, 0700)
		Expect(err).To(BeNil())
		err = os.MkdirAll(burnerAccountPath, 0700)
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		err := os.RemoveAll(path)
		Expect(err).To(BeNil())
	})

	Describe("AccountManager", func() {

		Describe(".AccountExist", func() {

			am := New(accountPath)

			It("should return true and err = nil when account exists", func() {
				seed := int64(1)
				address, _ := crypto.NewKey(&seed)
				passphrase := "edge123"
				err := am.CreateAccount(false, address, passphrase)
				Expect(err).To(BeNil())

				exist, err := am.AccountExist(address.Addr().String())
				Expect(err).To(BeNil())
				Expect(exist).To(BeTrue())
			})

			It("should return false and err = nil when account does not exist", func() {
				seed := int64(1)
				address, _ := crypto.NewKey(&seed)

				exist, err := am.AccountExist(address.Addr().String())
				Expect(err).To(BeNil())
				Expect(exist).To(BeFalse())
			})
		})

		Describe(".BurnerAccountExist", func() {

			am := New(accountPath)

			It("should return true and err = nil when account exists", func() {
				seed := int64(1)
				key, _ := crypto.NewSecp256k1(&seed, true, true)
				passphrase := "edge123"
				err := am.CreateBurnerAccount(key, passphrase)
				Expect(err).To(BeNil())

				exist, err := am.BurnerAccountExist(key.Addr())
				Expect(err).To(BeNil())
				Expect(exist).To(BeTrue())
			})

			It("should return false and err = nil when account does not exist", func() {
				seed := int64(1)
				key, _ := crypto.NewSecp256k1(&seed, true, true)

				exist, err := am.AccountExist(key.Addr())
				Expect(err).To(BeNil())
				Expect(exist).To(BeFalse())
			})
		})

		Describe(".GetDefault", func() {

			am := New(accountPath)

			It("should return the address whose keyfile ends with `_default`", func() {
				seed := int64(1)
				address, _ := crypto.NewKey(&seed)
				passphrase := "edge123"
				err := am.CreateAccount(true, address, passphrase)
				Expect(err).To(BeNil())
				time.Sleep(1 * time.Second)

				seed = int64(2)
				address2, _ := crypto.NewKey(&seed)
				passphrase = "edge123"
				err = am.CreateAccount(false, address2, passphrase)
				Expect(err).To(BeNil())

				account, err := am.GetDefault()
				Expect(err).To(BeNil())
				Expect(account).ToNot(BeNil())
				Expect(account.Address).To(Equal(address.Addr().String()))
			})

			It("should return ErrAccountNotFound if no address was found", func() {
				account, err := am.GetDefault()
				Expect(err).ToNot(BeNil())
				Expect(err).To(Equal(ErrAccountNotFound))
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
				err := am.CreateAccount(false, address, passphrase)
				Expect(err).To(BeNil())
				time.Sleep(1 * time.Second)

				seed = int64(2)
				address2, _ = crypto.NewKey(&seed)
				passphrase = "edge123"
				err = am.CreateAccount(false, address2, passphrase)
				Expect(err).To(BeNil())
			})

			It("should get accounts at index 0 and 1", func() {
				act, err := am.GetByIndex(0)
				Expect(err).To(BeNil())
				Expect(act.Address).To(Equal(address.Addr().String()))
				act, err = am.GetByIndex(1)
				Expect(act.Address).To(Equal(address2.Addr().String()))
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
				err := am.CreateAccount(false, address, passphrase)
				Expect(err).To(BeNil())
			})

			It("should successfully get account with address", func() {
				act, err := am.GetByAddress(address.Addr().String())
				Expect(err).To(BeNil())
				Expect(act.Address).To(Equal(address.Addr().String()))
			})

			It("should return err = 'account not found' when address does not exist", func() {
				_, err := am.GetByAddress("unknown_address")
				Expect(err).ToNot(BeNil())
				Expect(err).To(Equal(ErrAccountNotFound))
			})
		})
	})

	Describe(".GetBurnerAccountByAddress", func() {

		var key *crypto.Secp256k1Key
		am := New(accountPath)

		BeforeEach(func() {
			seed := int64(1)
			key, _ = crypto.NewSecp256k1(&seed, true, true)
			passphrase := "edge123"
			err := am.CreateBurnerAccount(key, passphrase)
			Expect(err).To(BeNil())
		})

		It("should successfully get account with address", func() {
			act, err := am.GetBurnerAccountByAddress(key.Addr())
			Expect(err).To(BeNil())
			Expect(act.Address).To(Equal(key.Addr()))
		})

		It("should return err = 'account not found' when address does not exist", func() {
			_, err := am.GetBurnerAccountByAddress("unknown_address")
			Expect(err).ToNot(BeNil())
			Expect(err).To(Equal(ErrAccountNotFound))
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
				err = am.CreateAccount(true, address, passphrase)
				Expect(err).To(BeNil())

				account, err = am.GetDefault()
				Expect(err).To(BeNil())
				Expect(account).ToNot(BeNil())
			})

			It("should return err = 'invalid password' when password is invalid", func() {
				err := account.Decrypt("invalid", false)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("invalid password"))
			})

			It("should return nil when decryption is successful. account.address must not be nil.", func() {
				err := account.Decrypt(passphrase, false)
				Expect(err).To(BeNil())
				Expect(account.key).ToNot(BeNil())
			})
		})
	})

	Describe("StoredAccountMeta", func() {
		Describe(".HasKey", func() {
			It("should return false when key does not exist", func() {
				sa := StoredAccount{meta: map[string]interface{}{}}
				r := sa.meta.HasKey("key")
				Expect(r).To(BeFalse())
			})

			It("should return true when key exist", func() {
				sa := StoredAccount{meta: map[string]interface{}{"key": 2}}
				r := sa.meta.HasKey("key")
				Expect(r).To(BeTrue())
			})
		})

		Describe(".Get", func() {
			It("should return nil when key does not exist", func() {
				sa := StoredAccount{meta: map[string]interface{}{}}
				r := sa.meta.Get("key")
				Expect(r).To(BeNil())
			})

			It("should return expected value when key exist", func() {
				sa := StoredAccount{meta: map[string]interface{}{"key": 2}}
				r := sa.meta.Get("key")
				Expect(r).To(Equal(2))
			})
		})
	})

})
