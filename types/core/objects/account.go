package objects

import (
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
	Nonce       uint64       `json:"nonce" msgpack:"nonce"`
	AccountInfo *AccountInfo `json:"accountInfo,omitempty" msgpack:"accountInfo"`
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

// GetNonce gets the nonce
func (a *Account) GetNonce() uint64 {
	return a.Nonce
}

// IncrNonce increments the nonce by 1
func (a *Account) IncrNonce() {
	a.Nonce++
}
