package console

import (
	"fmt"
	"path/filepath"

	"github.com/ellcrys/elld/rpc"
	"github.com/gobuffalo/packr"

	"github.com/ellcrys/elld/accountmgr"

	prompt "github.com/c-bata/go-prompt"

	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/rpc/jsonrpc"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/util/logger"
	"github.com/fatih/color"
	"github.com/fatih/structs"
	prettyjson "github.com/ncodes/go-prettyjson"
	"github.com/robertkrimen/otto"
)

// Executor is responsible for executing operations inside a
// javascript VM.
type Executor struct {

	// vm is an Otto instance for JS evaluation
	vm *otto.Otto

	// rpc holds rpc client and config
	rpc *RPCConfig

	// coinbase is the loaded account used
	// for signing blocks and transactions
	coinbase *crypto.Key

	// authToken is the token derived from the last login() invocation
	authToken string

	// log is a logger
	log logger.Logger

	// acctMgr is the account manager
	acctMgr *accountmgr.AccountManager

	// rpcServer is the rpc server to start/connect/stop
	rpcServer *rpc.Server

	// console is the console instance
	console *Console

	// scripts provides access to packed JS scripts
	scripts packr.Box
}

// NewExecutor creates a new executor
func newExecutor(coinbase *crypto.Key, l logger.Logger) *Executor {
	e := new(Executor)
	e.vm = otto.New()
	e.log = l
	e.coinbase = coinbase
	e.scripts = packr.NewBox("./scripts")
	return e
}

func (e *Executor) login(username, password string) interface{} {

	var arg = map[string]interface{}{
		"username": username,
		"password": password,
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

func (e *Executor) callRPCMethod(method string, arg interface{}) (map[string]interface{}, error) {
	rpcResp, err := e.rpc.Client.call(method, arg, e.authToken)
	if err != nil {
		return nil, err
	}

	// decode response object to a map
	s := structs.New(rpcResp)
	s.TagName = "json"
	return s.Map(), nil
}

// PrepareContext adds objects and functions into the VM's global
// contexts allowing users to have access to pre-defined values and objects
func (e *Executor) PrepareContext() ([]prompt.Suggest, error) {

	var suggestions = []prompt.Suggest{}

	// Add some methods to the global namespace
	e.vm.Set("pp", e.pp)
	e.vm.Set("runScript", e.runScript)
	e.vm.Set("rs", e.runScript)

	// nsObj is a namespace for storing
	// rpc methods and other categorized functions
	var nsObj = map[string]map[string]interface{}{
		"admin":    {},
		"personal": {},
		"ell":      {},
		"rpc":      {},
		"_system":  {},
	}

	// Add some methods to namespaces
	nsObj["rpc"]["started"] = e.isRPCServerStarted
	nsObj["rpc"]["start"] = e.startRPCServer
	nsObj["rpc"]["stop"] = e.stopRPCServer
	nsObj["admin"]["login"] = e.login
	nsObj["personal"]["loadAccount"] = e.loadAccount
	nsObj["personal"]["loadedAccount"] = e.loadedAccount
	nsObj["personal"]["createAccount"] = e.createAccount
	nsObj["personal"]["importAccount"] = e.importAccount

	// "private" functions used by system scripts
	nsObj["_system"]["balance"] = NewTxBuilder(e).Balance()

	defer func() {
		for ns, objs := range nsObj {
			e.vm.Set(ns, objs)
		}

		// Add system scripts
		e.runRaw(e.scripts.Bytes("transaction_builder.js"))
	}()

	// Add some methods to the suggestions
	suggestions = append(suggestions, prompt.Suggest{Text: "rpc.start", Description: "Start RPC Server"})
	suggestions = append(suggestions, prompt.Suggest{Text: "rpc.stop", Description: "Stop RPC Server"})
	suggestions = append(suggestions, prompt.Suggest{Text: "rpc.started", Description: "Check whether RPC server has started"})
	suggestions = append(suggestions, prompt.Suggest{Text: "admin.login", Description: "Authenticate the console RPC session"})
	suggestions = append(suggestions, prompt.Suggest{Text: "personal.loadAccount", Description: "Load and set an account as the default"})
	suggestions = append(suggestions, prompt.Suggest{Text: "personal.loadedAccount", Description: "Gets the address of the loaded account"})
	suggestions = append(suggestions, prompt.Suggest{Text: "personal.createAccount", Description: "Create an account"})
	suggestions = append(suggestions, prompt.Suggest{Text: "personal.importAccount", Description: "Import an account"})
	suggestions = append(suggestions, prompt.Suggest{Text: "ell.balance", Description: "Create and send a balance transaction"})

	// If the console is not in attach mode and
	// the rpc server is not started, we cannot
	// set up rpc methods in the namespace and
	// add them as suggestions
	if !e.console.attached && !e.rpcServer.IsStarted() {
		return suggestions, nil
	}

	// Get all the rpc methods information
	resp, err := e.rpc.Client.call("methods", nil, e.authToken)
	if err != nil {
		e.log.Error(color.RedString(RPCClientError(err.Error()).Error()))
		return suggestions, err
	}

	// Create console suggestions and collect methods info
	var methodsInfo = []jsonrpc.MethodInfo{}
	for _, m := range resp.Result.([]interface{}) {
		var mInfo jsonrpc.MethodInfo
		util.MapDecode(m, &mInfo)
		suggestions = append(suggestions, prompt.Suggest{
			Text:        fmt.Sprintf("%s.%s", mInfo.Namespace, mInfo.Name),
			Description: mInfo.Description,
		})
		methodsInfo = append(methodsInfo, mInfo)
	}

	// Add supported methods to the namespace object
	if len(methodsInfo) == 0 {
		return suggestions, nil
	}

	for _, methodInfo := range methodsInfo {
		mName := methodInfo.Name
		ns := methodInfo.Namespace
		if nsObj[ns] == nil {
			nsObj[ns] = map[string]interface{}{}
		}
		nsObj[ns][mName] = func(args ...interface{}) interface{} {

			// parse arguments.
			// App RPC functions can have zero or one argument
			var arg interface{}
			if len(args) > 0 {
				arg = args[0]
			}

			result, err := e.callRPCMethod(mName, arg)
			if err != nil {
				e.log.Error(color.RedString(RPCClientError(err.Error()).Error()))
				v, _ := otto.ToValue(nil)
				return v
			}

			return result
		}
	}

	return suggestions, nil
}

func (e *Executor) runScript(file string) {

	fullPath, err := filepath.Abs(file)
	if err != nil {
		panic(e.vm.MakeCustomError("ExecError", err.Error()))
	}

	script, err := e.vm.Compile(fullPath, nil)
	if err != nil {
		panic(e.vm.MakeCustomError("ExecError", err.Error()))
	}

	_, err = e.vm.Run(script)
	if err != nil {
		panic(e.vm.MakeCustomError("ExecError", err.Error()))
	}
}

func (e *Executor) runRaw(src interface{}) {
	script, err := e.vm.Compile("", src)
	if err != nil {
		panic(e.vm.MakeCustomError("ExecError", err.Error()))
	}

	_, err = e.vm.Run(script)
	if err != nil {
		panic(e.vm.MakeCustomError("ExecError", err.Error()))
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
		panic(e.vm.MakeCustomError("PrettyPrintError", err.Error()))
	}
	fmt.Println(string(bs))
}

// OnInput receives inputs and executes
func (e *Executor) OnInput(in string) {
	switch in {
	case ".help":
		e.help()
	default:
		e.exec(in)
	}
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
	if vExp != nil {
		bs, _ := prettyjson.Marshal(vExp)
		fmt.Println(string(bs))
	}
}

func (e *Executor) help() {
	for _, f := range commonFunc {
		fmt.Println(fmt.Sprintf("%s\t\t%s", f[0], f[1]))
	}
}
