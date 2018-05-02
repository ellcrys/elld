package wire

import (
	"crypto/sha256"
	"encoding/asn1"
	"fmt"

	"github.com/ellcrys/druid/crypto"
)

var (
	// TxTypeRepoCreate represents a transaction type for creating a repository
	TxTypeRepoCreate int64 = 0x1
)

type asn1Tx struct {
	Type         int64
	Nonce        int64
	To           string
	SenderPubKey string
	Timestamp    int64
	Fee          string
}

// NewTransaction creates a new transaction
func NewTransaction(txType int64, nonce int64, to string, senderPubKey string, fee string, timestamp int64) (tx *Transaction) {
	tx = new(Transaction)
	tx.Type = txType
	tx.Nonce = nonce
	tx.To = to
	tx.SenderPubKey = senderPubKey
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
		Timestamp:    tx.Timestamp,
		Fee:          tx.Fee,
	}
	result, err := asn1.Marshal(asn1Tx)
	if err != nil {
		panic(err)
	}
	return result
}

// Hash returns the SHA256 hash of the transaction
func (tx *Transaction) Hash() []byte {
	bs := tx.Bytes()
	hash := sha256.Sum256(bs)
	return hash[:]
}

// TxVerify checks whether a transaction's signature is valid.
// Expect tx.SenderPubKey and tx.Sig to be set.
func TxVerify(tx *Transaction) error {

	if tx == nil {
		return fmt.Errorf("nil tx")
	}

	if tx.SenderPubKey == "" {
		return fmt.Errorf("sender public not set")
	}

	if len(tx.Sig) == 0 {
		return fmt.Errorf("signature not set")
	}

	pubKey, err := crypto.PubKeyFromBase58(tx.SenderPubKey)
	if err != nil {
		return err
	}

	valid, err := pubKey.Verify(tx.Bytes(), tx.Sig)
	if err != nil {
		return err
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
