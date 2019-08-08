// Package accountmgr provides account creation and management
// functionalities.
package accountmgr

import (
	"crypto/sha256"
	"fmt"

	"golang.org/x/crypto/scrypt"

	"github.com/ellcrys/go-prompt"
)

var (
	accountEncryptionVersion       = "0.0.1"
	burnerAccountEncryptionVersion = "0.0.1"
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
