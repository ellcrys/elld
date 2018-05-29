package main

import (
	"github.com/ellcrys/stub.go"
)

type InvokeData struct {
	Function   string      `json:"Function"`
	ContractID string      `json:"ContractID"`
	Data       interface{} `json:"Data"`
}

//MyContract
type MyContract struct{}

//OnInit
func (contract *MyContract) OnInit(ctx *ellstub.Context) {

}

//OnTerminate
func (contract *MyContract) OnTerminate(ctx *ellstub.Context) {

}

//DoSomething
func (contract *MyContract) DoSomething(ctx *ellstub.Context) (interface{}, error) {
	return ctx, nil
}

func main() {

	ellstub.Run(new(MyContract), 4000)

}
