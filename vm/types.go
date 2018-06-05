package vm

import "github.com/ellcrys/elld/wire"

// BlockInfo represents information about the current block
type BlockInfo struct {
	BlockNumber uint64
}

// Tx represents details of an incoming transaction
type Tx struct {
	ID    string `json:"txId"`
	Value string `json:"value"`
}

// Args defines the structure of arguments sent to a blockcode
type Args struct {
	Func      string            `json:"func"`
	Payload   map[string]string `json:"payload"`
	Tx        *Tx               `json:"tx"`
	BlockInfo *BlockInfo        `json:"blockInfo"`
}

// BlockcodeMsg defines the message expected from a blockcode
type BlockcodeMsg struct {
	Type int               `json:"type"`
	Tx   *wire.Transaction `json:"tx"`
}

// Result represents the output of a function call
type Result struct {
	Error bool        `json:"error"`
	Body  interface{} `json:"body"`
}
