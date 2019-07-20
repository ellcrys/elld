package blockchain

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/ellcrys/elld/params"
	"github.com/ellcrys/elld/types"

	"github.com/shopspring/decimal"

	"github.com/ellcrys/elld/blockchain/common"
	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/util"
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

// validateHeader checks that an header fields and
// value format or type is valid.
func (v *BlockValidator) validateHeader(h types.Header) (errs []error) {

	// For non-genesis block, parent hash must be set
	if h.GetNumber() != 1 && h.GetParentHash().IsEmpty() {
		errs = append(errs, fieldError("parentHash", "parent hash is required"))
	} else if h.GetNumber() == 1 && !h.GetParentHash().IsEmpty() {
		// For genesis block, parent hash is not required
		errs = append(errs, fieldError("parentHash",
			"parent hash is not expected in a genesis block"))
	}

	// Number cannot be 0 or less
	if h.GetNumber() < 1 {
		errs = append(errs, fieldError("number",
			"number must be greater or equal to 1"))
	}

	// Creator's public key must be provided
	// and must be decodeable
	if len(h.GetCreatorPubKey()) == 0 {
		errs = append(errs, fieldError("creatorPubKey",
			"creator's public key is required"))
	} else if _, err := crypto.
		PubKeyFromBase58(h.GetCreatorPubKey().String()); err != nil {
		errs = append(errs, fieldError("creatorPubKey", err.Error()))
	}

	// Transactions root must be valid
	if !h.GetTransactionsRoot().Equal(common.ComputeTxsRoot(v.block.GetTransactions())) {
		errs = append(errs, fieldError("transactionsRoot",
			"transactions root is not valid"))
	}

	// State root must be provided
	if h.GetStateRoot() == util.EmptyHash {
		errs = append(errs, fieldError("stateRoot", "state root is required"))
	}

	// Difficulty must be a numeric value
	// and greater than zero
	if !h.GetParentHash().IsEmpty() {
		if h.GetDifficulty() == nil || h.GetDifficulty().Cmp(util.Big0) == 0 {
			errs = append(errs, fieldError("difficulty",
				"difficulty must be greater than zero"))
		}
	}

	// Timestamp is required and must not be more than
	// 15 seconds in the future
	if h.GetTimestamp() == 0 {
		errs = append(errs, fieldError("timestamp", "timestamp is required"))
	} else if time.Unix(h.GetTimestamp(), 0).After(time.Now().
		Add(15 * time.Second).UTC()) {
		errs = append(errs, fieldError("timestamp",
			"timestamp is too far in the future"))
	}

	return
}

// CheckPoW checks the PoW and difficulty values in the header.
// If chain is set, the parent chain is search within the provided
// chain, otherwise, the best chain is searched
func (v *BlockValidator) CheckPoW(opts ...types.CallOp) (errs []error) {
	return
}

// CheckSize checks the size of the blocks
func (v *BlockValidator) CheckSize() (errs []error) {

	maxBlockSize := params.MaxBlockTxsSize + params.MaxBlockNonTxsSize
	if v.block.GetSize() > maxBlockSize {
		errs = append(errs, fmt.Errorf("block size exceeded"))
	}

	return
}

// CheckFields checks the field and their values.
func (v *BlockValidator) CheckFields() (errs []error) {

	// Block must not be nil
	if v.block == nil {
		errs = append(errs, fmt.Errorf("nil block"))
		return
	}

	// Header is required
	header := v.block.GetHeader().(*core.Header)
	if header == nil {
		errs = append(errs, fieldError("header", "header is required"))
		return
	}
	for _, err := range v.validateHeader(v.block.GetHeader()) {
		errs = append(errs, fmt.Errorf(strings.Replace(err.Error(),
			"field:", "field:header.", -1)))
	}
	if len(errs) > 0 {
		return
	}

	// Hash must be provided
	if v.block.GetHash() == util.EmptyHash {
		errs = append(errs, fieldError("hash", "hash is required"))
	} else if v.block.GetHeader() != nil && !v.block.GetHash().
		Equal(v.block.ComputeHash()) {
		errs = append(errs, fieldError("hash", "hash is not correct"))
	}

	// Check the size of the block
	errs = append(errs, v.CheckSize()...)

	// Signature must be provided
	if len(v.block.GetSignature()) == 0 {
		errs = append(errs, fieldError("sig", "signature is required"))
	}

	// Check that the signature is valid
	if sigErrs := v.checkSignature(); len(sigErrs) > 0 {
		errs = append(errs, sigErrs...)
	}

	return
}

// CheckAllocs verifies allocation transactions
// such as transaction fees, mining rewards etc.
func (v *BlockValidator) CheckAllocs() (errs []error) {

	// No need performing allocation checks
	// for the genesis block
	if v.block.GetNumber() == 1 {
		return
	}

	var blockAllocs = [][]interface{}{}
	var expectedAllocs = [][]interface{}{}
	var transactions = v.block.GetTransactions()
	var totalFees = decimal.New(0, 0)

	// collect all the allocations transactions
	// and in doing so, calculate the total fees
	// for non-allocation transactions
	for _, tx := range transactions {
		if tx.GetType() == core.TxTypeAlloc {
			blockAllocs = append(blockAllocs, []interface{}{
				tx.GetFrom(),
				tx.GetTo(),
				tx.GetValue().Decimal().StringFixed(params.Decimals),
			})
			continue
		}
		totalFees = totalFees.Add(tx.GetFee().Decimal())
	}

	// Compare the allocations in the block
	// with the computed expected allocations.
	// If they don't match, then add an error
	if len(transactions) > 0 && !reflect.DeepEqual(blockAllocs, expectedAllocs) {
		errs = append(errs, fieldError("transactions",
			"block allocations and expected allocations do not match"))
		return
	}

	return
}

// checkSignature checks whether the signature is valid.
// Expects the block to have a valid creators public key and
// signature to be set.
func (v *BlockValidator) checkSignature() (errs []error) {

	pubKey, err := crypto.PubKeyFromBase58(v.block.GetHeader().
		GetCreatorPubKey().String())
	if err != nil {
		errs = append(errs, fieldError("header.creatorPubKey",
			err.Error()))
		return
	}

	valid, err := pubKey.Verify(v.block.GetBytesNoHashSig(),
		v.block.GetSignature())
	if err != nil {
		errs = append(errs, fieldError("sig", err.Error()))
	} else if !valid {
		errs = append(errs, fieldError("sig",
			"signature is not valid"))
	}

	return
}

// CheckTransactions validates all transactions in the
// block in relation to the block's destined chain.
func (v *BlockValidator) CheckTransactions(opts ...types.CallOp) (errs []error) {
	txValidator := NewTxsValidator(v.block.GetTransactions(), v.txpool, v.bChain)
	txValidator.addContext(v.contexts...)
	for _, err := range txValidator.Validate(opts...) {
		errs = append(errs, fmt.Errorf(strings.Replace(err.Error(), "index:", "tx:", -1)))
	}
	return
}
