package console

import (
	"fmt"
	"os"

	"github.com/c-bata/go-prompt"
	"github.com/ellcrys/druid/console/spell"
	"github.com/fatih/color"
	"github.com/robertkrimen/otto"
)

// Executor is responsible for interpreting and executing console inputs
type Executor struct {
	vm                   *otto.Otto
	suggestionUpdateFunc func([]prompt.Suggest)
	exit                 bool
	spell                *spell.Spell
}

// NewExecutor creates a new executor
func NewExecutor() *Executor {
	e := new(Executor)
	e.vm = otto.New()
	return e
}

// Init adds objects and functions into the VM's contexts
func (e *Executor) Init() {

	var EllSpellObj = map[string]interface{}{
		"send": e.spell.EllService.Send,
	}

	e.vm.Set("spell", map[string]interface{}{
		"ell": EllSpellObj,
	})
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

func (e *Executor) extendSuggestionsFromVM() {

	var symbolSuggestions []prompt.Suggest
	for name, v := range e.vm.Context().Symbols {
		symbolSuggestions = append(symbolSuggestions, prompt.Suggest{Text: name, Description: getType(v)})
	}

	if e.suggestionUpdateFunc != nil {
		e.suggestionUpdateFunc(symbolSuggestions)
	}
}

func (e *Executor) exec(in string) {

	v, err := e.vm.Run(in)
	if err != nil {
		color.Red("%s", err.Error())
		return
	}

	go e.extendSuggestionsFromVM()

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

func (e *Executor) setSuggestionUpdateFunc(f func([]prompt.Suggest)) {
	e.suggestionUpdateFunc = f
}

func getType(v otto.Value) string {

	if v.IsBoolean() {
		return "Boolean"
	}

	if v.IsFunction() {
		return "Function"
	}

	if v.IsNumber() {
		return "Number"
	}

	if v.IsObject() {
		return "Object"
	}

	if v.IsString() {
		return "String"
	}

	return ""
}
