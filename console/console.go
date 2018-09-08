package console

import (
	"fmt"
	"io/ioutil"
	"path"
	"runtime"

	"github.com/ellcrys/elld/accountmgr"
	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/util/logger"
	"github.com/vmihailenco/msgpack"

	prompt "github.com/c-bata/go-prompt"
)

// Console defines functionalities for create and using
// an interactive Javascript console to perform and query
// the system.
type Console struct {

	// prompt is the prompt mechanism
	// we are building the console on
	prompt *prompt.Prompt

	// executor is the javascript executor
	executor *Executor

	// suggestMgr managers prompt suggestions
	suggestMgr *SuggestionManager

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
}

// New creates a new Console instance.
// signatory is the address
func New(coinbase *crypto.Key, historyPath string, cfg *config.EngineConfig, log logger.Logger) *Console {

	c := new(Console)
	c.historyFile = historyPath
	c.executor = newExecutor(coinbase, log)
	c.suggestMgr = newSuggestionManager(initialSuggestions)
	c.coinbase = coinbase
	c.executor.acctMgr = accountmgr.New(path.Join(cfg.ConfigDir(), config.AccountDirName))
	c.cfg = cfg

	// retrieve the history
	var history []string
	data, _ := ioutil.ReadFile(historyPath)
	if len(data) > 0 {
		msgpack.Unmarshal(data, &history)
	}

	c.history = append(c.history, history...)

	return c
}

// ConfigureRPC configures the RPC client
func (c *Console) ConfigureRPC(rpcAddress string, secured bool) {
	c.executor.rpc = &RPCConfig{
		Client:  RPCClient(rpcAddress),
		Secured: secured,
	}

	// reinitialize the rpc client,
	// this time compute while taking
	// the secured field into account
	c.executor.rpc.Client = RPCClient(c.executor.rpc.GetAddr())
}

// Prepare sets up the console
func (c *Console) Prepare() error {

	// Set some options
	options := []prompt.Option{
		prompt.OptionPrefixTextColor(prompt.White),
		prompt.OptionAddKeyBind(prompt.KeyBind{
			Key: prompt.ControlC,
			Fn: func(*prompt.Buffer) {
				c.saveHistory()
				c.executor.exitProgram(false)
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

	c.suggestMgr.suggestions = append(c.suggestMgr.suggestions, suggestions...)

	// create new prompt and configure it
	// with the options create above
	c.prompt = prompt.New(func(in string) {
		c.history = append(c.history, in)
		c.executor.OnInput(in)
	}, c.suggestMgr.completer, options...)

	return nil
}

// Run the console
func (c *Console) Run() {
	c.about()
	c.prompt.Run()
}

// Exit stops console by killing the process
func (c *Console) Exit() {
	c.saveHistory()
	c.executor.exitProgram(true)
}

// about prints some information about
// the version of the client and some
// of its components.
func (c *Console) about() {
	fmt.Println("Welcome to Elld Javascript console!")
	fmt.Println(fmt.Sprintf("Client:%s, Protocol:%s, Go:%s", config.ClientVersion, config.ProtocolVersion, runtime.Version()))
	fmt.Println(" type '.exit' to exit console")
	fmt.Println("")
}

// saveHistory stores the console history collected so far
func (c *Console) saveHistory() {
	if len(c.history) == 0 {
		return
	}

	bs, _ := msgpack.Marshal(c.history)
	err := ioutil.WriteFile(c.historyFile, bs, 0644)
	if err != nil {
		panic(err)
	}
}
