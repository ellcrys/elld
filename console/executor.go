package console

import (
	"fmt"
	"os"

	"github.com/robertkrimen/otto"
)

// Executor is responsible for interpreting and executing console inputs
type Executor struct {
	vm   *otto.Otto
	exit bool
}

// NewExecutor creates a new executor
func NewExecutor() *Executor {
	e := new(Executor)
	e.vm = otto.New()
	return e
}

// OnInput receives inputs and executes
func (e *Executor) OnInput(in string) {

	switch in {
	case ".exit":
		e.exitProgram(true)
	case ".help":
		e.help()
	default:
		e.execJs(in)
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

func (e *Executor) execJs(in string) {
	v, err := e.vm.Run(in)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(v)
}

func (e *Executor) help() {
	for _, f := range commonFunc {
		fmt.Println(fmt.Sprintf("%s\t\t%s", f[0], f[1]))
	}
}
