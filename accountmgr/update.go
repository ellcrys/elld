package accountmgr

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/btcsuite/btcutil/base58"
	"github.com/vmihailenco/msgpack"

	"github.com/ellcrys/elld/util"
	"github.com/thoas/go-funk"
)

// UpdateCmd fetches and lists all accounts
func (am *AccountManager) UpdateCmd(address string) error {

	if len(address) == 0 {
		printErr("Address is required")
		return fmt.Errorf("Address is required")
	}

	// find the account with a matching address
	accounts, err := am.ListAccounts()
	if err != nil {
		return err
	}

	account := funk.Find(accounts, func(x *StoredAccount) bool {
		return x.Address == address
	})

	if account == nil {
		printErr("Account {%s} does not exist", address)
		return fmt.Errorf("account not found")
	}

	// collect account password
	fmt.Println(fmt.Sprintf("Enter your current password for the account {%s}", address))
	passphrase, err := am.AskForPasswordOnce()
	if err != nil {
		return err
	}

	// attempt to decrypt the account
	passphraseBs := hardenPassword([]byte(passphrase))
	acctBytes, err := util.Decrypt(account.(*StoredAccount).Cipher, passphraseBs[:])
	if err != nil {
		if funk.Contains(err.Error(), "invalid key") {
			printErr("Invalid password. Could not unlock account.")
			return err
		}
		return err
	}

	// we expect a base58check content, verify it
	acctBytesBase58Dec, _, err := base58.CheckDecode(string(acctBytes))
	if err != nil {
		printErr("Invalid password. Could not unlock account.")
		return err
	}

	// attempt to decode to ensure content is json encoded
	var accountData map[string]string
	if err := msgpack.Unmarshal(acctBytesBase58Dec, &accountData); err != nil {
		printErr("Unable to parse unlocked account data")
		return err
	}

	// collect new password
	fmt.Println("Enter your new password")
	newPassphrase, err := am.AskForPassword()
	if err != nil {
		printErr(err.Error())
		return err
	}

	// re-encrypt with new password
	newPassphraseHardened := hardenPassword([]byte(newPassphrase))
	updatedCipher, err := util.Encrypt(acctBytes, newPassphraseHardened[:])
	if err != nil {
		return err
	}

	acct := account.(*StoredAccount)
	filename := filepath.Join(am.accountDir, fmt.Sprintf("%d_%s", acct.CreatedAt.Unix(), acct.Address))
	err = ioutil.WriteFile(filename, updatedCipher, 0644)
	if err != nil {
		printErr("Unable to persist cipher. %s", err.Error())
		return err
	}

	fmt.Println("Successfully updated account")
	return nil
}
