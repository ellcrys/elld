package burner

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/ellcrys/elld/crypto"

	"github.com/ellcrys/elld/params"

	"github.com/shopspring/decimal"

	"github.com/ellcrys/go-prompt"

	"github.com/ellcrys/elld/accountmgr"
	"github.com/ellcrys/elld/ltcsuite/ltcd/rpcclient"
	"github.com/ellcrys/elld/ltcsuite/ltcutil"
	"github.com/ellcrys/elld/util"
	"github.com/thoas/go-funk"
)

// BurnCmd burns processes a burn request.
// The coinbase argument is the account of the beneficiary of the burned amount.
// The burnerAccount account argument is the account to burn from.
// If pwd is provided, it is used as the password to unlock the burner
// account. Otherwise, an interactive session is started to collect
// the password.
func BurnCmd(coinbase *accountmgr.StoredAccount,
	burnerAccount *accountmgr.StoredAccount, pwd, amount string) error {

	var content []byte
	var err error
	var fullPath, passphrase string

	// if no password or password file is provided, ask for password
	if len(pwd) == 0 {
		fmt.Println("The account needs to be unlocked. Please enter a password.")
		passphrase = prompt.Password("Passphrase")
		goto burn
	}

	// If pwd is not a path to a file,
	// use pwd as the passphrase.
	if !strings.HasPrefix(pwd, "./") && !strings.HasPrefix(pwd, "/") && filepath.Ext(pwd) == "" {
		passphrase = pwd
		goto burn
	}

	fullPath, err = filepath.Abs(pwd)
	if err != nil {
		util.PrintCLIError("Invalid file path {%s}: %s", pwd, err.Error())
		return err
	}

	content, err = ioutil.ReadFile(fullPath)
	if err != nil {
		if funk.Contains(err.Error(), "no such file") {
			util.PrintCLIError("Password file {%s} not found.", pwd)
		}
		if funk.Contains(err.Error(), "is a directory") {
			util.PrintCLIError("Password file path {%s} is a directory. Expects a file.", pwd)
		}
		return err
	}
	passphrase = strings.TrimSpace(strings.Trim(string(content), "/n"))

burn:

	// Decrypt the burn account
	err = burnerAccount.Decrypt(passphrase, true)
	if err != nil {
		util.PrintCLIError("Password is not valid")
		return err
	}

	// Get the WIF key.
	wifKey := burnerAccount.GetKey().(*ltcutil.WIF)

	// Ensure the amount is up to the minimum burn amount.
	amt, err := decimal.NewFromString(amount)
	if err != nil {
		return err
	} else if amt.LessThan(params.MinimumBurnAmt) {
		err = fmt.Errorf("Burn amount is below the minimum (%s)", params.MinimumBurnAmt.String())
		util.PrintCLIError(err.Error())
		return err
	}

	// Create an OP_RETURN transaction
	err = burnWithWIF(wifKey)
	if err != nil {
		util.PrintCLIError("Failed to burn coins: %s", err)
		return err
	}

	// pp.Println(wifKey.String())

	return nil
}

// GetClient returns a client to the burner chain RPC server
func GetClient(host, rpcUser, rpcPass string, disableTLS bool) (*rpcclient.Client, error) {
	connCfg := &rpcclient.ConnConfig{
		Host:       host,
		Endpoint:   "ws",
		User:       rpcUser,
		Pass:       rpcPass,
		DisableTLS: disableTLS,
	}

	client, err := rpcclient.New(connCfg, nil)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// burnWithWIF executes the burn operation
func burnWithWIF(wif *ltcutil.WIF) error {

	connCfg := &rpcclient.ConnConfig{
		Host:       "localhost:19334",
		Endpoint:   "ws",
		User:       "admin",
		Pass:       "admin",
		DisableTLS: true,
	}

	client, err := rpcclient.New(connCfg, nil)
	if err != nil {
		return err
	}

	// Get the address from the WIF
	addr, err := crypto.WIFToAddress(wif)
	if err != nil {
		return fmt.Errorf("failed to get address from WIF: %s", err)
	}

	_ = client
	_ = addr

	return nil
}
