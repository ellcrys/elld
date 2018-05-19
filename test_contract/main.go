package main

import (
	"fmt"

	"github.com/ellcrys/stub.go"
)

//TestContract ..
type TestContract struct{}

//OnInit ..
func (contract *TestContract) OnInit(ctx *ellstub.Context) {

}

//OnTerminate ..
func (contract *TestContract) OnTerminate(ctx *ellstub.Context) {

}

//DoSomething ..
func (contract *TestContract) DoSomething(ctx *ellstub.Context) (interface{}, error) {
	fmt.Printf("%v\n", "hello world")
	fmt.Printf("%v\n", ctx.Data)
	return ctx.Data, nil
}

func main() {
	ellstub.Run(new(TestContract), 4000)

	// type MyData struct {
	// 	Amount float64
	// }
	// var myargs *ellstub.InvokeData
	// myargs = &ellstub.InvokeData{
	// 	Function:   "DoSomething",
	// 	ContractID: "983890276903",
	// 	Data: &MyData{
	// 		Amount: 10.50,
	// 	},
	// }

	// conn, _ := net.Dial("tcp", "127.0.0.1:62987")

	// Client := rpc2.NewClientWithCodec(jsonrpc.NewJSONCodec(conn))
	// go Client.Run()
	// _ = Client.Call("invoke", &myargs, nil)

}
