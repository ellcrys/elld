package blockchain

import (
	"fmt"

	"github.com/ellcrys/elld/types"

	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/util/logger"
)

// VContexts manages validation contexts
type VContexts struct {
	contexts []types.ValidationContext
}

// Has checks whether a context exists
func (c *VContexts) has(ctx types.ValidationContext) bool {
	for _, c := range c.contexts {
		if c == ctx {
			return true
		}
	}
	return false
}

// clearContexts clears removes all validation contexts
func (c *VContexts) clearContexts() {
	c.contexts = []types.ValidationContext{}
}

// addContext adds a validation context
func (c *VContexts) addContext(contexts ...types.ValidationContext) {
	c.contexts = append(c.contexts, contexts...)
}

func fieldError(field, err string) error {
	var fieldArg = "field:%s, "
	if field == "" {
		fieldArg = "%s"
	}
	return fmt.Errorf(fmt.Sprintf(fieldArg+"error:%s", field, err))
}

func fieldErrorWithIndex(index int, field, err string) error {
	var fieldArg = "field:%s, "
	if field == "" {
		fieldArg = "%s"
	}
	return fmt.Errorf(fmt.Sprintf("index:%d, "+
		fieldArg+"error:%s", index, field, err))
}

// BlockValidator implements a validator for checking
// syntactic, contextual and state correctness of a block
// in relation to various parts of the system.
type BlockValidator struct {
	VContexts

	// block is the block to be validated
	block types.Block

	// txpool refers to the transaction pool
	txpool types.TxPool

	// bChain is the blockchain manager. We use it
	// to query transactions and blocks
	bChain types.Blockchain
}

// NewBlockValidator creates and returns a BlockValidator object
func NewBlockValidator(block types.Block, txPool types.TxPool,
	bChain types.Blockchain, cfg *config.EngineConfig,
	log logger.Logger) *BlockValidator {
	return &BlockValidator{
		block:  block,
		txpool: txPool,
		bChain: bChain,
	}
}

// setContext sets the validation context
func (v *BlockValidator) setContext(ctx types.ValidationContext) {
	if !v.has(ctx) {
		v.contexts = append(v.contexts, ctx)
	}
}
