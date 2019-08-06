package accountmgr

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/ellcrys/elld/util"

	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/ltcsuite/ltcutil"
	"github.com/fatih/color"

	funk "github.com/thoas/go-funk"
)

// ImportBurnerCmd takes a keyfile containing a Litecoin WIF key to create
// a new burner account. Keyfile must be a path to a file that exists.
// If pwd is provide and it is not a file path, it is used as
// the password. Otherwise, the file is read, trimmed of newline
// characters (left and right) and used as the password. When pwd
// is set, interactive password collection is not used.
// The testnet argument indicates that the burner account should be created
// according to the burn chain's testnet address format
func (am *AccountManager) ImportBurnerCmd(keyfile, pwd string, testnet bool) error {

	if keyfile == "" {
		util.PrintCLIError("Keyfile is required.")
		return fmt.Errorf("Keyfile is required")
	}

	fullKeyfilePath, err := filepath.Abs(keyfile)
	if err != nil {
		util.PrintCLIError("Invalid keyfile path {%s}", keyfile)
		return fmt.Errorf("Invalid keyfile path")
	}

	keyFileContent, err := ioutil.ReadFile(fullKeyfilePath)
	if err != nil {
		if funk.Contains(err.Error(), "no such file") {
			util.PrintCLIError("Keyfile {%s} not found.", keyfile)
		}
		if funk.Contains(err.Error(), "is a directory") {
			util.PrintCLIError("Keyfile {%s} is a directory. Expects a file.", keyfile)
		}
		return err
	}

	// Attempt to validate and instantiate a ltcutil.WIF instance
	fileContentStr := strings.TrimSpace(string(keyFileContent))
	wif, err := ltcutil.DecodeWIF(fileContentStr)
	if err != nil {
		util.PrintCLIError("File content is not a valid WIF key")
		return fmt.Errorf("File content is not a valid WIF key")
	}

	var content []byte

	// if no password or password file is provided, ask for password
	passphrase := ""
	if len(pwd) == 0 {
		fmt.Println("Your new account needs to be locked with a password. Please enter a password.")
		passphrase, err = am.AskForPassword()
		if err != nil {
			util.PrintCLIError(err.Error())
			return err
		}
		goto create
	}

	if !strings.HasPrefix(pwd, "./") && !strings.HasPrefix(pwd, "/") && filepath.Ext(pwd) == "" {
		passphrase = pwd
		goto create
	}

	content, err = ioutil.ReadFile(pwd)
	if err != nil {
		if funk.Contains(err.Error(), "no such file") {
			util.PrintCLIError("Password file {%s} not found.", pwd)
		}
		if funk.Contains(err.Error(), "is a directory") {
			util.PrintCLIError("Password file path {%s} is a directory. Expects a file.", pwd)
		}
		return err
	}
	passphrase = string(content)
	passphrase = strings.TrimSpace(strings.Trim(passphrase, "/n"))

create:

	key := crypto.NewSecp256k1FromWIF(wif, testnet, true)

	if err := am.CreateBurnerAccount(key, passphrase); err != nil {
		util.PrintCLIError(err.Error())
		return err
	}

	fmt.Println("Import successful. New burner account created, encrypted and stored.")
	fmt.Println("Address:", color.CyanString(key.Addr()))

	return nil
}
