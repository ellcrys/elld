// Package console provides javascript-enabled console
// environment for interacting with the client. It includes
// pre-loaded methods that access the node's RPC interface
// allowing access to the state and condition of the client.
package console

import (
	"fmt"
	"io/ioutil"
	"path"
	"sync"

	"github.com/ellcrys/mother/util"

	"github.com/fatih/color"

	"github.com/ellcrys/mother/rpc"

	"github.com/ellcrys/mother/accountmgr"
	"github.com/ellcrys/mother/config"
	"github.com/ellcrys/mother/crypto"
	"github.com/ellcrys/mother/util/logger"

	prompt "github.com/c-bata/go-prompt"
)

// Console defines functionalities for create and using
// an interactive Javascript console to perform and query
// the system.
type Console struct {
	sync.RWMutex

	// prompt is the prompt mechanism
	// we are building the console on
	prompt *prompt.Prompt

	// executor is the javascript executor
	executor *Executor

	// suggestMgr managers prompt suggestions
	suggestMgr *SuggestionManager

	// attached indicates whether the console
	// is in attach mode.
	attached bool

	// coinbase is the default account required
	// for signing secure operations
	coinbase *crypto.Key

	// historyFile is the path to the file
	// where the file is stored.
	historyFile string

	// history contains the commands
	// collected during this console session.
	history []string

	// cfg is the client config
	cfg *config.EngineConfig

	confirmedStop bool

	// onStopFunc is called when the
	// console exists. Console caller
	// use this to perform clean up etc
	onStopFunc func()

	// Versions
	protocol string
	client   string
	runtime  string
	commit   string
}

// New creates a new Console instance.
// signatory is the address
func New(coinbase *crypto.Key, historyPath string, cfg *config.EngineConfig, log logger.Logger) *Console {

	c := new(Console)
	c.historyFile = historyPath
	c.executor = newExecutor(coinbase, log)
	c.suggestMgr = newSuggestionManager(initialSuggestions)
	c.coinbase = coinbase
	c.executor.acctMgr = accountmgr.New(path.Join(cfg.DataDir(), config.AccountDirName))
	c.executor.console = c
	c.cfg = cfg

	// retrieve the history
	var history []string
	data, _ := ioutil.ReadFile(historyPath)
	if len(data) > 0 {
		util.BytesToObject(data, &history)
	}

	c.history = append(c.history, history...)

	return c
}

// NewAttached is like New but enables attach mode
func NewAttached(coinbase *crypto.Key, historyPath string, cfg *config.EngineConfig, log logger.Logger) *Console {
	c := New(coinbase, historyPath, cfg, log)
	c.attached = true
	return c
}

// SetRPCServerAddr sets the address of the
// RPC server to be dialled
func (c *Console) SetRPCServerAddr(addr string, secured bool) {
	c.Lock()
	defer c.Unlock()
	c.executor.rpc = &RPCConfig{
		Client: RPCClient{
			Address: makeAddr(addr, secured),
		},
	}
}

// SetRPCServer sets the rpc server
// so we can start and stop it.
// It will panic if this is called on
// a console with attach mode enabled.
func (c *Console) SetRPCServer(rpcServer *rpc.Server, secured bool) {
	c.RLock()
	if c.attached {
		c.RUnlock()
		panic("we don't need a server in attach mode")
	}
	c.RUnlock()
	c.Lock()
	c.executor.rpcServer = rpcServer
	c.Unlock()
	c.SetRPCServerAddr(rpcServer.GetAddr(), secured)
}

// Prepare sets up the console's prompt
// colors, suggestions etc.
func (c *Console) Prepare() error {
	// Set some options
	options := []prompt.Option{
		prompt.OptionPrefixTextColor(prompt.White),
		prompt.OptionAddKeyBind(prompt.KeyBind{
			Key: prompt.ControlC,
			Fn: func(b *prompt.Buffer) {
				c.Stop(false)
			},
		}),
		prompt.OptionDescriptionBGColor(prompt.Black),
		prompt.OptionDescriptionTextColor(prompt.White),
		prompt.OptionSuggestionTextColor(prompt.Turquoise),
		prompt.OptionSuggestionBGColor(prompt.Black),
		prompt.OptionHistory(c.history),
	}

	suggestions, err := c.executor.PrepareContext()
	if err != nil {
		return err
	}

	c.suggestMgr.add(suggestions...)

	// create new prompt and configure it
	// with the options create above
	p := prompt.New(func(in string) {
		c.history = append(c.history, in)
		switch in {

		// handle exit command
		case ".exit":
			c.Stop(true)

		// pass other expressions
		// to the JS executor
		default:
			c.confirmedStop = false
			c.executor.OnInput(in)
		}
	}, c.suggestMgr.completer, options...)

	c.Lock()
	c.prompt = p
	c.Unlock()

	return nil
}

// OnStop sets a function that is called
// when the console is stopped
func (c *Console) OnStop(f func()) {
	c.Lock()
	defer c.Unlock()
	c.onStopFunc = f
}

// Run the console
func (c *Console) Run() {
	c.about()
	c.prompt.Run()
}

// Stop stops console. It saves command
// history and calls the stop callback
func (c *Console) Stop(immediately bool) {
	c.Lock()
	defer c.Unlock()
	if c.confirmedStop || immediately {
		c.saveHistory()
		if c.onStopFunc != nil {
			c.onStopFunc()
		}
	}
	fmt.Println("(To exit, press ^C again or type .exit)")
	c.confirmedStop = true
}

// SetVersions sets the versions of components
func (c *Console) SetVersions(protocol, client, runtime, commit string) {
	c.Lock()
	defer c.Unlock()
	c.protocol = protocol
	c.client = client
	c.runtime = runtime
	c.commit = commit
}

// about prints some information about
// the version of the client and some
// of its components.
func (c *Console) about() {
	c.RLock()
	defer c.RUnlock()
	fmt.Println(color.CyanString("Welcome to Elld Javascript console!"))
	fmt.Println(fmt.Sprintf("Client:%s, Protocol:%s, Commit:%s, Go:%s", c.client, c.protocol, util.String(c.commit).SS(), c.runtime))
	fmt.Println(" type '.exit' to exit console")
	fmt.Println("")
}

// saveHistory stores the commands
// that have been cached in this session.
// Note: Read lock must be called by the caller
func (c *Console) saveHistory() {
	if len(c.history) == 0 {
		return
	}

	bs := util.ObjectToBytes(c.history)
	err := ioutil.WriteFile(c.historyFile, bs, 0644)
	if err != nil {
		panic(err)
	}
}
