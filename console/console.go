package console

import (
	"fmt"
	"runtime"

	"github.com/ellcrys/druid/console/spell"
	"github.com/ellcrys/druid/util"

	prompt "github.com/c-bata/go-prompt"
)

// Console defines functionalities for create and using
// an interactive Javascript console to perform and query
// the system.
type Console struct {
	prompt     *prompt.Prompt
	executor   *Executor
	suggestMgr *SuggestionManager
}

// New creates a new Console instance
func New(rpcAddr string) (*Console, error) {

	var err error

	c := new(Console)
	c.executor = NewExecutor()
	c.suggestMgr = NewSuggestionManager(initialSuggestions)
	c.executor.setSuggestionUpdateFunc(c.suggestMgr.extend)

	c.executor.spell, err = spell.NewSpell(rpcAddr)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to rpc server. %s", err)
	}

	c.executor.Init()

	exitKeyBind := prompt.KeyBind{
		Key: prompt.ControlC,
		Fn: func(*prompt.Buffer) {
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
	}

	c.prompt = prompt.New(c.executor.OnInput, c.suggestMgr.completer, options...)

	return c, nil
}

// Run the console
func (c *Console) Run() {
	c.about()
	c.prompt.Run()
}

// Exit stops console by killing the process
func (c *Console) Exit() {
	c.executor.exitProgram(true)
}

func (c *Console) about() {
	fmt.Println("Welcome to Druid Javascript console!")
	fmt.Println(fmt.Sprintf("Client:%s, Protocol:%s, Go:%s", util.ClientVersion, util.ProtocolVersion, runtime.Version()))
	fmt.Println(" type '.exit' to exit console")
	fmt.Println("")
}
