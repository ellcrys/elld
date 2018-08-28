package objects

import (
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/util"
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
	Address     util.String  `json:"address" msgpack:"address"`
	Balance     util.String  `json:"balance" msgpack:"balance"`
	AccountInfo *AccountInfo `json:"accountInfo" msgpack:"accountInfo"`
}

// ValidateAccount checks the fields of an Account object
// ensuring the value are valid and expected.
func ValidateAccount(account *Account) error {

	if len(account.Address) == 0 {
		return fieldError("address", "address is required")
	}

	if err := crypto.IsValidAddr(account.Address.String()); err != nil {
		return fieldError("address", err.Error())
	}

	return nil
}

// GetAddress gets the address
func (a *Account) GetAddress() util.String {
	return a.Address
}

// GetBalance gets the balance
func (a *Account) GetBalance() util.String {
	return a.Balance
}

// SetBalance sets the balance
func (a *Account) SetBalance(b util.String) {
	a.Balance = b
}
