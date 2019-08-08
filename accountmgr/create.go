package accountmgr

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"

	"github.com/btcsuite/btcutil/base58"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/util"
	funk "github.com/thoas/go-funk"
	"github.com/vmihailenco/msgpack"
)

// HasDefaultAccount checks whether a default account exists
func HasDefaultAccount(am *AccountManager) bool {
	filePath := path.Join(am.accountDir, "default")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	}
	return true
}

// CreateAccount creates a new account
func (am *AccountManager) CreateAccount(defaultAccount bool, address *crypto.Key,
	passphrase string) error {

	if address == nil {
		return fmt.Errorf("Address is required")
	} else if passphrase == "" {
		return fmt.Errorf("Passphrase is required")
	}

	exist, err := am.AccountExist(address.Addr().String())
	if err != nil {
		return err
	}

	if exist {
		return fmt.Errorf("An account with a matching seed already exist")
	}

	// hash passphrase to get 32 bit encryption key
	passphraseHardened := hardenPassword([]byte(passphrase))

	// construct, json encode and encrypt account data
	acctDataBs, _ := msgpack.Marshal(map[string]string{
		"addr": address.Addr().String(),
		"sk":   address.PrivKey().Base58(),
		"pk":   address.PubKey().Base58(),
		"v":    accountEncryptionVersion,
	})

	// base58 check encode
	b58AcctBs := base58.CheckEncode(acctDataBs, 1)

	ct, err := util.Encrypt([]byte(b58AcctBs), passphraseHardened[:])
	if err != nil {
		return err
	}

	// Persist encrypted account data
	now := time.Now().Unix()
	fileName := path.Join(am.accountDir, fmt.Sprintf("%d_%s", now, address.Addr()))
	if defaultAccount {
		fileName = path.Join(am.accountDir, fmt.Sprintf("%d_%s_default", now, address.Addr()))
	}

	f, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(ct)
	if err != nil {
		return err
	}

	return nil
}

// CreateCmd creates a new account and interactively obtains encryption passphrase.
// defaultAccount argument indicates that this account should be marked as default.
// If seed is non-zero, it is used. Otherwise, one will be randomly generated.
// If pwd is provide and it is not a file path, it is used as
// the password. Otherwise, the file is read, trimmed of newline
// characters (left and right) and used as the password. When pwd
// is set, interactive password collection is not used.
func (am *AccountManager) CreateCmd(defaultAccount bool, seed int64, pwd string) (*crypto.Key, error) {

	var passphrase string
	var err error

	// If no password is provided, we start an interactive session to
	// collect the password or passphrase
	if len(pwd) == 0 {
		fmt.Println("Your new account needs to be locked with a password. Please enter a password.")
		passphrase, err = am.AskForPassword()
		if err != nil {
			util.PrintCLIError(err.Error())
			return nil, err
		}
	}

	// But if the password is set and is a valid file, read it and use as password
	if len(pwd) > 0 && (os.IsPathSeparator(pwd[0]) || (len(pwd) >= 2 && pwd[:2] == "./")) {
		content, err := ioutil.ReadFile(pwd)
		if err != nil {
			if funk.Contains(err.Error(), "no such file") {
				util.PrintCLIError("Password file {%s} not found.", pwd)
			}
			if funk.Contains(err.Error(), "is a directory") {
				util.PrintCLIError("Password file path {%s} is a directory. Expects a file.", pwd)
			}
			return nil, err
		}
		passphrase = string(content)
		passphrase = strings.TrimSpace(strings.Trim(passphrase, "/n"))
	} else if len(pwd) > 0 {
		passphrase = pwd
	}

	// Generate an address (which includes a private key)
	var address *crypto.Key
	address, err = crypto.NewKey(nil)
	if seed != 0 {
		address, err = crypto.NewKey(&seed)
	}

	if err != nil {
		return nil, err
	}

	// Create and encrypted the account on disk
	if err := am.CreateAccount(defaultAccount, address, passphrase); err != nil {
		util.PrintCLIError(err.Error())
		return nil, err
	}

	return address, nil
}
