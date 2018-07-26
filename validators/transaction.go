package validators

import (
	"fmt"
	"time"

	"github.com/ellcrys/elld/blockchain"
	"github.com/ellcrys/elld/blockchain/common"
	"github.com/ellcrys/elld/txpool"

	"github.com/shopspring/decimal"

	"github.com/thoas/go-funk"

	"github.com/ellcrys/elld/constants"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/wire"
)

// KnownTransactionTypes are the supported transaction types
var KnownTransactionTypes = []int64{wire.TxTypeBalance}

// TxsValidator implements a validator for checking
// syntactic, contextual and state correctness of transactions
// in relation to various parts of the system.
type TxsValidator struct {

	// txs are the transactions to be validated
	txs []*wire.Transaction

	// txpool refers to the transaction pool
	txpool *txpool.TxPool

	// bchain is the blockchain manager. We use it
	// to query transactions
	bchain *blockchain.Blockchain

	// allowDuplicateCheck enables duplication checks on other
	// collections. If set to true, a transaction existing in
	// a collection such as the transaction pool, chains etc
	// will be considered invalid.
	allowDuplicateCheck bool

	// currentTxIndexInLoop is the current index of the the current
	// transaction being validated.
	currentTxIndexInLoop int
}

// NewTxsValidator creates an instance of TxsValidator
func NewTxsValidator(txs []*wire.Transaction, txPool *txpool.TxPool,
	bchain *blockchain.Blockchain, allowDupCheck bool) *TxsValidator {
	return &TxsValidator{
		txs:                 txs,
		txpool:              txPool,
		bchain:              bchain,
		allowDuplicateCheck: allowDupCheck,
	}
}

// NewTxValidator is like NewTxsValidator except it accepts a single transaction
func NewTxValidator(tx *wire.Transaction, txPool *txpool.TxPool,
	bchain *blockchain.Blockchain, allowDupCheck bool) *TxsValidator {
	return &TxsValidator{
		txs:                 []*wire.Transaction{tx},
		txpool:              txPool,
		bchain:              bchain,
		allowDuplicateCheck: allowDupCheck,
	}
}

// Validate execute validation checks on the
// transactions, returning all the errors encountered. This
func (v *TxsValidator) Validate() (errs []error) {
	for i, tx := range v.txs {
		v.currentTxIndexInLoop = i
		errs = append(errs, v.ValidateTx(tx)...)
	}
	return
}

// statelessChecks checks the field and their values and
// does no integration checks with other components.
func (v *TxsValidator) statelessChecks(tx *wire.Transaction) (errs []error) {

	// Transaction must not be nil
	if tx == nil {
		errs = append(errs, fmt.Errorf("nil tx"))
		return
	}

	// Transaction type must be known and acceptable
	if !funk.ContainsInt64(KnownTransactionTypes, tx.Type) {
		errs = append(errs, fieldErrorWithIndex(v.currentTxIndexInLoop,
			"type", "unsupported transaction type"))
	}

	// Nonce must be a non-negative integer
	if tx.Nonce < 0 {
		errs = append(errs, fieldErrorWithIndex(v.currentTxIndexInLoop,
			"nonce", "nonce must be non-negative"))
	}

	// Recipient's address must be set and it must be valid
	if tx.To == "" {
		errs = append(errs, fieldErrorWithIndex(v.currentTxIndexInLoop,
			"to", "recipient address is required"))
	} else if err := crypto.IsValidAddr(tx.To); err != nil {
		errs = append(errs, fieldErrorWithIndex(v.currentTxIndexInLoop,
			"to", "recipient address is not valid"))
	}

	// Sender's address must be set and it must be valid
	if tx.From == "" {
		errs = append(errs, fieldErrorWithIndex(v.currentTxIndexInLoop,
			"from", "sender address is required"))
	} else if err := crypto.IsValidAddr(tx.From); err != nil {
		errs = append(errs, fieldErrorWithIndex(v.currentTxIndexInLoop,
			"from", "sender address is not valid"))
	}

	// Sender public key is required and must be valid.
	if tx.SenderPubKey == "" {
		errs = append(errs, fieldErrorWithIndex(v.currentTxIndexInLoop,
			"senderPubKey", "sender public key is required"))
	} else if _, err := crypto.PubKeyFromBase58(tx.SenderPubKey); err != nil {
		errs = append(errs, fieldErrorWithIndex(v.currentTxIndexInLoop,
			"senderPubKey", "sender public key is not valid"))
	}

	// For balance transactions, value cannot be an empty string
	// and it must be convertible to decimal and not a zero value.
	if tx.Type == wire.TxTypeBalance {
		if tx.Value == "" {
			errs = append(errs, fieldErrorWithIndex(v.currentTxIndexInLoop,
				"value", "value is required"))
		}
		_value, err := decimal.NewFromString(tx.Value)
		if err != nil {
			errs = append(errs, fieldErrorWithIndex(v.currentTxIndexInLoop,
				"value", "could not convert to decimal"))
		}
		if _value.LessThanOrEqual(decimal.Zero) {
			errs = append(errs, fieldErrorWithIndex(v.currentTxIndexInLoop,
				"value", "value must be greater than zero"))
		}
	}

	// Timestamp is required and cannot be a time in the future
	if tx.Timestamp == 0 {
		errs = append(errs, fieldErrorWithIndex(v.currentTxIndexInLoop,
			"timestamp", "timestamp is required"))
	} else if tx.Timestamp > time.Now().Unix() {
		errs = append(errs, fieldErrorWithIndex(v.currentTxIndexInLoop,
			"timestamp", "timestamp cannot be a future time"))
	}

	// Fee cannot be empty or less than 0
	if tx.Fee == "" {
		errs = append(errs, fieldErrorWithIndex(v.currentTxIndexInLoop,
			"fee", "fee is required"))
	} else {
		fee, err := decimal.NewFromString(tx.Fee)
		if err != nil {
			errs = append(errs, fieldErrorWithIndex(v.currentTxIndexInLoop,
				"fee", "could not convert to decimal"))
		} else if tx.Type == wire.TxTypeBalance && fee.LessThanOrEqual(constants.BalanceTxMinimumFee) {
			errs = append(errs, fieldErrorWithIndex(v.currentTxIndexInLoop,
				"fee", fmt.Sprintf("fee cannot be below the minimum balance transaction fee {%s}", constants.BalanceTxMinimumFee.StringFixed(8))))
		}
	}

	// Ensure the transaction hash is provided and
	// that the hash matches the computed value.
	if tx.Hash == "" {
		errs = append(errs, fieldErrorWithIndex(v.currentTxIndexInLoop,
			"hash", "hash is required"))
	} else if tx.Hash != tx.ComputeHash2() {
		errs = append(errs, fieldErrorWithIndex(v.currentTxIndexInLoop,
			"hash", "hash is not correct"))
	}

	// Check that signature is provided
	if tx.Sig == "" {
		errs = append(errs, fieldErrorWithIndex(v.currentTxIndexInLoop,
			"sig", "signature is required"))
	}

	return
}

// checkSignature checks whether the signature is valid.
// Expects the transaction to have a valid sender public key
func (v *TxsValidator) checkSignature(tx *wire.Transaction) (errs []error) {

	pubKey, err := crypto.PubKeyFromBase58(tx.SenderPubKey)
	if err != nil {
		errs = append(errs, fieldErrorWithIndex(v.currentTxIndexInLoop,
			"senderPubKey", err.Error()))
		return
	}

	decSig, _ := util.FromHex(tx.Sig)
	valid, err := pubKey.Verify(tx.Bytes(), decSig)
	if err != nil {
		errs = append(errs, fieldErrorWithIndex(v.currentTxIndexInLoop,
			"sig", err.Error()))
	} else if !valid {
		errs = append(errs, fieldErrorWithIndex(v.currentTxIndexInLoop,
			"sig", "signature is not valid"))
	}

	return
}

// duplicateCheck checks whether the transaction exists in some
// other components that do not accept duplicates. E.g transaction
// pool, chains etc
func (v *TxsValidator) duplicateCheck(tx *wire.Transaction) (errs []error) {

	if v.txpool != nil && v.txpool.Has(tx) {
		errs = append(errs, fieldErrorWithIndex(v.currentTxIndexInLoop,
			"", "transaction already exist in tx pool"))
		return
	}

	if v.bchain != nil {
		_, err := v.bchain.GetTransaction(tx.Hash)
		if err != nil {
			if err != common.ErrTxNotFound {
				errs = append(errs, fmt.Errorf("get transaction error: %s", err))
			}
		} else {
			errs = append(errs, fieldErrorWithIndex(v.currentTxIndexInLoop,
				"", "transaction already exist in main chain"))
		}
	}

	return
}

// ValidateTx validates a single transaction coming received
// by the gossip handler..
func (v *TxsValidator) ValidateTx(tx *wire.Transaction) []error {
	errs := v.statelessChecks(tx)
	errs = append(errs, v.checkSignature(tx)...)
	if v.allowDuplicateCheck {
		errs = append(errs, v.duplicateCheck(tx)...)
	}
	return errs
}
