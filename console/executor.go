package console

import (
	"fmt"
	"os"

	"github.com/ellcrys/druid/console/spell"
	"github.com/fatih/color"
	"github.com/robertkrimen/otto"
)

func recoverFunc(rpcMethod string) {
	if r := recover(); r != nil {
		color.Red("%s: %s", rpcMethod, r)
	}
}

// Executor is responsible for executing operations inside a
// javascript VM.
type Executor struct {
	vm    *otto.Otto
	exit  bool
	spell *spell.Spell
}

// NewExecutor creates a new executor
func NewExecutor() *Executor {
	e := new(Executor)
	e.vm = otto.New()
	return e
}

// PrepareContext adds objects and functions into the VM's global
// contexts allowing users to have access to pre-defined values and objects
func (e *Executor) PrepareContext() error {

	var spell = map[string]interface{}{
		"ell": map[string]interface{}{
			"send": e.spell.Ell.Send,
		},
		"account": map[string]interface{}{
			"getAccounts": e.spell.Account.GetAccounts,
		},
	}

	go func() {
		defer recoverFunc("spell.GetAccounts")
		spell["accounts"] = e.spell.Account.GetAccounts()
	}()

	e.vm.Set("spell", spell)

	return nil
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

	v, err := e.vm.Run(in)
	if err != nil {
		color.Red("%s", err.Error())
		return
	}

	if v.IsNull() || v.IsUndefined() {
		color.Magenta("%s", v)
		return
	}

	v, err = e.vm.Call("JSON.stringify", nil, v, nil, 2)
	if err != nil {
		color.Red("%s", err.Error())
		return
	}

	fmt.Println(v)
}

func (e *Executor) help() {
	for _, f := range commonFunc {
		fmt.Println(fmt.Sprintf("%s\t\t%s", f[0], f[1]))
	}
}
