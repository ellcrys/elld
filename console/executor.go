package console

import (
	"fmt"
	"os"
)

// Executor is responsible for interpreting and executing console inputs
type Executor struct {
}

// NewExecutor creates a new executor
func NewExecutor() *Executor {
	e := new(Executor)
	return e
}

// OnInput receives inputs and executes
func (e *Executor) OnInput(in string) {

	switch in {
	case "exit":
		e.exitProgram()
	}
}

func (e *Executor) exitProgram() {
	fmt.Println("exited")
	os.Exit(0)
}
