package console

import (
	"fmt"

	"github.com/ellcrys/druid/util"

	prompt "github.com/c-bata/go-prompt"
)

// Console defines functionalities for create and using
// an interactive Javascript console to perform and query
// the system.
type Console struct {
	prompt   *prompt.Prompt
	executor *Executor
}

// New creates a new Console instance
func New() *Console {
	c := new(Console)
	c.executor = NewExecutor()

	exitKeyBind := prompt.KeyBind{
		Key: prompt.ControlC,
		Fn: func(*prompt.Buffer) {
			c.executor.exitProgram(false)
		},
	}

	options := []prompt.Option{
		prompt.OptionPrefixTextColor(prompt.White),
		prompt.OptionAddKeyBind(exitKeyBind),
	}

	c.prompt = prompt.New(c.executor.OnInput, completer, options...)

	return c
}

// Run the console
func (c *Console) Run() {
	c.about()
	c.prompt.Run()
}

func (c *Console) about() {
	fmt.Println("Welcome to Druid Javascript console!")
	fmt.Println(fmt.Sprintf("Client:%s, Protocol:%s", util.ClientVersion, util.ProtocolVersion))
	fmt.Println(" type '.exit' to exit console")
	fmt.Println("")
}
