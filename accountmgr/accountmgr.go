package accountmgr

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"

	"github.com/btcsuite/btcutil/base58"
	funk "github.com/thoas/go-funk"
	"github.com/vmihailenco/msgpack"
	"golang.org/x/crypto/scrypt"

	"github.com/fatih/color"

	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/util"
	"github.com/segmentio/go-prompt"
)

var (
	accountEncryptionVersion = "0.0.1"
)

// PasswordPrompt reprents a function that can collect user input
type PasswordPrompt func(string, ...interface{}) string

// AccountManager defines functionalities to create,
// update, fetch and import accounts. An account encapsulates
// an address and private key and are stored in an encrypted format
// locally.
type AccountManager struct {
	accountDir  string
	getPassword PasswordPrompt
}

// New creates an account manager.
// accountDir is where encrypted account files are stored.
// Caller is expected to have created the accountDir before calling New
func New(accountDir string) *AccountManager {
	am := new(AccountManager)
	am.accountDir = accountDir
	am.getPassword = prompt.Password
	return am
}

// AskForPassword starts an interactive prompt to collect password.
// Returns error if password and repeated passwords do not match
func (am *AccountManager) AskForPassword() (string, error) {
	for {

		passphrase := am.getPassword("Passphrase")
		if len(passphrase) == 0 {
			continue
		}

		passphraseRepeat := am.getPassword("Repeat Passphrase")

		if passphrase != passphraseRepeat {
			return "", fmt.Errorf("Passphrases did not match")
		}

		return passphrase, nil
	}
}

// AskForPasswordOnce is like askForPassword but it does not
// ask to confirm password.
func (am *AccountManager) AskForPasswordOnce() (string, error) {
	for {

		passphrase := am.getPassword("Passphrase")
		if len(passphrase) == 0 {
			continue
		}

		return passphrase, nil
	}
}

// CreateAccount creates a new account
func (am *AccountManager) CreateAccount(address *crypto.Key, passphrase string) error {

	if address == nil {
		return fmt.Errorf("Address is required")
	}

	if passphrase == "" {
		return fmt.Errorf("Passphrase is required")
	}

	exist, err := am.AccountExist(address.Addr())
	if err != nil {
		return err
	}

	if exist {
		return fmt.Errorf("Account already exist")
	}

	// hash passphrase to get 32 bit encryption key
	passphraseHardened := hardenPassword([]byte(passphrase))

	// construct, json encode and encrypt account data
	acctDataBs, _ := msgpack.Marshal(map[string]string{
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

// CreateCmd creates a new account and interactively obtains
// encryption passphrase.
// If seed is non-zero, it is used. Otherwise, one will be randomly generated.
// If pwd is provide and it is not a file path, it is used as
// the password. Otherwise, the file is read, trimmed of newline
// characters (left and right) and used as the password. When pwd
// is set, interactive password collection is not used.
func (am *AccountManager) CreateCmd(seed int64, pwd string) (*crypto.Key, error) {

	var passphrase string
	var err error

	if len(pwd) == 0 {
		fmt.Println("Your new account needs to be locked with a password. Please enter a password.")
		passphrase, err = am.AskForPassword()
		if err != nil {
			printErr(err.Error())
			return nil, err
		}
	}

	// pwd is set and is a valid file, read it and use as password
	if len(pwd) > 0 && (os.IsPathSeparator(pwd[0]) || (len(pwd) >= 2 && pwd[:2] == "./")) {
		content, err := ioutil.ReadFile(pwd)
		if err != nil {
			if funk.Contains(err.Error(), "no such file") {
				printErr("Password file {%s} not found.", pwd)
			}
			if funk.Contains(err.Error(), "is a directory") {
				printErr("Password file path {%s} is a directory. Expects a file.", pwd)
			}
			return nil, err
		}
		passphrase = string(content)
		passphrase = strings.TrimSpace(strings.Trim(passphrase, "/n"))
	} else if len(pwd) > 0 {
		passphrase = pwd
	}

	// create address using random seed
	var address *crypto.Key
	if seed == 0 {
		rBytes := make([]byte, 32)
		io.ReadFull(rand.Reader, rBytes)
		seed = int64(binary.BigEndian.Uint64(rBytes))
	}
	address, err = crypto.NewKey(&seed)
	if err != nil {
		return nil, err
	}

	if err := am.CreateAccount(address, passphrase); err != nil {
		printErr(err.Error())
		return nil, err
	}

	fmt.Println("New account created, encrypted and stored")
	fmt.Println("Address:", color.CyanString(address.Addr()))

	return address, nil
}

func printErr(msg string, args ...interface{}) {
	fmt.Println(color.RedString("Error:"), fmt.Sprintf(msg, args...))
}

// harden improves a password's security and hardens it against
// bruteforce attacks by passing it to an RDF like scrypt.
func hardenPassword(pass []byte) []byte {

	passHash := sha256.Sum256(pass)
	var salt = passHash[16:]

	newPass, err := scrypt.Key(pass, salt, 32768, 8, 1, 32)
	if err != nil {
		panic(err)
	}

	return newPass
}
