package console

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/util/logger"
	"github.com/fatih/color"
	"github.com/fatih/structs"
	prettyjson "github.com/ncodes/go-prettyjson"
	"github.com/robertkrimen/otto"
)

// FuncCallError creates an error describing
// an issue with the way a function was called.
func FuncCallError(msg string) error {
	return fmt.Errorf("function call error: %s", msg)
}

// Executor is responsible for executing operations inside a
// javascript VM.
type Executor struct {

	// vm is an Otto instance for JS evaluation
	vm *otto.Otto

	// exit indicates a request to exit the executor
	exit bool

	// rpc holds rpc client and config
	rpc *RPCConfig

	// coinbase is the loaded account used
	// for signing blocks and transactions
	coinbase *crypto.Key

	// authToken is the token derived from the last login() invocation
	authToken string

	// log is a logger
	log logger.Logger
}

// NewExecutor creates a new executor
func newExecutor(coinbase *crypto.Key, l logger.Logger) *Executor {
	e := new(Executor)
	e.vm = otto.New()
	e.log = l
	return e
}

func (e *Executor) login(args ...interface{}) interface{} {

	// parse arguments.
	// App RPC functions can have zero or one argument
	var arg interface{}
	if len(args) > 0 {
		arg = args[0]
	}

	// Call the auth RPC method
	rpcResp, err := e.rpc.Client.call("auth", arg, "")
	if err != nil {
		e.log.Error(color.RedString(RPCClientError(err.Error()).Error()))
		v, _ := otto.ToValue(nil)
		return v
	}

	if !rpcResp.IsError() {
		e.authToken = rpcResp.Result.(string)
		return nil
	}

	// decode response object to a map
	s := structs.New(rpcResp)
	s.TagName = "json"
	return s.Map()
}

// PrepareContext adds objects and functions into the VM's global
// contexts allowing users to have access to pre-defined values and objects
func (e *Executor) PrepareContext() error {

	e.vm.Set("pp", e.pp)
	e.vm.Set("runScript", e.runScript)
	e.vm.Set("rs", e.runScript)

	// Get all the methods
	resp, err := e.rpc.Client.call("methods", nil, e.authToken)
	if err != nil {
		e.log.Error(color.RedString(RPCClientError(err.Error()).Error()))
	}

	// Define global object
	var globalObj = map[string]interface{}{}

	// Add supported methods to the global
	// objects map
	if resp != nil {

		// set methods as a global variable for quick
		e.vm.Set("methods", resp.Result)
		e.vm.Set("login", e.login)

		for _, methodName := range resp.Result.([]interface{}) {
			var mName = methodName.(string)
			globalObj[mName] = func(args ...interface{}) interface{} {

				// parse arguments.
				// App RPC functions can have zero or one argument
				var arg interface{}
				if len(args) > 0 {
					arg = args[0]
				}

				// Call the RPC method passing the RPC API params
				rpcResp, err := e.rpc.Client.call(mName, arg, e.authToken)
				if err != nil {
					e.log.Error(color.RedString(RPCClientError(err.Error()).Error()))
					v, _ := otto.ToValue(nil)
					return v
				}

				// decode response object to a map
				s := structs.New(rpcResp)
				s.TagName = "json"
				return s.Map()
			}
		}
	}

	e.vm.Set("ell", globalObj)

	return nil
}

func (e *Executor) runScript(file string) {

	fullPath, err := filepath.Abs(file)
	if err != nil {
		panic(err)
	}

	script, err := ioutil.ReadFile(fullPath)
	if err != nil {
		panic(err)
	}

	_, err = e.vm.Run(string(script))
	if err != nil {
		panic(err)
	}
}

// pp pretty prints a slice of arbitrary objects
func (e *Executor) pp(values ...interface{}) {
	var v interface{} = values
	if len(values) == 1 {
		v = values[0]
	}
	bs, err := prettyjson.Marshal(v)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(bs))
}

// OnInput receives inputs and executes
func (e *Executor) OnInput(in string) {

	e.exit = false

	switch in {
	case ".exit":
		e.exitProgram(true)
	case ".help":
		e.help()
	default:

		e.exec(in)
	}
}

func (e *Executor) exitProgram(immediately bool) {
	if !immediately && !e.exit {
		fmt.Println("(To exit, press ^C again or type .exit)")
		e.exit = true
		return
	}
	os.Exit(0)
}

func (e *Executor) exec(in string) {

	// RecoverFunc recovers from panics.
	defer func() {
		if r := recover(); r != nil {
			color.Red("Panic: %s", r)
		}
	}()

	v, err := e.vm.Run(in)
	if err != nil {
		color.Red("%s", err.Error())
		return
	}

	if v.IsNull() || v.IsUndefined() {
		color.Magenta("%s", v)
		return
	}

	vExp, _ := v.Export()
	bs, err := prettyjson.Marshal(vExp)
	fmt.Println(string(bs))
}

func (e *Executor) help() {
	for _, f := range commonFunc {
		fmt.Println(fmt.Sprintf("%s\t\t%s", f[0], f[1]))
	}
}
