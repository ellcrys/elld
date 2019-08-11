package burner

import (
	"fmt"
	"io/ioutil"
	"net"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ellcrys/elld/types"
	"github.com/fatih/color"

	"github.com/ellcrys/elld/rpc/jsonrpc"

	"github.com/ellcrys/elld/rpc/client"

	"github.com/ellcrys/elld/config"

	"github.com/ellcrys/go-prompt"

	"github.com/ellcrys/elld/accountmgr"
	"github.com/ellcrys/elld/ltcsuite/ltcd/rpcclient"
	"github.com/ellcrys/elld/ltcsuite/ltcutil"
	"github.com/ellcrys/elld/util"
	"github.com/thoas/go-funk"
)

// BurnCmd burns processes a burn request.
// The nodeKey argument is the account of the beneficiary of the burned amount.
// The burnerAccount account argument is the account to burn from.
// If pwd is provided, it is used as the password to unlock the burner
// account. Otherwise, an interactive session is started to collect
// the password.
func BurnCmd(cfg *config.EngineConfig,
	nodeKey *accountmgr.StoredAccount,
	rpcAddress string,
	burnerAccount *accountmgr.StoredAccount,
	pwd, producerAddress, amount string,
	fee float64) error {

	var content []byte
	var err error
	var fullPath, passphrase string

	// If no password or password file is provided, ask for password
	if len(pwd) == 0 {
		fmt.Println("The account needs to be unlocked. Please enter a password.")
		passphrase = prompt.Password("Passphrase")
		goto burn
	}

	// If pwd is not a path to a file, use pwd as the passphrase.
	if !strings.HasPrefix(pwd, "./") && !strings.HasPrefix(pwd, "/") && filepath.Ext(pwd) == "" {
		passphrase = pwd
		goto burn
	}

	// Construct the full-length file path of password,
	fullPath, err = filepath.Abs(pwd)
	if err != nil {
		util.PrintCLIError("Invalid file path {%s}: %s", pwd, err.Error())
		return err
	}

	// The the password file content
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

	// Trim the password of unwanted trailing spaces
	passphrase = strings.TrimSpace(strings.Trim(string(content), "/n"))

burn:

	// Decrypt the burn account using the provided passphrase
	err = burnerAccount.Decrypt(passphrase, true)
	if err != nil {
		util.PrintCLIError("Password is not valid")
		return err
	}

	// Get the WIF key of the burn account
	wifKey := burnerAccount.GetKey().(*ltcutil.WIF)

	// Create an OP_RETURN transaction
	txHash, err := burn(rpcAddress, wifKey, producerAddress, amount, fee)
	if err != nil {
		util.PrintCLIError("Failed to burn coins: %s", err)
		return err
	}

	fmt.Println(color.GreenString("Coin burn transaction has been successfully sent!"))
	fmt.Println(color.HiMagentaString("Coin Burn Summary:"))
	fmt.Println("Burner Account:    ", burnerAccount.Address)
	fmt.Println("Producer Address:  ", producerAddress)
	fmt.Println("Amount Burnt (LTC):", amount)
	fmt.Println("Tx. Hash:          ", txHash)

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

// burn executes the burn operation and returns the transaction hash or error
func burn(rpcAddress string, wif *ltcutil.WIF, producerAddress, amount string, fee float64) (string, error) {

	// Split and ensure the RPC address is valid
	host, port, err := net.SplitHostPort(rpcAddress)
	if err != nil {
		return "", fmt.Errorf("invalid RPC address format")
	}

	portInt, _ := strconv.Atoi(port)
	clientOpts := client.Options{
		Host: host,
		Port: portInt,
	}

	cl := client.NewClient(&clientOpts)

	burnerRPCMethod := jsonrpc.MakeFullAPIName(types.NamespaceBurner, "burn")
	res, err := cl.Call(burnerRPCMethod, map[string]interface{}{
		"wif":      wif.String(),
		"producer": producerAddress,
		"amount":   amount,
		"fee":      fee,
	})

	if err != nil {
		return "", fmt.Errorf("failed to execute burn request: %s", err)
	}

	m := res.(map[string]interface{})
	return m["hash"].(string), nil
}
