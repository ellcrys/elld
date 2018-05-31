package vm

import (
	"github.com/ellcrys/druid/blockcode"
	"github.com/ellcrys/druid/wire"
)

// Block defines an interface for a block
type Block interface {
	GetTransactions() []*wire.Transaction
}

// Blockchain interface defines a structure for accessing the blockchain
// and all its primitives.
type Blockchain interface {
	GetBlockCode(address string) *blockcode.Blockcode
}

// LangBuilder provides information about how to build and run a blockcode of a specific language
type LangBuilder interface {
	GetRunScript() []string
	GetBuildScript() []string
}
