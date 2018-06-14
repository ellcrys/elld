package console

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/rpc"
	"runtime"

	"github.com/ellcrys/elld/console/spell"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/util"

	prompt "github.com/c-bata/go-prompt"
)

// Console defines functionalities for create and using
// an interactive Javascript console to perform and query
// the system.
type Console struct {
	prompt      *prompt.Prompt
	executor    *Executor
	suggestMgr  *SuggestionManager
	rpcClient   *rpc.Client
	signatory   *crypto.Key
	historyFile string
	history     []string
}

// New creates a new Console instance.
// signatory is the address
func New(signatory *crypto.Key, historyFilePath string) *Console {

	c := new(Console)
	c.historyFile = historyFilePath
	c.executor = NewExecutor()
	c.suggestMgr = NewSuggestionManager(initialSuggestions)
	c.executor.spell = spell.NewSpell(signatory)

	var history []string
	histBs, _ := ioutil.ReadFile(historyFilePath)
	if len(histBs) > 0 {
		json.Unmarshal(histBs, &history)
	}

	histOpt := prompt.OptionHistory(history)

	exitKeyBind := prompt.KeyBind{
		Key: prompt.ControlC,
		Fn: func(*prompt.Buffer) {
			c.saveHistory()
			c.executor.exitProgram(false)
		},
	}

	options := []prompt.Option{
		prompt.OptionPrefixTextColor(prompt.White),
		prompt.OptionAddKeyBind(exitKeyBind),
		prompt.OptionDescriptionBGColor(prompt.Black),
		prompt.OptionDescriptionTextColor(prompt.White),
		prompt.OptionSuggestionTextColor(prompt.Turquoise),
		prompt.OptionSuggestionBGColor(prompt.Black),
		histOpt,
	}

	c.prompt = prompt.New(func(in string) {
		c.history = append(c.history, in)
		c.executor.OnInput(in)
	}, c.suggestMgr.completer, options...)

	return c
}

// DialRPCServer dials the RPC server
func (c *Console) DialRPCServer(rpcAddr string) error {
	var err error
	c.rpcClient, err = rpc.DialHTTP("tcp", rpcAddr)
	if err != nil {
		return err
	}
	c.executor.spell.SetClient(c.rpcClient)
	return nil
}

// PrepareVM sets up the VM executors context
func (c *Console) PrepareVM() error {
	return c.executor.PrepareContext()
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

func (c *Console) about() {
	fmt.Println("Welcome to Druid Javascript console!")
	fmt.Println(fmt.Sprintf("Client:%s, Protocol:%s, Go:%s", util.ClientVersion, util.ProtocolVersion, runtime.Version()))
	fmt.Println(" type '.exit' to exit console")
	fmt.Println("")
}

func (c *Console) saveHistory() {
	if len(c.history) == 0 {
		return
	}
	bs, _ := json.Marshal(c.history)
	err := ioutil.WriteFile(c.historyFile, bs, 0644)
	if err != nil {
		panic(err)
	}
}
