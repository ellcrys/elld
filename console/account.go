package console

import (
	"github.com/ellcrys/elld/crypto"
	"github.com/robertkrimen/otto"
)

func (e *Executor) accountError(msg string) otto.Value {
	return e.vm.MakeCustomError("AccountError", msg)
}

// createAccount creates an account encrypted
// with the given passphrase
func (e *Executor) createAccount(passphrase string) string {
	key, _ := crypto.NewKey(nil)
	if err := e.acctMgr.CreateAccount(key, passphrase); err != nil {
		panic(e.accountError(err.Error()))
	}
	return key.Addr()
}

// loadAccount loads an account and
// sets it as the default account
func (e *Executor) loadAccount(address, password string) {

	// Get the account from the account manager
	sa, err := e.acctMgr.GetByAddress(address)
	if err != nil {
		panic(e.accountError(err.Error()))
	}

	if err := sa.Decrypt(password); err != nil {
		panic(e.accountError(err.Error()))
	}

	e.coinbase = sa.GetKey()
}

// loadedAccount returns the currently loaded account
func (e *Executor) loadedAccount() string {
	if e.coinbase == nil {
		return ""
	}
	return e.coinbase.Addr()
}

// importAccount creates an account with the given
// private key and encrypts with the passphrase.
func (e *Executor) importAccount(privateKey, passphrase string) string {

	privKey, err := crypto.PrivKeyFromBase58(privateKey)
	if err != nil {
		panic(e.accountError("invalid private key: " + err.Error()))
	}

	key := crypto.NewKeyFromPrivKey(privKey)
	if err := e.acctMgr.CreateAccount(key, passphrase); err != nil {
		panic(e.accountError(err.Error()))
	}

	return key.Addr()
}
