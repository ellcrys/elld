package wire

import (
	"github.com/ellcrys/elld/crypto"
)

const (
	// AccountTypeBalance represents a balance account
	AccountTypeBalance = iota

	// AccountTypeRepo represents a repo account
	AccountTypeRepo
)

// ValidateAccount checks the fields of an Account object
// ensuring the value are valid and expected.
func ValidateAccount(account *Account) error {

	if len(account.Address) == 0 {
		return fieldError("address", "address is required")
	}

	if err := crypto.IsValidAddr(account.Address); err != nil {
		return fieldError("address", err.Error())
	}

	return nil
}
