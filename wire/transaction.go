package wire

import (
	bytes "bytes"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/shopspring/decimal"

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

	// TxTypeEndorserTicketCreate represents a transaction to create an endorser ticket
	TxTypeEndorserTicketCreate = 0x2
)

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

// Bytes return the ASN.1 marshalled representation of the transaction.
func (tx *Transaction) Bytes() []byte {

	var invokeArgsBs []byte
	if tx.InvokeArgs != nil {
		invokeArgsBs = tx.InvokeArgs.Bytes()
	}

	asn1Data := []interface{}{
		tx.Type,
		tx.Nonce,
		tx.To,
		tx.SenderPubKey,
		tx.From,
		tx.Value,
		tx.Fee,
		tx.Timestamp,
		invokeArgsBs,
	}

	return getBytes(asn1Data)
}

// ComputeHash returns the SHA256 hash of the transaction.
func (tx *Transaction) ComputeHash() []byte {
	bs := tx.Bytes()
	hash := sha256.Sum256(bs)
	return hash[:]
}

// ComputeHash2 computes the SHA256 hash of the transaction and encodes to hex.
func (tx *Transaction) ComputeHash2() string {
	bs := tx.Bytes()
	hash := sha256.Sum256(bs)
	return ToHex(hash[:])
}

// ID returns the hex representation of Hash()
func (tx *Transaction) ID() string {
	return tx.ComputeHash2()
}

// TxVerify checks whether a transaction's signature is valid.
// Expect tx.SenderPubKey and tx.Sig to be set.
func TxVerify(tx *Transaction) error {

	if tx == nil {
		return fmt.Errorf("nil tx")
	}

	if tx.SenderPubKey == "" {
		return fieldError("senderPubKey", "sender public not set")
	}

	if len(tx.Sig) == 0 {
		return fieldError("sig", "signature not set")
	}

	pubKey, err := crypto.PubKeyFromBase58(tx.SenderPubKey)
	if err != nil {
		return fieldError("senderPubKey", err.Error())
	}

	decSig, _ := FromHex(tx.Sig)
	valid, err := pubKey.Verify(tx.Bytes(), decSig)
	if err != nil {
		return fieldError("sig", err.Error())
	}

	if !valid {
		return crypto.ErrTxVerificationFailed
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

	sig, err := pKey.Sign(tx.Bytes())
	if err != nil {
		return nil, err
	}

	return sig, nil
}

// Validate the transaction
func (tx *Transaction) Validate() error {

	now := time.Now()

	if tx.Type != TxTypeBalance {
		return fieldError("type", "type is unknown")
	}

	if len(tx.SenderPubKey) == 0 {
		return fieldError("senderPubKey", "sender public key is required")
	}

	if _, err := crypto.PubKeyFromBase58(tx.SenderPubKey); err != nil {
		return fieldError("senderPubKey", err.Error())
	}

	if len(tx.To) == 0 {
		return fieldError("to", "recipient address is required")
	}

	if err := crypto.IsValidAddr(tx.To); err != nil {
		return fieldError("to", "address is not valid")
	}

	if err := crypto.IsValidAddr(tx.From); err != nil {
		return fieldError("from", "address is not valid")
	}

	if len(tx.From) == 0 {
		return fieldError("from", "sender address is required")
	}

	if senderPubKey, _ := crypto.PubKeyFromBase58(tx.SenderPubKey); senderPubKey.Addr() != tx.From {
		return fieldError("from", "address not derived from 'senderPubKey'")
	}

	if _, err := decimal.NewFromString(tx.Value); err != nil {
		return fieldError("value", "value must be numeric")
	}

	if val, _ := decimal.NewFromString(tx.Value); val.LessThanOrEqual(decimal.New(0, 0)) && tx.Type == TxTypeBalance {
		return fieldError("value", "value must be a non-zero or non-negative number")
	}

	fee, err := decimal.NewFromString(tx.Fee)
	if err != nil {
		return fieldError("fee", "fee must be numeric")
	}

	if tx.Type == TxTypeBalance && fee.LessThanOrEqual(decimal.New(0, 0)) {
		return fieldError("fee", "fee must be a non-zero or non-negative number")
	}

	if now.Before(time.Unix(tx.Timestamp, 0)) {
		return fieldError("timestamp", "timestamp cannot be a future time")
	}

	if now.Add(-7 * 24 * time.Hour).After(time.Unix(tx.Timestamp, 0)) {
		return fieldError("timestamp", "timestamp cannot over 7 days ago")
	}

	if len(tx.Hash) == 0 {
		return fieldError("hash", "hash is required")
	}

	if len(tx.Hash) < 66 {
		return fieldError("hash", "expected 66 characters")
	}

	if tx.Hash[:2] != "0x" || !bytes.Equal([]byte(ToHex(tx.ComputeHash())), []byte(tx.Hash)) {
		return fieldError("hash", "hash is not valid")
	}

	if len(tx.Sig) == 0 {
		return fieldError("sig", "signature is required")
	}

	return nil
}

// Bytes returns the byte equivalent
func (m *InvokeArgs) Bytes() []byte {
	return getBytes([]interface{}{
		m.Func,
		m.Params,
	})
}
