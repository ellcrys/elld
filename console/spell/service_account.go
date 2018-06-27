package spell

import (
	"fmt"

	net_rpc "net/rpc"

	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/rpc"
)

// Account provides account access and management functions
type Account struct {
	client *net_rpc.Client
	key    *crypto.Key
}

// NewAccount creates a new instance of AccountService
func NewAccount(client *net_rpc.Client, key *crypto.Key) *Account {
	es := new(Account)
	es.client = client
	es.key = key
	return es
}

// GetAccounts fetches all the accounts on the node
func (es *Account) GetAccounts() []string {

	if es.client == nil {
		Panic("Accounts.GetAccounts", "client not initialized")
	}

	var args = new(rpc.AccountGetAllPayload)
	var result rpc.Result
	err := es.client.Call("Service.AccountGetAll", args, &result)
	if err != nil {
		Panic("Accounts.GetAccounts", err.Error())
	}

	if result.Error != "" {
		Panic("Accounts.GetAccounts", fmt.Sprintf("%s (code: %d)", result.Error, result.ErrCode))
	}

	return result.Data["accounts"].([]string)
}
