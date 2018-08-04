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

// AccountInfo represents the data specific to a regular account
type AccountInfo struct {
}

// Account represents an entity on the network.
type Account struct {
	Type        int32        `json:"type" msgpack:"type"`
	Address     string       `json:"address" msgpack:"address"`
	Balance     string       `json:"balance" msgpack:"balance"`
	AccountInfo *AccountInfo `json:"accountInfo" msgpack:"accountInfo"`
}

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
