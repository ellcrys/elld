package vm

import (
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
