package accountmgr

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/btcsuite/btcutil/base58"
	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/util"
	funk "github.com/thoas/go-funk"
	"github.com/vmihailenco/msgpack"
)

// CreateBurnerAccount creates a new burner account.
func (am *AccountManager) CreateBurnerAccount(key *crypto.Secp256k1Key, passphrase string) error {

	// Ensure key is provided as well as the passphrase
	if key == nil {
		return fmt.Errorf("WIF structure is required")
	} else if passphrase == "" {
		return fmt.Errorf("Passphrase is required")
	}

	// Check whether this burner account has been previously created
	exist, err := am.BurnerAccountExist(key.Addr())
	if err != nil {
		return err
	} else if exist {
		return fmt.Errorf("An account with a matching seed already exist")
	}

	// hash passphrase to get 32 bit encryption key
	passphraseHardened := hardenPassword([]byte(passphrase))

	// Get the WIF structure from the key
	wif, err := key.WIF()
	if err != nil {
		return fmt.Errorf("failed to get wif struct from key: %s", err)
	}

	// construct, json encode and encrypt account data
	burnAccData, _ := msgpack.Marshal(map[string]string{
		"addr":    key.Addr(),
		"sk":      wif.String(),
		"testnet": "1",
		"v":       burnerAccountEncryptionVersion,
	})

	// base58 check encode
	b58AcctBs := base58.CheckEncode(burnAccData, 1)

	ct, err := util.Encrypt([]byte(b58AcctBs), passphraseHardened[:])
	if err != nil {
		return err
	}

	// Persist encrypted burner account data
	now := time.Now().Unix()
	name := fmt.Sprintf("%d_%s", now, key.Addr())
	if key.ForTestnet() {
		name += "_testnet"
	}
	fullpath := path.Join(filepath.Join(am.accountDir, config.BurnerAccountDirName), name)
	f, err := os.Create(fullpath)
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

// CreateBurnerCmd creates a new burner account and interactively obtains
// encryption passphrase. If seed is non-zero, it is used. Otherwise,
// one will be randomly generated. If pwd is provide and it is not a
// file path, it is used as the password. Otherwise, the file is read,
// trimmed of newline characters (left and right) and used as the password. When pwd
// is set, interactive password collection is not used.
// The testnet argument indicates the burner account is meant to be used
// on the burn chain testnet.
func (am *AccountManager) CreateBurnerCmd(seed int64, pwd string, testnet bool) (*crypto.Secp256k1Key, error) {

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

	// Create the key
	var key *crypto.Secp256k1Key
	key, err = crypto.NewSecp256k1(nil, testnet, true)
	if seed != 0 {
		key, err = crypto.NewSecp256k1(&seed, testnet, true)
	}

	// Create the burner account
	if err = am.CreateBurnerAccount(key, passphrase); err != nil {
		util.PrintCLIError("Failed to create burner account: %s", err)
		return nil, err
	}

	return key, nil
}
