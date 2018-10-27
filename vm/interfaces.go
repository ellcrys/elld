package vm

import (
	"github.com/ellcrys/elld/blockcode"
	"github.com/ellcrys/elld/types/core"
)

// Block defines an interface for a block
type Block interface {
	GetTransactions() []*core.Transaction
}

// Blockchain interface defines a structure for accessing the blockchain
// and all its primitives.
type Blockchain interface {
	GetBlockCode(address string) (*blockcode.Blockcode, error)
}

// LangBuilder provides information about how to build and run a blockcode of a specific language
type LangBuilder interface {
	GetRunScript() []string
	GetBuildScript() []string
}
