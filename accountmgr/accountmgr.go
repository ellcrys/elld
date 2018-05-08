package accountmgr

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"

	"github.com/btcsuite/btcutil/base58"

	"github.com/fatih/color"

	"github.com/ellcrys/druid/crypto"
	"github.com/ellcrys/druid/util"
	"github.com/segmentio/go-prompt"
)

var accountEncryptionVersion = "0.0.1"

// AccountManager defines functionalities to create,
// update, fetch and import accounts. An account encapsulates
// an address and private key and are stored in an encrypted format
// locally.
type AccountManager struct {
	accountDir string
}

// New creates an account manager.
// accountDir is where encrypted account files are stored.
// Caller is expected to have created the accountDir before calling New
func New(accountDir string) *AccountManager {
	am := new(AccountManager)
	am.accountDir = accountDir
	return am
}

// askForPassword starts an interactive prompt to collect password.
// Returns error if password and repeated passwords do not match
func (am *AccountManager) askForPassword() (string, error) {
	for {

		passphrase := prompt.Password("Passphrase")
		if len(passphrase) == 0 {
			continue
		}

		passphraseRepeat := prompt.Password("Repeat Passphrase")

		if passphrase != passphraseRepeat {
			fmt.Println("Passphrases did not match")
			return "", fmt.Errorf("Passphrases did not match")
		}

		return passphrase, nil
	}
}

// askForPasswordOnce is like askForPassword but it does not
// ask to confirm password.
func (am *AccountManager) askForPasswordOnce() (string, error) {
	for {

		passphrase := prompt.Password("Passphrase")
		if len(passphrase) == 0 {
			continue
		}

		return passphrase, nil
	}
}

// createAccount creates a new account
func (am *AccountManager) createAccount(address *crypto.Address, passphrase string) error {

	if address == nil {
		return fmt.Errorf("address is required")
	}

	if passphrase == "" {
		return fmt.Errorf("passphrase is required")
	}

	// hash passphrase to get 32 bit encryption key
	passphraseHardened := hardenPassword([]byte(passphrase))

	// construct, json encode and encrypt account data
	acctDataBs, _ := json.Marshal(map[string]string{
		"addr": address.Addr(),
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

	// persist encrypted account data
	now := time.Now().Unix()
	fileName := path.Join(am.accountDir, fmt.Sprintf("%d_%s", now, address.Addr()))
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

// Create creates a new account and interactively obtains
// encryption passphrase.
// If pwd is provide and it is not a file path, it is used as
// the password. Otherwise, the file is read, trimmed of newline
// characters (left and right) and used as the password. When pwd
// is set, interactive password collection is not used.
func (am *AccountManager) Create(pwd string) error {

	var passphrase string
	var err error

	if len(pwd) == 0 {
		fmt.Println("Your new account needs to be locked with a password. Please enter a password.")
		passphrase, err = am.askForPassword()
		if err != nil {
			return err
		}
	}

	// pwd is set and is a valid file, read it and use as password
	if len(pwd) > 0 && (os.IsPathSeparator(pwd[0]) || pwd[:2] == "./") {
		content, err := ioutil.ReadFile(pwd)
		if err != nil {
			return err
		}
		passphrase = string(content)
		passphrase = strings.TrimSpace(strings.Trim(passphrase, "/n"))
	} else if len(pwd) > 0 {
		passphrase = pwd
	}

	// create address using random seed
	seed := make([]byte, 32)
	io.ReadFull(rand.Reader, seed)
	var seedUint64 = int64(binary.BigEndian.Uint64(seed))
	address, err := crypto.NewAddress(&seedUint64)
	if err != nil {
		return err
	}

	if err := am.createAccount(address, passphrase); err != nil {
		return err
	}

	fmt.Println("New account created, encrypted and stored")
	fmt.Println("Address:", color.CyanString(address.Addr()))

	return nil
}

func printErr(msg string, args ...interface{}) {
	fmt.Println(color.RedString("Error:"), fmt.Sprintf(msg, args...))
}

// harden improves a password passed by a user.
// TODO: use a proper KDF
func hardenPassword(pass []byte) [32]byte {
	return sha256.Sum256(pass)
}
