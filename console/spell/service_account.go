package spell

import (
	"fmt"

	net_rpc "net/rpc"

	"github.com/ellcrys/druid/rpc"
	"github.com/fatih/color"
)

// AccountService provides account access and management functions
type AccountService struct {
	client *net_rpc.Client
}

// NewAccountService creates a new instance of AccountService
func NewAccountService(client *net_rpc.Client) *AccountService {
	es := new(AccountService)
	es.client = client
	return es
}

// GetAccounts fetches all the accounts on the node
func (es *AccountService) GetAccounts() []string {

	if es.client == nil {
		color.Red("rpc: rpc mode not enabled")
		panic(fmt.Errorf("client not initialized"))
	}

	var args = new(rpc.AccountsGetPayload)
	var result rpc.Result
	err := es.client.Call("Service.GetAccounts", args, &result)
	if err != nil {
		color.Red("%s", err)
		panic(err)
	}

	if result.Error != "" {
		panic(fmt.Errorf(fmt.Sprintf("%s (code: %d)", result.Error, result.ErrCode)))
	}

	return result.Data["accounts"].([]string)
}
