package console

import (
	"fmt"

	"github.com/fatih/color"

	"github.com/ellcrys/elld/crypto"
	goprompt "github.com/ellcrys/go-prompt"
	"github.com/robertkrimen/otto"
)

func (e *Executor) accountError(msg string) otto.Value {
	return e.vm.MakeCustomError("AccountError", msg)
}

// createAccount creates an account encrypted
// with the given passphrase
func (e *Executor) createAccount(passphrase ...string) string {

	var pass string
	if len(passphrase) != 0 {
		pass = passphrase[0]
	}

	// When password is not provided, we assume the
	// caller intends to enter interactive mode.
	// Prompt user to enter passphrase util she does.
	if len(pass) == 0 {
		fmt.Println("Please enter your passphrase below:")
		fmt.Println(color.HiMagentaString("Warning: Do not forget your " +
			"passphrase, you will not be able to access your account if you do."))
		for len(pass) == 0 {
			pass = goprompt.Password("Passphrase")
		}
		repeatPass := goprompt.Password("Repeat Passphrase")
		if pass != repeatPass {
			panic(e.accountError("Passphrase and Repeated passphrase do not match"))
		}
	}

	key, _ := crypto.NewKey(nil)
	if err := e.acctMgr.CreateAccount(false, key, pass); err != nil {
		panic(e.accountError(err.Error()))
	}

	return key.Addr().String()
}

// loadAccount loads an account and
// sets it as the default account
func (e *Executor) loadAccount(address string, optionalArgs ...string) {

	var passphrase string

	if len(optionalArgs) == 1 {
		passphrase = optionalArgs[0]
	}

	// When passphrase is not provided, we assume the
	// caller intends to enter interactive mode.
	// Prompt user to enter passphrase util she does.
	if len(passphrase) == 0 {
		fmt.Println("Please enter your passphrase below:")
		for len(passphrase) == 0 {
			passphrase = goprompt.Password("Passphrase")
		}
	}

	// Get the account from the account manager
	sa, err := e.acctMgr.GetByAddress(address)
	if err != nil {
		panic(e.accountError(err.Error()))
	}

	// Decrypt the account
	if err := sa.Decrypt(passphrase, false); err != nil {
		panic(e.accountError(err.Error()))
	}

	e.coinbase = sa.GetKey().(*crypto.Key)
}

// loadedAccount returns the currently loaded account
func (e *Executor) loadedAccount() string {
	if e.coinbase == nil {
		return ""
	}
	return e.coinbase.Addr().String()
}

// listLocalAccounts returns a list of
// addresses of accounts that exist on the node
func (e *Executor) listLocalAccounts() (addresses []string) {
	accounts, _ := e.acctMgr.ListAccounts()
	for _, a := range accounts {
		addresses = append(addresses, a.Address)
	}
	return
}

// importAccount creates an account with the given
// private key and encrypts with the passphrase.
func (e *Executor) importAccount(privateKey string, optionalArgs ...string) string {

	privKey, err := crypto.PrivKeyFromBase58(privateKey)
	if err != nil {
		panic(e.accountError("invalid private key: " + err.Error()))
	}

	var passphrase string
	if len(optionalArgs) == 1 {
		passphrase = optionalArgs[0]
	}

	// When passphrase is not provided, we assume the
	// caller intends to enter interactive mode.
	// Prompt user to enter passphrase util she does.
	if len(passphrase) == 0 {
		fmt.Println("Please enter your passphrase below:")
		fmt.Println(color.HiMagentaString("Warning: Do not forget your " +
			"passphrase, you will not be able to access your account if you do."))
		for len(passphrase) == 0 {
			passphrase = goprompt.Password("Passphrase")
		}
		repeatPass := goprompt.Password("Repeat Passphrase")
		if passphrase != repeatPass {
			panic(e.accountError("Passphrase and Repeated passphrase do not match"))
		}
	}

	key := crypto.NewKeyFromPrivKey(privKey)
	if err := e.acctMgr.CreateAccount(false, key, passphrase); err != nil {
		panic(e.accountError(err.Error()))
	}

	return key.Addr().String()
}
