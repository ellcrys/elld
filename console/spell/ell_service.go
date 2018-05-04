package spell

import (
	"fmt"
	"net/rpc"

	"github.com/fatih/color"

	"github.com/ellcrys/druid/node"
)

// ELLService provides implementation of actions
// that attempts to alter the state of the blockchain.
type ELLService struct {
	client *rpc.Client
}

// NewELL creates a new ELL service instance
func NewELL(client *rpc.Client) *ELLService {
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

	args := &node.Args{A: 3, B: 4}
	var result node.Result
	err := es.client.Call("Service.Plus", args, &result)
	if err != nil {
		color.Red("%s", err)
		return
	}

	fmt.Println(err, result)
	return
}
