package accountmgr

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/btcsuite/btcutil/base58"
	"github.com/ellcrys/druid/crypto"
	"github.com/ellcrys/druid/util"
	funk "github.com/thoas/go-funk"
)

var (
	// ErrAccountNotFound represents an error about a missing account
	ErrAccountNotFound = fmt.Errorf("account not found")
)

// StoredAccount represents an encrypted account stored on disk
type StoredAccount struct {
	Address   string
	Cipher    []byte
	address   *crypto.Key
	CreatedAt time.Time
}

// AccountExist checks if an account with a matching address exists
func (am *AccountManager) AccountExist(address string) (bool, error) {

	accounts, err := am.GetAccountsOnDisk()
	if err != nil {
		return false, err
	}

	for _, acct := range accounts {
		if acct.Address == address {
			return true, nil
		}
	}

	return false, nil
}

// GetDefault gets the oldest account. Usually the account with 0 index.
func (am *AccountManager) GetDefault() (*StoredAccount, error) {

	accounts, err := am.GetAccountsOnDisk()
	if err != nil {
		return nil, err
	}

	if len(accounts) == 0 {
		return nil, nil
	}

	return accounts[0], nil
}

// GetByIndex returns an account by its current position in the
// list of accounts which is ordered by the time of creation.
func (am *AccountManager) GetByIndex(i int) (*StoredAccount, error) {

	accounts, err := am.GetAccountsOnDisk()
	if err != nil {
		return nil, err
	}

	if acctLen := len(accounts); acctLen-1 < i {
		return nil, ErrAccountNotFound
	}

	return accounts[i], nil
}

// GetByAddress gets an account by its address in the
// list of accounts which is ordered by the time of creation.
func (am *AccountManager) GetByAddress(addr string) (*StoredAccount, error) {

	accounts, err := am.GetAccountsOnDisk()
	if err != nil {
		return nil, err
	}

	account := funk.Find(accounts, func(x *StoredAccount) bool {
		return x.Address == addr
	})

	if account == nil {
		return nil, ErrAccountNotFound
	}

	return account.(*StoredAccount), nil
}

// GetAddress returns the address object which contains the private
// key and public key. Must call Decrypt() first.
func (sa *StoredAccount) GetAddress() *crypto.Key {
	return sa.address
}

// Decrypt decrypts the account cipher and initializes the address field
func (sa *StoredAccount) Decrypt(passphrase string) error {

	passphraseBs := hardenPassword([]byte(passphrase))
	acctBytes, err := util.Decrypt(sa.Cipher, passphraseBs[:])
	if err != nil {
		if funk.Contains(err.Error(), "invalid key") {
			return fmt.Errorf("invalid password. %s", err)
		}
		return err
	}

	// we expect a base58check content, verify it
	acctBytesBase58Dec, _, err := base58.CheckDecode(string(acctBytes))
	if err != nil {
		return fmt.Errorf("invalid password. %s", err)
	}

	// attempt to decode to ensure content is json encoded
	var accountData map[string]string
	if err := json.Unmarshal(acctBytesBase58Dec, &accountData); err != nil {
		return fmt.Errorf("unable to parse account data")
	}

	privKey, err := crypto.PrivKeyFromBase58(accountData["sk"])
	if err != nil {
		return err
	}

	sa.address = crypto.NewKeyFromPrivKey(privKey)
	return nil
}
