package accountmgr

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/ellcrys/elld/util"

	"github.com/ellcrys/elld/ltcsuite/ltcutil"

	"github.com/fatih/color"
	funk "github.com/thoas/go-funk"
)

// RevealBurnerCmd decrypts a burner account and outputs the WIF key.
// If pwd is provide and it is not a file path, it is used as
// the password. Otherwise, the file is read, trimmed of newline
// characters (left and right) and used as the password. When pwd
// is set, interactive password collection is not initialized.
func (am *AccountManager) RevealBurnerCmd(address, pwd string) error {

	var passphrase string

	if address == "" {
		util.PrintCLIError("Address is required.")
		return fmt.Errorf("address is required")
	}

	storedAcct, err := am.GetBurnerAccountByAddress(address)
	if err != nil {
		util.PrintCLIError(err.Error())
		return err
	}

	var content []byte
	var fullPath string

	// if no password or password file is provided, ask for password
	if len(pwd) == 0 {
		fmt.Println("The account needs to be unlocked. Please enter a password.")
		passphrase, err = am.AskForPasswordOnce()
		if err != nil {
			util.PrintCLIError(err.Error())
			return err
		}
		goto unlock
	}

	// If pwd is not a path to a file,
	// use pwd as the passphrase.
	if !strings.HasPrefix(pwd, "./") && !strings.HasPrefix(pwd, "/") && filepath.Ext(pwd) == "" {
		passphrase = pwd
		goto unlock
	}

	fullPath, err = filepath.Abs(pwd)
	if err != nil {
		util.PrintCLIError("Invalid file path {%s}: %s", pwd, err.Error())
		return err
	}

	content, err = ioutil.ReadFile(fullPath)
	if err != nil {
		if funk.Contains(err.Error(), "no such file") {
			util.PrintCLIError("Password file {%s} not found.", pwd)
		}
		if funk.Contains(err.Error(), "is a directory") {
			util.PrintCLIError("Password file path {%s} is a directory. Expects a file.", pwd)
		}
		return err
	}
	passphrase = strings.TrimSpace(strings.Trim(string(content), "/n"))

unlock:

	if err = storedAcct.Decrypt(passphrase, true); err != nil {
		util.PrintCLIError("Invalid password. Could not unlock account.")
		return err
	}

	fmt.Println(color.HiCyanString("Private Key (WIF):"), storedAcct.key.(*ltcutil.WIF).String())

	return nil
}
