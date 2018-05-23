package vm

import (
	"github.com/ellcrys/druid/wire"
)

// Block defines an interface for a block
type Block interface {
	GetTransactions() []*wire.Transaction
}

// Blockchain interface defines a structure for accessing the blockchain
// and all its primitives.
type Blockchain interface {
	GetBlockCode(address string) BlockCode
}

// ExecFunc type destermines the structure of the function expected by the RunScript function
type ExecFunc func() error

// LangBuilder determines the interface of the language builder
type LangBuilder interface {
	GetRunScript() []string
	Build(containerID string) error
}
