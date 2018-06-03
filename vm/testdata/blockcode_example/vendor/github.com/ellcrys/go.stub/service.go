package stub

import "fmt"

// Ctx holds information about a function invocation.
// It is updated on each function invocation
var Ctx *Context

// BlockInfo represents information about the current block
type BlockInfo struct {
	BlockNumber uint64
}

// Tx represents details of an incoming transaction
type Tx struct {
	ID    string `json:"txId"`
	Value string `json:"value"`
}

// Context represents the context of a transaction
type Context struct {
	Tx        *Tx
	BlockInfo *BlockInfo
}

// Args represents the parameters of a function call
type Args struct {
	Func      string            `json:"func"`
	Payload   map[string]string `json:"payload"`
	Tx        *Tx               `json:"tx"`
	BlockInfo *BlockInfo        `json:"blockInfo"`
}

// Result represents the output of a function call
type Result struct {
	Error bool        `json:"error"`
	Body  interface{} `json:"body"`
}

// Service describes the RPC functions that enables
// interactions between the blockcode and external callers (e.g vm)
type Service struct {
	stub *stub
}

func newService(stub *stub) *Service {
	s := new(Service)
	s.stub = stub
	return s
}

// Invoke invokes a blockcode function
// - Call OnInit on the blockcode for initialization to take place
// - Get the function to be invoked.
// - Call the function
func (s *Service) Invoke(args Args) *Result {

	Ctx = &Context{
		Tx:        args.Tx,
		BlockInfo: args.BlockInfo,
	}

	s.stub.blockcode.OnInit()

	bFunc := getFunc(args.Func)
	if bFunc == nil {
		return &Result{
			Error: true,
			Body:  fmt.Sprintf("unknown function `%s`", args.Func),
		}
	}

	var result = &Result{}
	out, err := bFunc()
	if err != nil {
		result.Error = true
		result.Body = err.Error()
		return result
	}

	result.Body = out

	return result
}
