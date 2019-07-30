package accountmgr

import (
	"fmt"
	"time"

	"github.com/btcsuite/btcutil/base58"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/util"
	funk "github.com/thoas/go-funk"
	"github.com/vmihailenco/msgpack"
)

var (
	// ErrAccountNotFound represents an error about a missing account
	ErrAccountNotFound = fmt.Errorf("account not found")
)

// StoredAccount represents an encrypted account stored on disk
type StoredAccount struct {
	Address   string
	Cipher    []byte
	key       *crypto.Key
	CreatedAt time.Time
}

// AccountExist checks if an account with a matching address exists
func (am *AccountManager) AccountExist(address string) (bool, error) {

	accounts, err := am.ListAccounts()
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

// BurnerAccountExist checks if a burner account with a matching address exists
func (am *AccountManager) BurnerAccountExist(address string) (bool, error) {

	accounts, err := am.ListBurnerAccounts()
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

	accounts, err := am.ListAccounts()
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

	accounts, err := am.ListAccounts()
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

	accounts, err := am.ListAccounts()
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
	return sa.key
}

// GetKey gets the decrypted key
func (sa *StoredAccount) GetKey() *crypto.Key {
	return sa.key
}

// Decrypt decrypts the account cipher and initializes the address field
func (sa *StoredAccount) Decrypt(passphrase string) error {

	passphraseBs := hardenPassword([]byte(passphrase))
	acctBytes, err := util.Decrypt(sa.Cipher, passphraseBs[:])
	if err != nil {
		if funk.Contains(err.Error(), "invalid key") {
			return fmt.Errorf("invalid password")
		}
		return err
	}

	// we expect a base58check content, verify it
	acctData, _, err := base58.CheckDecode(string(acctBytes))
	if err != nil {
		return fmt.Errorf("invalid password")
	}

	// attempt to decode to ensure content is json encoded
	var accountData map[string]string
	if err := msgpack.Unmarshal(acctData, &accountData); err != nil {
		return fmt.Errorf("unable to parse account data")
	}

	privKey, err := crypto.PrivKeyFromBase58(accountData["sk"])
	if err != nil {
		return err
	}

	sa.key = crypto.NewKeyFromPrivKey(privKey)
	return nil
}
