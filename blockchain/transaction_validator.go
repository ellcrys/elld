package blockchain

import (
	"fmt"

	"github.com/go-ozzo/ozzo-validation"

	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/types/core"

	"github.com/shopspring/decimal"

	"github.com/thoas/go-funk"

	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/params"
	"github.com/ellcrys/elld/util"
)

// KnownTransactionTypes are the supported transaction types
var KnownTransactionTypes = []int64{
	core.TxTypeBalance,
	core.TxTypeAlloc,
}

// TxsValidator implements a validator for checking
// syntactic, contextual and state correctness of transactions
// in relation to various parts of the system.
type TxsValidator struct {
	VContexts

	// txs are the transactions to be validated
	txs []types.Transaction

	// txpool refers to the transaction pool
	txpool types.TxPool

	// bchain is the blockchain manager. We use it
	// to query transactions
	bchain types.Blockchain

	// curIndex is the current index of the the current
	// transaction being validated.
	curIndex int

	// nonces caches valid nonces
	nonces map[string]uint64
}

func appendErr(dest []error, err error) []error {
	if err != nil {
		return append(dest, err)
	}
	return dest
}

// NewTxsValidator creates an instance of TxsValidator
func NewTxsValidator(txs []types.Transaction, txPool types.TxPool,
	bchain types.Blockchain) *TxsValidator {
	return &TxsValidator{
		txs:    txs,
		txpool: txPool,
		bchain: bchain,
		nonces: make(map[string]uint64),
	}
}

// NewTxValidator is like NewTxsValidator
// except it accepts a single transaction
func NewTxValidator(tx types.Transaction, txPool types.TxPool,
	bchain types.Blockchain) *TxsValidator {
	return &TxsValidator{
		txs:    []types.Transaction{tx},
		txpool: txPool,
		bchain: bchain,
		nonces: make(map[string]uint64),
	}
}

// Validate execute validation checks
// against each transactions
func (v *TxsValidator) Validate(opts ...types.CallOp) (errs []error) {
	var seenTxs = make(map[string]struct{})
	for i, tx := range v.txs {
		v.curIndex = i

		// check duplicate
		if _, ok := seenTxs[tx.GetHash().HexStr()]; ok {
			errs = appendErr(errs,
				fieldErrorWithIndex(v.curIndex, "", "duplicate transaction"))
			continue
		}

		txErrs := v.ValidateTx(tx, opts...)
		if len(txErrs) > 0 {
			errs = append(errs, txErrs...)
			continue
		}

		// At this point, the transaction is valid.
		// We should cache the nonce so subsequent
		// transactions of the sender retrieve the
		// latest nonce from cache instead of querying
		// the account which has not been updated with
		// the lastest nonce.
		v.nonces[tx.GetFrom().String()] = tx.GetNonce()

		// cache the hash
		seenTxs[tx.GetHash().HexStr()] = struct{}{}
	}
	return
}

// CheckFields checks validates the transaction
// fields and values.
func (v *TxsValidator) CheckFields(tx types.Transaction) (errs []error) {

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

	var isDerivedFromPublicKeyRule = func(err error) func(interface{}) error {
		return func(val interface{}) error {
			pk, _ := crypto.PubKeyFromBase58(tx.GetSenderPubKey().String())
			if !pk.Addr().Equal(val.(util.String)) {
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

	var validValueRule = func(field string) func(interface{}) error {
		return func(val interface{}) error {
			dVal, _err := decimal.NewFromString(val.(util.String).String())
			if _err != nil {
				return fieldErrorWithIndex(v.curIndex, field, "could not convert to decimal")
			}
			if dVal.LessThan(decimal.Zero) {
				return fieldErrorWithIndex(v.curIndex, field, "negative value not allowed")
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
		validation.By(validTypeRule(fieldErrorWithIndex(v.curIndex, "type",
			"unsupported transaction type"))),
	))

	// Recipient's address must be set and it must be valid
	errs = appendErr(errs, validation.Validate(tx.GetTo(),
		validation.Required.Error(fieldErrorWithIndex(v.curIndex, "to",
			"recipient address is required").Error()),
		validation.By(validAddrRule(fieldErrorWithIndex(v.curIndex, "to",
			"recipient address is not valid"))),
	))

	// Value must be >= 0 and it must be valid number
	errs = appendErr(errs, validation.Validate(tx.GetValue(),
		validation.Required.Error(fieldErrorWithIndex(v.curIndex, "value",
			"value is required").Error()),
		validation.By(validValueRule("value")),
	))

	// Timestamp is required.
	errs = appendErr(errs, validation.Validate(tx.GetTimestamp(),
		validation.Required.Error(fieldErrorWithIndex(v.curIndex, "timestamp",
			"timestamp is required").Error()),
	))

	// Sender's public key is required and must be a valid base58 encoded key
	errs = appendErr(errs, validation.Validate(tx.GetSenderPubKey(),
		validation.Required.Error(fieldErrorWithIndex(v.curIndex, "senderPubKey",
			"sender public key is required").Error()),
		validation.By(validPubKeyRule(fieldErrorWithIndex(v.curIndex, "senderPubKey",
			"sender public key is not valid"))),
	))

	// Sender's address must be set and must also be valid
	errs = appendErr(errs, validation.Validate(tx.GetFrom(),
		validation.Required.Error(fieldErrorWithIndex(v.curIndex, "from",
			"sender address is required").Error()),
		validation.By(validAddrRule(fieldErrorWithIndex(v.curIndex, "from",
			"sender address is not valid"))),
		validation.By(isDerivedFromPublicKeyRule(fieldErrorWithIndex(v.curIndex, "from",
			"sender address is not derived from the sender public key"))),
	))

	// Hash is required. It must also be correct
	errs = appendErr(errs, validation.Validate(tx.GetHash(),
		validation.By(requireHashRule(fieldErrorWithIndex(v.curIndex, "hash",
			"hash is required"))),
		validation.By(isSameHashRule(tx.ComputeHash(), fieldErrorWithIndex(v.curIndex,
			"hash", "hash is not correct"))),
	))

	// Signature must be set
	errs = appendErr(errs, validation.Validate(tx.GetSignature(),
		validation.Required.Error(fieldErrorWithIndex(v.curIndex, "sig",
			"signature is required").Error()),
	))

	// For non allocations, fee is required.
	// It must be a number. It must be equal to the
	// minimum required fee for the size of the
	// transaction.
	if tx.GetType() != core.TxTypeAlloc {
		err := validation.Validate(tx.GetFee(),
			validation.Required.Error(fieldErrorWithIndex(v.curIndex, "fee",
				"fee is required").Error()),
			validation.By(validValueRule("fee")),
		)
		errs = appendErr(errs, err)

		// Calculate and check fee only if
		// the fee passed format validation
		if err == nil {
			fee := tx.GetFee().Decimal()
			txSize := decimal.NewFromFloat(float64(tx.GetSizeNoFee()))

			// Calculate the expected fee
			expectedMinimumFee := params.FeePerByte.Mul(txSize)

			// Compare the expected fee with the provided fee
			if expectedMinimumFee.GreaterThan(fee) {
				errs = appendErr(errs, fieldErrorWithIndex(v.curIndex, "fee",
					fmt.Sprintf("fee is too low. Minimum fee expected: %s (for %s bytes)",
						expectedMinimumFee.String(), txSize.String())))
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
func (v *TxsValidator) checkSignature(tx types.Transaction) (errs []error) {

	pubKey, err := crypto.PubKeyFromBase58(tx.GetSenderPubKey().String())
	if err != nil {
		errs = append(errs, fieldErrorWithIndex(v.curIndex,
			"senderPubKey", err.Error()))
		return
	}

	valid, err := pubKey.Verify(tx.GetBytesNoHashAndSig(), tx.GetSignature())
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
// against the current state of the blockchain.
func (v *TxsValidator) consistencyCheck(tx types.Transaction, opts ...types.CallOp) (errs []error) {

	// No need for consistency check for
	// TxTypeAlloc transactions
	if tx.GetType() == core.TxTypeAlloc {
		return
	}

	// If the caller's intent is not to validate a block
	// for inclusion in a chain, then we must
	// check that the transaction does not have a
	// duplicate in the pool
	if !v.has(types.ContextBlock) && v.txpool.Has(tx) {
		errs = append(errs, fieldErrorWithIndex(v.curIndex,
			"", "transaction already exist in the transactions pool"))
		return
	}

	// If the caller intends to validate a block that was
	// received when the client is not sync with another peer,
	// the transaction must exist in the transactions pool
	if v.has(types.ContextBlock) && !v.has(types.ContextBlockSync) {
		if !v.txpool.Has(tx) {
			errs = append(errs, fieldErrorWithIndex(v.curIndex,
				"", "transaction does not exist in the transactions pool"))
		}
	}

	// No need performing nonce and balance checks for
	// transactions inside a block that may be appended
	// to a branch. This will be performed if the the
	// branch grows long enough to cause a re-org.
	if v.has(types.ContextBranch) {
		return
	}

	// If the callers intent is not to append a block
	// to the main chain, we must ensure the transaction
	// does not exist on the main chain.
	_, err := v.bchain.GetTransaction(tx.GetHash(), opts...)
	if err != nil {
		if err != core.ErrTxNotFound {
			errs = append(errs, fmt.Errorf("failed to get transaction: %s", err))
			return
		}
	} else {
		errs = append(errs, fieldErrorWithIndex(v.curIndex,
			"", "transaction already exist in main chain"))
		return
	}

	// Get the sender account
	account, err := v.bchain.GetAccount(tx.GetFrom(), opts...)
	if err != nil {
		if err == core.ErrAccountNotFound {
			errs = append(errs, fieldErrorWithIndex(v.curIndex,
				"from", "sender account not found"))
			return
		}
		errs = append(errs, fmt.Errorf("failed to get account: %s", err))
		return
	}

	// Check whether the sender has sufficient
	// balance to cover the value + fee
	deductable := tx.GetValue().Decimal().Add(tx.GetFee().Decimal())
	if account.GetBalance().Decimal().LessThan(deductable) {
		errs = append(errs, fieldErrorWithIndex(v.curIndex, "",
			fmt.Sprintf("insufficient account balance")))
		return
	}

	// Get the nonce of the cache, otherwise use
	// the nonce on the sender account
	accountNonce, ok := v.nonces[tx.GetFrom().String()]
	if !ok {
		accountNonce = account.GetNonce()
	}

	// If the caller does not intend to add the
	// transaction into a block (e.g the tx pool),
	// then the nonce must be greater than the
	// account's current nonce by at least 1
	if !v.has(types.ContextBlock) && (tx.GetNonce() <= accountNonce) {
		errs = append(errs, fieldErrorWithIndex(v.curIndex, "",
			fmt.Sprintf("invalid nonce: has %d, wants from %d",
				tx.GetNonce(), accountNonce+1)))
		return
	}

	// If the caller intends to append a block of which
	// this transaction is part of, then the nonce must
	// be greater than the account's current nonce by 1
	if v.has(types.ContextBlock) && tx.GetNonce() > accountNonce &&
		tx.GetNonce()-accountNonce != 1 {
		errs = append(errs, fieldErrorWithIndex(v.curIndex, "",
			fmt.Sprintf("invalid nonce: has %d, wants %d",
				tx.GetNonce(), accountNonce+1)))
		return
	}

	return
}

// ValidateTx validates a single transaction coming received
// by the gossip handler..
func (v *TxsValidator) ValidateTx(tx types.Transaction, opts ...types.CallOp) []error {

	errs := v.CheckFields(tx)
	if len(errs) > 0 {
		return errs
	}

	errs = append(errs, v.consistencyCheck(tx, opts...)...)
	return errs
}
