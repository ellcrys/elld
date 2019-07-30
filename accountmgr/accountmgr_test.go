package accountmgr

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ellcrys/elld/config"

	"github.com/thoas/go-funk"

	"github.com/ellcrys/elld/crypto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func testPrompt(resp string) PasswordPrompt {
	return func(prompt string, args ...interface{}) string {
		return resp
	}
}

// testPrompt2 will return response with index equal to count
// count is incremented each time the function is called.
func testPrompt2(count *int, responses []string) PasswordPrompt {
	return func(prompt string, args ...interface{}) string {
		resp := responses[*count]
		*count++
		return resp
	}
}

var _ = Describe("AccountMgr", func() {

	path := filepath.Join("./", "test_cfg")
	accountPath := filepath.Join(path, config.AccountDirName)
	burnerAccountPath := filepath.Join(accountPath, config.BurnerAccountDirName)

	BeforeEach(func() {
		err := os.MkdirAll(accountPath, 0700)
		Expect(err).To(BeNil())
		err = os.MkdirAll(burnerAccountPath, 0700)
		Expect(err).To(BeNil())
	})

	Describe(".hardenPassword", func() {
		It("should return [215, 59, 34, 12, 157, 105, 253, 31, 243, 128, 41, 222, 216, 93, 165, 77, 67, 179, 85, 192, 127, 47, 171, 121, 32, 117, 125, 119, 109, 243, 32, 95]", func() {
			bs := hardenPassword([]byte("abc"))
			Expect(bs).To(Equal([]byte{215, 59, 34, 12, 157, 105, 253, 31, 243, 128, 41, 222, 216, 93, 165, 77, 67, 179, 85, 192, 127, 47, 171, 121, 32, 117, 125, 119, 109, 243, 32, 95}))
		})
	})

	Describe(".askForPassword", func() {
		am := New(accountPath)

		It("should return err = 'Passphrases did not match' when passphrase and repeat passphrase don't match", func() {
			count := 0
			am.getPassword = testPrompt2(&count, []string{"passAbc", "passAb"})
			_, err := am.AskForPassword()
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("Passphrases did not match"))
		})

		It("should return input even when no password is provided the first time", func() {
			count := 0
			am.getPassword = testPrompt2(&count, []string{"", "passAb", "passAb"})
			password, err := am.AskForPassword()
			Expect(err).To(BeNil())
			Expect(password).To(Equal("passAb"))
		})
	})

	Describe(".askForPasswordOnce", func() {
		am := New(accountPath)

		It("should return the first input received", func() {
			count := 0
			am.getPassword = testPrompt2(&count, []string{"", "", "passAb"})
			password, err := am.AskForPasswordOnce()
			Expect(err).To(BeNil())
			Expect(password).To(Equal("passAb"))
		})
	})

	Describe(".CreateAccount", func() {
		am := New(accountPath)

		It("should return err = 'Address is required' when address is nil", func() {
			err := am.CreateAccount(nil, "")
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("Address is required"))
		})

		It("should return err = 'Passphrase is required' when passphrase is empty", func() {
			seed := int64(1)
			address, _ := crypto.NewKey(&seed)
			err := am.CreateAccount(address, "")
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("Passphrase is required"))
		})

		It("should return nil when account has been created", func() {
			seed := int64(1)
			address, _ := crypto.NewKey(&seed)
			passphrase := "edge123"
			err := am.CreateAccount(address, passphrase)
			Expect(err).To(BeNil())
		})

		When("account has been created", func() {

			var address *crypto.Key

			BeforeEach(func() {
				seed := int64(1)
				address, _ = crypto.NewKey(&seed)
				passphrase := "edge123"
				err := am.CreateAccount(address, passphrase)
				Expect(err).To(BeNil())
			})

			It("should have a keyfile in the account directory", func() {
				kfs, err := ioutil.ReadDir(accountPath)
				Expect(err).To(BeNil())
				found := funk.Find(kfs, func(x os.FileInfo) bool {
					return funk.Contains(x.Name(), address.Addr())
				})
				Expect(found).ToNot(BeNil())
			})

			It("should return err = 'An account with a matching seed already exist' when account with same address already exist", func() {
				seed := int64(1)
				address, _ = crypto.NewKey(&seed)
				passphrase := "edge123"
				err := am.CreateAccount(address, passphrase)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("An account with a matching seed already exist"))
			})
		})
	})

	Describe(".CreateBurnerAccount", func() {
		am := New(accountPath)

		It("should return err = 'Address is required' when address is nil", func() {
			err := am.CreateBurnerAccount(nil, "")
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("WIF structure is required"))
		})

		It("should return err = 'Passphrase is required' when passphrase is empty", func() {
			seed := int64(1)
			key, _ := crypto.NewSecp256k1(&seed, true, false)
			err := am.CreateBurnerAccount(key, "")
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("Passphrase is required"))
		})

		It("should return nil when account has been created", func() {
			seed := int64(1)
			key, _ := crypto.NewSecp256k1(&seed, true, false)
			passphrase := "edge123"
			err := am.CreateBurnerAccount(key, passphrase)
			Expect(err).To(BeNil())
		})

		When("account has been created", func() {

			var key *crypto.Secp256k1Key

			BeforeEach(func() {
				seed := int64(1)
				key, _ = crypto.NewSecp256k1(&seed, true, false)
				passphrase := "edge123"
				err := am.CreateBurnerAccount(key, passphrase)
				Expect(err).To(BeNil())
			})

			It("should have a keyfile in the burner account directory", func() {
				kfs, err := ioutil.ReadDir(burnerAccountPath)
				Expect(err).To(BeNil())
				found := funk.Find(kfs, func(x os.FileInfo) bool {
					return funk.Contains(x.Name(), key.Addr())
				})
				Expect(found).ToNot(BeNil())
			})

			It("should return err = 'An account with a matching seed already exist' when account with same address already exist", func() {
				seed := int64(1)
				key, _ = crypto.NewSecp256k1(&seed, true, false)
				passphrase := "edge123"
				err := am.CreateBurnerAccount(key, passphrase)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("An account with a matching seed already exist"))
			})
		})
	})

	AfterEach(func() {
		err := os.RemoveAll(path)
		Expect(err).To(BeNil())
	})

})
