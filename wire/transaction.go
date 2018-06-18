package wire

import (
	"crypto/sha256"
	"encoding/asn1"
	"encoding/hex"
	"fmt"
	"time"

	validator "gopkg.in/asaskevich/govalidator.v4"

	"github.com/ellcrys/elld/crypto"
)

var (
	// ErrTxVerificationFailed means a transaction signature could not be verified
	ErrTxVerificationFailed = fmt.Errorf("signature verification failed")

	// ErrTxInsufficientFee means fee is insufficient
	ErrTxInsufficientFee = fmt.Errorf("insufficient fee")

	// ErrTxLowValue means transaction value is less than or equal to zero
	ErrTxLowValue = fmt.Errorf("value must be greater than zero")

	//ErrTxTypeUnknown means transaction type is unknown
	ErrTxTypeUnknown = fmt.Errorf("unknown transaction type")
)

var (
	// TxTypeBalance represents a transaction from an account to another account
	TxTypeBalance int64 = 0x1

	// TxTypeRepoCreate represents a transaction type for creating a repository
	TxTypeRepoCreate int64 = 0x2

	// TxTypeA2B represents a transaction type targeting a blockcode
	TxTypeA2B int64 = 0x3
)

type asn1Tx struct {
	Type         int64  `json:"type"`
	Nonce        int64  `json:"nonce"`
	To           string `json:"to" asn1:"utf8"`
	SenderPubKey string `json:"senderPubKey" asn1:"utf8"`
	Value        string `json:"value" asn1:"utf8"`
	Fee          string `json:"fee" asn1:"utf8"`
	Timestamp    int64  `json:"timestamp"`
}

// NewTransaction creates a new transaction
func NewTransaction(txType int64, nonce int64, to string, senderPubKey string, value string, fee string, timestamp int64) (tx *Transaction) {
	tx = new(Transaction)
	tx.Type = txType
	tx.Nonce = nonce
	tx.To = to
	tx.SenderPubKey = senderPubKey
	tx.Value = value
	tx.Timestamp = timestamp
	tx.Fee = fee
	return
}

// Bytes return the ASN.1 marshalled representation of the transaction
func (tx *Transaction) Bytes() []byte {

	asn1Tx := asn1Tx{
		Type:         tx.Type,
		Nonce:        tx.Nonce,
		To:           tx.To,
		SenderPubKey: tx.SenderPubKey,
		Value:        tx.Value,
		Fee:          tx.Fee,
		Timestamp:    tx.Timestamp,
	}

	result, err := asn1.Marshal(asn1Tx)
	if err != nil {
		panic(err)
	}
	return result
}

// GetHash returns the SHA256 hash of the transaction
func (tx *Transaction) GetHash() []byte {
	bs := tx.Bytes()
	hash := sha256.Sum256(bs)
	return hash[:]
}

// ID returns the hex representation of Hash()
func (tx *Transaction) ID() string {
	return hex.EncodeToString(tx.GetHash())
}

// TxVerify checks whether a transaction's signature is valid.
// Expect tx.SenderPubKey and tx.Sig to be set.
func TxVerify(tx *Transaction) error {

	if tx == nil {
		return fmt.Errorf("nil tx")
	}

	if tx.SenderPubKey == "" {
		return txFieldError("senderPubKey", "sender public not set")
	}

	if len(tx.Sig) == 0 {
		return txFieldError("sig", "signature not set")
	}

	pubKey, err := crypto.PubKeyFromBase58(tx.SenderPubKey)
	if err != nil {
		return txFieldError("senderPubKey", err.Error())
	}

	valid, err := pubKey.Verify(tx.Bytes(), tx.Sig)
	if err != nil {
		return txFieldError("sig", err.Error())
	}

	if !valid {
		return crypto.ErrVerifyFailed
	}

	return nil
}

// TxSign signs a transaction.
// Expects private key in base58Check encoding
func TxSign(tx *Transaction, privKey string) ([]byte, error) {

	if tx == nil {
		return nil, fmt.Errorf("nil tx")
	}

	pKey, err := crypto.PrivKeyFromBase58(privKey)
	if err != nil {
		return nil, err
	}

	tx.Sig = []byte{}
	sig, err := pKey.Sign(tx.Bytes())
	if err != nil {
		return nil, err
	}

	return sig, nil
}

func txFieldError(field, err string) error {
	return fmt.Errorf(fmt.Sprintf("field = %s, msg=%s", field, err))
}

// TxValidate validates the fields of a transaction
// - Sender public key must be set and valid
// - receiver's address must be valid
// - Timestamp cannot be a future time
// - Timestamp cannot be a week
// - Signature is required
func TxValidate(tx *Transaction) (errs []error) {

	now := time.Now()

	if validator.IsNull(tx.SenderPubKey) {
		errs = append(errs, txFieldError("senderPubKey", "sender public key is required"))
	} else if _, err := crypto.PubKeyFromBase58(tx.SenderPubKey); err != nil {
		errs = append(errs, txFieldError("senderPubKey", err.Error()))
	}

	if validator.IsNull(tx.To) {
		errs = append(errs, txFieldError("to", "recipient address is required"))
	} else if err := crypto.IsValidAddr(tx.To); err != nil {
		errs = append(errs, txFieldError("to", "address is not valid"))
	}

	if now.Before(time.Unix(tx.Timestamp, 0)) {
		errs = append(errs, txFieldError("timestamp", "timestamp cannot be a future time"))
	}

	if now.Add(-7 * 24 * time.Hour).After(time.Unix(tx.Timestamp, 0)) {
		errs = append(errs, txFieldError("timestamp", "timestamp cannot over 7 days ago"))
	}

	if len(tx.Sig) == 0 {
		errs = append(errs, txFieldError("sig", "signature is required"))
	}

	return
}
