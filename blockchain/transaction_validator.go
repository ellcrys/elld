package blockchain

import (
	"fmt"

	"github.com/go-ozzo/ozzo-validation"

	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/types/core"
	"github.com/ellcrys/elld/types/core/objects"

	"github.com/shopspring/decimal"

	"github.com/thoas/go-funk"

	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/params"
	"github.com/ellcrys/elld/util"
)

// KnownTransactionTypes are the supported transaction types
var KnownTransactionTypes = []int64{
	objects.TxTypeBalance,
	objects.TxTypeAlloc,
}

// TxsValidator implements a validator for checking
// syntactic, contextual and state correctness of transactions
// in relation to various parts of the system.
type TxsValidator struct {

	// txs are the transactions to be validated
	txs []core.Transaction

	// txpool refers to the transaction pool
	txpool types.TxPool

	// bchain is the blockchain manager. We use it
	// to query transactions
	bchain core.Blockchain

	// curIndex is the current index of the the current
	// transaction being validated.
	curIndex int

	// ctx is the current validation context
	ctx ValidationContext
}

func appendErr(dest []error, err error) []error {
	if err != nil {
		return append(dest, err)
	}
	return dest
}

// NewTxsValidator creates an instance of TxsValidator
func NewTxsValidator(txs []core.Transaction, txPool types.TxPool,
	bchain core.Blockchain, allowDupCheck bool) *TxsValidator {
	return &TxsValidator{
		txs:    txs,
		txpool: txPool,
		bchain: bchain,
	}
}

// NewTxValidator is like NewTxsValidator except it accepts a single transaction
func NewTxValidator(tx core.Transaction, txPool types.TxPool,
	bchain core.Blockchain) *TxsValidator {
	return &TxsValidator{
		txs:    []core.Transaction{tx},
		txpool: txPool,
		bchain: bchain,
	}
}

// SetContext sets the validation context
func (v *TxsValidator) SetContext(ctx ValidationContext) {
	v.ctx = ctx
}

// Validate execute validation checks on the
// transactions, returning all the errors encountered. This
func (v *TxsValidator) Validate(opts ...core.CallOp) (errs []error) {
	for i, tx := range v.txs {
		v.curIndex = i
		errs = append(errs, v.ValidateTx(tx, opts...)...)
	}
	return
}

// fieldsCheck checks validates the transaction
// fields and values.
func (v *TxsValidator) fieldsCheck(tx core.Transaction) (errs []error) {

	// Transaction must not be nil
	if tx == nil {
		errs = append(errs, fmt.Errorf("nil tx"))
		return
	}

	var validTypeRule = func(err error) func(interface{}) error {
		return func(val interface{}) error {
			if !funk.ContainsInt64(KnownTransactionTypes, val.(int64)) {
				return err
			}
			return nil
		}
	}

	var validAddrRule = func(err error) func(interface{}) error {
		return func(val interface{}) error {
			if _err := crypto.IsValidAddr(val.(util.String).String()); _err != nil {
				return err
			}
			return nil
		}
	}

	var validPubKeyRule = func(err error) func(interface{}) error {
		return func(val interface{}) error {
			if _, _err := crypto.PubKeyFromBase58(val.(util.String).String()); _err != nil {
				return err
			}
			return nil
		}
	}

	var requireHashRule = func(err error) func(interface{}) error {
		return func(val interface{}) error {
			if val.(util.Hash).IsEmpty() {
				return err
			}
			return nil
		}
	}

	var validValueRule = func(err error) func(interface{}) error {
		return func(val interface{}) error {
			if _, _err := decimal.NewFromString(val.(util.String).String()); _err != nil {
				return err
			}
			return nil
		}
	}

	var isZeroLessRule = func(err error) func(interface{}) error {
		return func(val interface{}) error {
			dec, _ := decimal.NewFromString(val.(util.String).String())
			if dec.LessThanOrEqual(decimal.Zero) {
				return err
			}
			return nil
		}
	}

	var isSameHashRule = func(val2 util.Hash, err error) func(interface{}) error {
		return func(val interface{}) error {
			if !val.(util.Hash).Equal(val2) {
				return err
			}
			return nil
		}
	}

	// Transaction type is required and must match the known types
	errs = appendErr(errs, validation.Validate(tx.GetType(),
		validation.By(validTypeRule(fieldErrorWithIndex(v.curIndex, "type", "unsupported transaction type"))),
	))

	// Recipient's address must be set and it must be valid
	errs = appendErr(errs, validation.Validate(tx.GetTo(),
		validation.Required.Error(fieldErrorWithIndex(v.curIndex, "to", "recipient address is required").Error()),
		validation.By(validAddrRule(fieldErrorWithIndex(v.curIndex, "to", "recipient address is not valid"))),
	))

	// Value must be set and it must be valid number
	errs = appendErr(errs, validation.Validate(tx.GetValue(),
		validation.Required.Error(fieldErrorWithIndex(v.curIndex, "value", "value is required").Error()),
		validation.By(validValueRule(fieldErrorWithIndex(v.curIndex, "value", "could not convert to decimal"))),
	))

	// For non-allocations, the value must be greater than 0
	if tx.GetType() != objects.TxTypeAlloc {
		errs = appendErr(errs, validation.Validate(tx.GetValue(),
			validation.By(isZeroLessRule(fieldErrorWithIndex(v.curIndex, "value", "value must be greater than zero"))),
		))
	}

	// Timestamp is required.
	errs = appendErr(errs, validation.Validate(tx.GetTimestamp(),
		validation.Required.Error(fieldErrorWithIndex(v.curIndex, "timestamp", "timestamp is required").Error()),
	))

	// Sender's address must be set and must also be valid
	errs = appendErr(errs, validation.Validate(tx.GetFrom(),
		validation.Required.Error(fieldErrorWithIndex(v.curIndex, "from", "sender address is required").Error()),
		validation.By(validAddrRule(fieldErrorWithIndex(v.curIndex, "from", "sender address is not valid"))),
	))

	// Sender's public key is required and must be a valid base58 encoded key
	errs = appendErr(errs, validation.Validate(tx.GetSenderPubKey(),
		validation.Required.Error(fieldErrorWithIndex(v.curIndex, "senderPubKey", "sender public key is required").Error()),
		validation.By(validPubKeyRule(fieldErrorWithIndex(v.curIndex, "senderPubKey", "sender public key is not valid"))),
	))

	// Hash is required. It must also be correct
	errs = appendErr(errs, validation.Validate(tx.GetHash(),
		validation.By(requireHashRule(fieldErrorWithIndex(v.curIndex, "hash", "hash is required"))),
		validation.By(isSameHashRule(tx.ComputeHash(), fieldErrorWithIndex(v.curIndex, "hash", "hash is not correct"))),
	))

	// Signature must be set
	errs = appendErr(errs, validation.Validate(tx.GetSignature(),
		validation.Required.Error(fieldErrorWithIndex(v.curIndex, "sig", "signature is required").Error()),
	))

	// For non allocations, fee is required.
	// It must be a number. It must be equal to the
	// minimum required fee for the size of the
	// transaction.
	if tx.GetType() != objects.TxTypeAlloc {
		err := validation.Validate(tx.GetFee(),
			validation.Required.Error(fieldErrorWithIndex(v.curIndex, "fee", "fee is required").Error()),
			validation.By(validValueRule(fieldErrorWithIndex(v.curIndex, "fee", "could not convert to decimal"))),
		)
		errs = appendErr(errs, err)

		// Calculate and check fee only if
		// the fee passed format validation
		if err == nil {
			fee := tx.GetFee().Decimal()
			txSize := decimal.NewFromFloat(float64(tx.SizeNoFee()))
			expectedMinimumFee := params.FeePerByte.Mul(txSize).Round(2)
			if expectedMinimumFee.GreaterThan(fee) {
				errs = appendErr(errs, fieldErrorWithIndex(v.curIndex, "fee",
					fmt.Sprintf("fee is too low. Minimum fee expected: %s (for %s bytes)",
						expectedMinimumFee.Round(3), txSize.String())))
			}
		}
	}

	// Check signature validity
	if sigErr := v.checkSignature(tx); len(sigErr) > 0 {
		errs = append(errs, sigErr...)
	}

	return
}

// checkSignature checks whether the signature is valid.
// Expects the transaction to have a valid sender public key
func (v *TxsValidator) checkSignature(tx core.Transaction) (errs []error) {

	pubKey, err := crypto.PubKeyFromBase58(tx.GetSenderPubKey().String())
	if err != nil {
		errs = append(errs, fieldErrorWithIndex(v.curIndex,
			"senderPubKey", err.Error()))
		return
	}

	valid, err := pubKey.Verify(tx.Bytes(), tx.GetSignature())
	if err != nil {
		errs = append(errs, fieldErrorWithIndex(v.curIndex,
			"sig", err.Error()))
	} else if !valid {
		errs = append(errs, fieldErrorWithIndex(v.curIndex,
			"sig", "signature is not valid"))
	}

	return
}

// consistencyCheck checks whether the transaction
// exist as a duplicate in the main chain or in the
// transaction pool. It also performs nonce checks.
func (v *TxsValidator) consistencyCheck(tx core.Transaction, opts ...core.CallOp) (errs []error) {

	if v.txpool.Has(tx) {
		errs = append(errs, fieldErrorWithIndex(v.curIndex,
			"", "transaction already exist in the transactions pool"))
		return
	}

	// Ensure the transaction does not exist
	// on the main chain
	_, err := v.bchain.GetTransaction(tx.GetHash(), opts...)
	if err != nil {
		if err != core.ErrTxNotFound {
			errs = append(errs, fmt.Errorf("failed to get transaction: %s", err))
			return
		}
	} else {
		errs = append(errs, fieldErrorWithIndex(v.curIndex,
			"", "transaction already exist in main chain"))
	}

	// Validate nonce for non-TxTypeAlloc transactions
	if tx.GetType() == objects.TxTypeAlloc {
		return
	}

	// Get the nonce of the originator account
	accountNonce, err := v.bchain.GetAccountNonce(tx.GetFrom(), opts...)
	if err != nil {
		if err == core.ErrAccountNotFound {
			errs = append(errs, fieldErrorWithIndex(v.curIndex,
				"from", "sender account not found"))
			return
		}
		errs = append(errs, fmt.Errorf("failed to get account: %s", err))
		return
	}

	// For transactions intended to the added into
	// the transaction pool, their nonce must be greater than
	// the account's current nonce value by at least 1
	if v.ctx != ContextBlock && tx.GetNonce()-accountNonce < 1 {
		errs = append(errs, fieldErrorWithIndex(v.curIndex, "",
			fmt.Sprintf("invalid nonce: has %d, wants from %d", tx.GetNonce(), accountNonce+1)))
		return
	}

	// For transactions in blocks that will be appended to a
	// a chain, their nonce must be greater than the account's
	// current nonce value by a maximum of 1
	if v.ctx == ContextBlock && tx.GetNonce()-accountNonce != 1 {
		errs = append(errs, fieldErrorWithIndex(v.curIndex, "",
			fmt.Sprintf("invalid nonce: has %d, wants %d", tx.GetNonce(), accountNonce+1)))
		return
	}

	return
}

// ValidateTx validates a single transaction coming received
// by the gossip handler..
func (v *TxsValidator) ValidateTx(tx core.Transaction, opts ...core.CallOp) []error {
	errs := v.fieldsCheck(tx)
	errs = append(errs, v.consistencyCheck(tx, opts...)...)
	return errs
}
