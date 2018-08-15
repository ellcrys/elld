package blockchain

import (
	"fmt"
	"strings"
	"time"

	"github.com/ellcrys/elld/config"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/miner/blakimoto"
	"github.com/ellcrys/elld/txpool"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/util/logger"
	"github.com/ellcrys/elld/wire"
)

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
	return fmt.Errorf(fmt.Sprintf("index:%d, "+fieldArg+"error:%s", index, field, err))
}

// BlockValidator implements a validator for checking
// syntactic, contextual and state correctness of a block
// in relation to various parts of the system.
type BlockValidator struct {

	// block is the block to be validated
	block core.Block

	// txpool refers to the transaction pool
	txpool *txpool.TxPool

	// bchain is the blockchain manager. We use it
	// to query transactions and blocks
	bchain core.Blockchain

	// allowDuplicateCheck enables duplication checks on other
	// collections. If set to true, a transaction existing in
	// a collection such as the transaction pool, chains etc
	// will be considered invalid.
	allowDuplicateCheck bool

	// blakimoto is an instance of PoW implementation
	blakimoto *blakimoto.Blakimoto

	// verSeal seal instructs the validator whether or not
	// to verify the difficult and PoW fields of a given block
	verSeal bool
}

// NewBlockValidator creates and returns a BlockValidator object
func NewBlockValidator(block core.Block, txPool *txpool.TxPool,
	bchain core.Blockchain, allowDupCheck bool, cfg *config.EngineConfig, log logger.Logger) *BlockValidator {
	return &BlockValidator{
		block:               block,
		txpool:              txPool,
		bchain:              bchain,
		allowDuplicateCheck: allowDupCheck,
		blakimoto:           blakimoto.ConfiguredBlakimoto(blakimoto.ModeNormal, log),
	}
}

func (v *BlockValidator) verifySeal() {
	v.verSeal = true
}

// Validate runs a series of checks against the loaded block
// returning all errors found.
func (v *BlockValidator) Validate() (errs []error) {
	errs = v.check()
	errs = append(errs, v.checkSignature()...)
	if v.allowDuplicateCheck {
		errs = append(errs, v.duplicateCheck(v.block)...)
	}
	return errs
}

// validateHeader checks that an header fields and
// value format or type is valid.
func (v *BlockValidator) validateHeader(h core.Header) (errs []error) {

	// For non-genesis block, parent hash must be set
	if h.GetNumber() != 1 && h.GetParentHash().IsEmpty() {
		errs = append(errs, fieldError("parentHash", "parent hash is required"))
	} else if h.GetNumber() == 1 && !h.GetParentHash().IsEmpty() {
		// For genesis block, parent hash is not required
		errs = append(errs, fieldError("parentHash", "parent hash is not expected in a genesis block"))
	}

	// Number cannot be 0 or less
	if h.GetNumber() < 1 {
		errs = append(errs, fieldError("number", "number must be greater or equal to 1"))
	}

	// Creator's public key must be provided
	// and must be decodeable
	if len(h.GetCreatorPubKey()) == 0 {
		errs = append(errs, fieldError("creatorPubKey", "creator's public key is required"))
	} else if _, err := crypto.PubKeyFromBase58(h.GetCreatorPubKey().String()); err != nil {
		errs = append(errs, fieldError("creatorPubKey", err.Error()))
	}

	// Transactions root must be provided
	if h.GetTransactionsRoot() == util.EmptyHash {
		errs = append(errs, fieldError("transactionsRoot", "transaction root is required"))
	}

	// State root must be provided
	if h.GetStateRoot() == util.EmptyHash {
		errs = append(errs, fieldError("stateRoot", "state root is required"))
	}

	// Difficulty must be a numeric value
	// and greater than zero
	if h.GetDifficulty() == nil || h.GetDifficulty().Cmp(util.Big0) == 0 {
		errs = append(errs, fieldError("difficulty", "difficulty must be non-zero and non-negative"))
	}

	// Timestamp must not be zero or greater than
	// 2 hours in the future
	if h.GetTimestamp() <= 0 {
		errs = append(errs, fieldError("timestamp", "timestamp must not be greater or equal to 1"))
	} else if time.Unix(h.GetTimestamp(), 0).After(time.Now().Add(2 * time.Hour).UTC()) {
		errs = append(errs, fieldError("timestamp", "timestamp is over 2 hours in the future"))
	}

	// Verify the proof of work and difficulty
	if v.verSeal && !h.GetParentHash().IsEmpty() {
		errs = append(errs, v.checkPoW(h)...)
	}

	return
}

// checkPoW checks the PoW and difficulty values in the header
func (v *BlockValidator) checkPoW(h core.Header) (errs []error) {

	// get the parent header
	parentHeader, err := v.bchain.ChainReader().GetHeaderByHash(h.GetParentHash())
	if err != nil {
		errs = append(errs, fieldError("parentHash", err.Error()))
	}

	if err := v.blakimoto.VerifyHeader(v.bchain.ChainReader(), h, parentHeader, v.verSeal); err != nil {
		errs = append(errs, fieldError("parentHash", err.Error()))
	}

	return
}

// check checks the field and their values and
// does no integration checks with other components.
func (v *BlockValidator) check() (errs []error) {

	// Transaction must not be nil
	if v.block == nil {
		errs = append(errs, fmt.Errorf("nil block"))
		return
	}

	// Transaction type must be known and acceptable
	if v.block.GetHeader().(*wire.Header) == nil {
		errs = append(errs, fieldError("header", "header is required"))
	} else {
		for _, err := range v.validateHeader(v.block.GetHeader()) {
			errs = append(errs, fmt.Errorf(strings.Replace(err.Error(), "field:", "field:header.", -1)))
		}
	}

	// Must have at least one transaction, otherwise,
	// the transactions must be valid
	if len(v.block.GetTransactions()) == 0 {
		errs = append(errs, fieldError("transactions", "at least one transaction is required"))
	} else {
		txValidator := NewTxsValidator(v.block.GetTransactions(), v.txpool, v.bchain, v.allowDuplicateCheck)
		for _, err := range txValidator.Validate() {
			errs = append(errs, fmt.Errorf(strings.Replace(err.Error(), "index:", "tx:", -1)))
		}
	}

	// Hash must be provided
	if v.block.GetHash() == util.EmptyHash {
		errs = append(errs, fieldError("hash", "hash is required"))
	} else if v.block.GetHeader() != nil && !v.block.GetHash().Equal(v.block.ComputeHash()) {
		errs = append(errs, fieldError("hash", "hash is not correct"))
	}

	// Signature must be provided
	if len(v.block.GetSignature()) == 0 {
		errs = append(errs, fieldError("sig", "signature is required"))
	}

	return
}

// checkSignature checks whether the signature is valid.
// Expects the block to have a valid creators public key and
// signature to be set.
func (v *BlockValidator) checkSignature() (errs []error) {

	pubKey, err := crypto.PubKeyFromBase58(v.block.GetHeader().GetCreatorPubKey().String())
	if err != nil {
		errs = append(errs, fieldError("header.creatorPubKey", err.Error()))
		return
	}

	valid, err := pubKey.Verify(v.block.Bytes(), v.block.GetSignature())
	if err != nil {
		errs = append(errs, fieldError("sig", err.Error()))
	} else if !valid {
		errs = append(errs, fieldError("sig", "signature is not valid"))
	}

	return
}

// duplicateCheck checks whether the block exists in some
// other components that do not accept duplicates. E.g main, side chains,
// orphan index.
func (v *BlockValidator) duplicateCheck(b core.Block) (errs []error) {

	if v.bchain != nil {
		known, reason, err := v.bchain.IsKnownBlock(b.GetHash())
		if err != nil {

			errs = append(errs, fmt.Errorf("duplicate check error: %s", err))
		} else if known {
			errs = append(errs, fieldError("", fmt.Sprintf("block found in %s", reason)))
		}
	}

	return
}
