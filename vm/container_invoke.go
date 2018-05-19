package vm

import (
	"fmt"

	"github.com/cenkalti/hub"
)

//Transaction ..
type Transaction struct {
	Function string      `json:"Function"`
	Data     interface{} `json:"Data"`
}

//Payload data to be sent to container for execution
type InvokeData struct {
	ContractID string      `json:"ContractID"`
	Function   string      `json:"Function"`
	Data       interface{} `json:"Data"`
}

//InvokeResponseData is the structure of data expected from Invoke request
type InvokeResponseData struct {
	Status string `json:"status"`
	Code   int    `json:"code"`
	Data   []byte `json:"data"`
}

const onResponse hub.Kind = iota

//func (InvokeResponseData) Kind() hub.Kind { return onResponse }

//Invoke a contract
func (container *Container) Invoke(block *Transaction) error {
	go container.service.Run()
	payload := &InvokeData{
		Function:   block.Function,
		ContractID: container.contractID,
		Data:       block.Data,
	}

	err := container.service.Call("invoke", payload, struct{}{})
	if err != nil {
		return err
	}
	return nil
}

//OnResponse event from invoke
func (container *Container) OnResponse(callback func(interface{})) {
	container.eventHub.Subscribe(onResponse, func(e hub.Event) {
		fmt.Printf("%v\n", e)
		go callback(e.(interface{}))
	})
}
