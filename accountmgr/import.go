package accountmgr

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/ellcrys/druid/crypto"
	"github.com/fatih/color"

	funk "github.com/thoas/go-funk"
)

// ImportCmd takes a keyfile containing unencrypted password to create
// a new account. Keyfile must be a path to a file that exists.
// If pwd is provide and it is not a file path, it is used as
// the password. Otherwise, the file is read, trimmed of newline
// characters (left and right) and used as the password. When pwd
// is set, interactive password collection is not used.
func (am *AccountManager) ImportCmd(keyfile, pwd string) error {

	var passphrase string

	if keyfile == "" {
		printErr("Keyfile is required.")
		return fmt.Errorf("Keyfile is required")
	}

	// read keyfile content
	keyFileContent, err := ioutil.ReadFile(keyfile)
	if err != nil {
		if funk.Contains(err.Error(), "no such file") {
			printErr("Keyfile {%s} not found.", keyfile)
		}
		if funk.Contains(err.Error(), "is a directory") {
			printErr("Keyfile {%s} is a directory. Expects a file.", keyfile)
		}
		return err
	}

	// attempt to validate and instantiate the private key
	sk, err := crypto.PrivKeyFromBase58(string(keyFileContent))
	if err != nil {
		printErr("Keyfile contains invalid private key")
		return err
	}

	// if no password or password file is provided, ask for password
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
			if funk.Contains(err.Error(), "no such file") {
				printErr("Password file {%s} not found.", pwd)
			}
			if funk.Contains(err.Error(), "is a directory") {
				printErr("Password file path {%s} is a directory. Expects a file.", pwd)
			}
			return err
		}
		passphrase = string(content)
		passphrase = strings.TrimSpace(strings.Trim(passphrase, "/n"))
	} else if len(pwd) > 0 {
		passphrase = pwd
	}

	address := crypto.NewAddressFromPrivKey(sk)
	if err := am.createAccount(address, passphrase); err != nil {
		printErr(err.Error())
		return err
	}

	fmt.Println("Import successful. New account created, encrypted and stored")
	fmt.Println("Address:", color.CyanString(address.Addr()))

	return nil
}
