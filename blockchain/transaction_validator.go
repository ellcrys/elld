package blockchain

import (
	"fmt"

	"github.com/go-ozzo/ozzo-validation"

	"github.com/ellcrys/elld/blockchain/common"
	"github.com/ellcrys/elld/txpool"
	"github.com/ellcrys/elld/types"

	"github.com/shopspring/decimal"

	"github.com/thoas/go-funk"

	"github.com/ellcrys/elld/constants"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/wire"
)

// KnownTransactionTypes are the supported transaction types
var KnownTransactionTypes = []int64{
	wire.TxTypeBalance,
	wire.TxTypeAllocCoin,
}

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
	bchain types.Blockchain

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
	bchain types.Blockchain, allowDupCheck bool) *TxsValidator {
	return &TxsValidator{
		txs:                 txs,
		txpool:              txPool,
		bchain:              bchain,
		allowDuplicateCheck: allowDupCheck,
	}
}

// NewTxValidator is like NewTxsValidator except it accepts a single transaction
func NewTxValidator(tx *wire.Transaction, txPool *txpool.TxPool,
	bchain types.Blockchain, allowDupCheck bool) *TxsValidator {
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

func appendErr(dest []error, err error) []error {
	if err != nil {
		return append(dest, err)
	}
	return dest
}

// check check the field and their values and
// does no integration check with other components.
func (v *TxsValidator) check(tx *wire.Transaction) (errs []error) {

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
			if _err := crypto.IsValidAddr(val.(string)); _err != nil {
				return err
			}
			return nil
		}
	}

	var validPubKeyRule = func(err error) func(interface{}) error {
		return func(val interface{}) error {
			if _, _err := crypto.PubKeyFromBase58(val.(string)); _err != nil {
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
			if _, _err := decimal.NewFromString(val.(string)); _err != nil {
				return err
			}
			return nil
		}
	}

	var isZeroLessRule = func(err error) func(interface{}) error {
		return func(val interface{}) error {
			dec, _ := decimal.NewFromString(val.(string))
			if dec.LessThanOrEqual(decimal.Zero) {
				return err
			}
			return nil
		}
	}

	var isValidFeeRule = func(err error) func(interface{}) error {
		return func(val interface{}) error {
			dec, _ := decimal.NewFromString(val.(string))
			if dec.LessThan(constants.BalanceTxMinimumFee) {
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

	var isSameStrRule = func(val2 string, err error) func(interface{}) error {
		return func(val interface{}) error {
			if val.(string) != val2 {
				return err
			}
			return nil
		}
	}

	// Transaction type is required and must match the known types
	errs = appendErr(errs, validation.Validate(tx.Type,
		validation.By(validTypeRule(fieldErrorWithIndex(v.currentTxIndexInLoop, "type", "unsupported transaction type"))),
	))

	// Nonce must be a non-negative integer
	errs = appendErr(errs, validation.Validate(tx.Nonce,
		validation.Min(0).Error(fieldErrorWithIndex(v.currentTxIndexInLoop, "nonce", "nonce must be non-negative").Error()),
	))

	// Recipient's address must be set and it must be valid
	errs = appendErr(errs, validation.Validate(tx.To,
		validation.Required.Error(fieldErrorWithIndex(v.currentTxIndexInLoop, "to", "recipient address is required").Error()),
		validation.By(validAddrRule(fieldErrorWithIndex(v.currentTxIndexInLoop, "to", "recipient address is not valid"))),
	))

	// Sender's address must be set and it must be valid non-zero decimal
	errs = appendErr(errs, validation.Validate(tx.Value,
		validation.Required.Error(fieldErrorWithIndex(v.currentTxIndexInLoop, "value", "value is required").Error()),
		validation.By(validValueRule(fieldErrorWithIndex(v.currentTxIndexInLoop, "value", "could not convert to decimal"))),
		validation.By(isZeroLessRule(fieldErrorWithIndex(v.currentTxIndexInLoop, "value", "value must be greater than zero"))),
	))

	// Timestamp is required.
	errs = appendErr(errs, validation.Validate(tx.Timestamp,
		validation.Required.Error(fieldErrorWithIndex(v.currentTxIndexInLoop, "timestamp", "timestamp is required").Error()),
	))

	// Sender's address must be set and must also be valid
	errs = appendErr(errs, validation.Validate(tx.From,
		validation.Required.Error(fieldErrorWithIndex(v.currentTxIndexInLoop, "from", "sender address is required").Error()),
		validation.By(validAddrRule(fieldErrorWithIndex(v.currentTxIndexInLoop, "from", "sender address is not valid"))),
	))

	// Sender's public key is required and must be a valid base58 encoded key
	errs = appendErr(errs, validation.Validate(tx.SenderPubKey,
		validation.Required.Error(fieldErrorWithIndex(v.currentTxIndexInLoop, "senderPubKey", "sender public key is required").Error()),
		validation.By(validPubKeyRule(fieldErrorWithIndex(v.currentTxIndexInLoop, "senderPubKey", "sender public key is not valid"))),
	))

	// Hash is required. It must also be correct
	errs = appendErr(errs, validation.Validate(tx.Hash,
		validation.By(requireHashRule(fieldErrorWithIndex(v.currentTxIndexInLoop, "hash", "hash is required"))),
		validation.By(isSameHashRule(tx.ComputeHash(), fieldErrorWithIndex(v.currentTxIndexInLoop, "hash", "hash is not correct"))),
	))

	// Signature must be set
	errs = appendErr(errs, validation.Validate(tx.Sig,
		validation.Required.Error(fieldErrorWithIndex(v.currentTxIndexInLoop, "sig", "signature is required").Error()),
	))

	if tx.Type == wire.TxTypeBalance {
		errs = appendErr(errs, validation.Validate(tx.Fee,
			validation.Required.Error(fieldErrorWithIndex(v.currentTxIndexInLoop, "fee", "fee is required").Error()),
			validation.By(validValueRule(fieldErrorWithIndex(v.currentTxIndexInLoop, "fee", "could not convert to decimal"))),
			validation.By(isValidFeeRule(fieldErrorWithIndex(v.currentTxIndexInLoop, "fee", fmt.Sprintf("fee cannot be below the minimum balance transaction fee {%s}", constants.BalanceTxMinimumFee.StringFixed(16))))),
		))
	}

	if tx.Type == wire.TxTypeAllocCoin {
		// Transaction sender must be the same as the recipient
		errs = appendErr(errs, validation.Validate(tx.From,
			validation.By(isSameStrRule(tx.To, fieldErrorWithIndex(v.currentTxIndexInLoop, "from", "sender and recipient must be same address"))),
		))
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

	valid, err := pubKey.Verify(tx.Bytes(), tx.Sig)
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
		_, err := v.bchain.GetTransaction(tx.Hash.HexStr())
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
	errs := v.check(tx)
	errs = append(errs, v.checkSignature(tx)...)
	if v.allowDuplicateCheck {
		errs = append(errs, v.duplicateCheck(tx)...)
	}
	return errs
}
