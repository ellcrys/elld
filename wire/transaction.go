package wire

import (
	"crypto/sha256"
	"fmt"

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
	TxTypeEndorserTicketCreate int64 = 0x2
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

// NewTx creates a new, signed transaction
func NewTx(txType int64, nonce int64, to string, senderKey *crypto.Key, value string, fee string, timestamp int64) (tx *Transaction) {
	tx = new(Transaction)
	tx.Type = txType
	tx.Nonce = nonce
	tx.To = to
	tx.SenderPubKey = senderKey.PubKey().Base58()
	tx.From = senderKey.Addr()
	tx.Value = value
	tx.Timestamp = timestamp
	tx.Fee = fee
	tx.Hash = tx.ComputeHash2()

	sig, err := TxSign(tx, senderKey.PrivKey().Base58())
	if err != nil {
		panic(err)
	}
	tx.Sig = ToHex(sig)
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

// Sign the transaction
func (tx *Transaction) Sign(privKey string) ([]byte, error) {
	return TxSign(tx, privKey)
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

// Bytes returns the byte equivalent
func (m *InvokeArgs) Bytes() []byte {
	return getBytes([]interface{}{
		m.Func,
		m.Params,
	})
}
