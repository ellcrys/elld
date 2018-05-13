package spell

import (
	"fmt"

	net_rpc "net/rpc"

	"github.com/ellcrys/druid/rpc"
	"github.com/fatih/color"
)

// ELLService provides implementation of actions
// that attempts to alter the state of the blockchain.
type ELLService struct {
	client *net_rpc.Client
}

// NewELL creates a new ELL service instance
func NewELL(client *net_rpc.Client) *ELLService {
	es := new(ELLService)
	es.client = client
	return es
}

// Send sends ELL from one account to another
func (es *ELLService) Send() {

	if es.client == nil {
		color.Red("rpc: rpc mode not enabled")
		return
	}

	args := &rpc.Args{A: 3, B: 4}
	var result rpc.Result
	err := es.client.Call("Service.Plus", args, &result)
	if err != nil {
		color.Red("%s", err)
		return
	}

	fmt.Println(err, result)
	return
}
