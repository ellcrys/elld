package core

import (
	"encoding/json"
	"fmt"

	"github.com/btcsuite/btcutil/base58"
	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/util"
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

	// TxTypeAlloc represents a transaction to alloc coins to an account
	TxTypeAlloc int64 = 0x2
)

// Base58CheckVersionTxPayload is the base58 encode version adopted
// for compressed transaction payload
var Base58CheckVersionTxPayload byte = 95

// InvokeArgs describes a function to be executed by a blockcode
type InvokeArgs struct {
	Func   string            `json:"func" msgpack:"func"`
	Params map[string][]byte `json:"params" msgpack:"params"`
}

// Transaction represents a transaction
type Transaction struct {
	Type         int64       `json:"type" msgpack:"type"`
	Nonce        uint64      `json:"nonce" msgpack:"nonce"`
	To           util.String `json:"to" msgpack:"to"`
	From         util.String `json:"from" msgpack:"from"`
	SenderPubKey util.String `json:"senderPubKey" msgpack:"senderPubKey"`
	Value        util.String `json:"value" msgpack:"value"`
	Timestamp    int64       `json:"timestamp" msgpack:"timestamp"`
	Fee          util.String `json:"fee" msgpack:"fee"`
	InvokeArgs   *InvokeArgs `json:"invokeArgs,omitempty" msgpack:"invokeArgs"`
	Sig          []byte      `json:"sig" msgpack:"sig"`
	Hash         util.Hash   `json:"hash" msgpack:"hash"`
}

// NewTransaction creates a new transaction
func NewTransaction(txType int64, nonce uint64, to util.String,
	senderPubKey util.String, value util.String,
	fee util.String, timestamp int64) (tx *Transaction) {
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
func NewTx(txType int64, nonce uint64, to util.String, senderKey *crypto.Key, value util.String, fee util.String, timestamp int64) (tx *Transaction) {
	tx = new(Transaction)
	tx.Type = txType
	tx.Nonce = nonce
	tx.To = to
	tx.SenderPubKey = util.String(senderKey.PubKey().Base58())
	tx.From = util.String(senderKey.Addr())
	tx.Value = value
	tx.Timestamp = timestamp
	tx.Fee = fee
	tx.Hash = tx.ComputeHash()

	sig, err := TxSign(tx, senderKey.PrivKey().Base58())
	if err != nil {
		panic(err)
	}
	tx.Sig = sig
	return
}

// SetFrom sets the sender
func (tx *Transaction) SetFrom(from util.String) {
	tx.From = from
}

// GetSignature gets the signature
func (tx *Transaction) GetSignature() []byte {
	return tx.Sig
}

// SetSignature sets the signature
func (tx *Transaction) SetSignature(s []byte) {
	tx.Sig = s
}

// GetSenderPubKey gets the sender public key
func (tx *Transaction) GetSenderPubKey() util.String {
	return tx.SenderPubKey
}

// SetSenderPubKey sets the sender public key
func (tx *Transaction) SetSenderPubKey(pk util.String) {
	tx.SenderPubKey = pk
}

// ToBase58 returns base58 encoded equivalent of the transaction
func (tx *Transaction) ToBase58() string {
	bs, _ := json.Marshal(tx)
	return base58.CheckEncode(bs, Base58CheckVersionTxPayload)
}

// GetTimestamp gets the timestamp
func (tx *Transaction) GetTimestamp() int64 {
	return tx.Timestamp
}

// SetTimestamp set the unix timestamp
func (tx *Transaction) SetTimestamp(t int64) {
	tx.Timestamp = t
}

// GetNonce gets the nonce
func (tx *Transaction) GetNonce() uint64 {
	return tx.Nonce
}

// GetFee gets the value
func (tx *Transaction) GetFee() util.String {
	return tx.Fee
}

// GetValue gets the value
func (tx *Transaction) GetValue() util.String {
	return tx.Value
}

// SetValue gets the value
func (tx *Transaction) SetValue(v util.String) {
	tx.Value = v
}

// GetTo gets the address of receiver
func (tx *Transaction) GetTo() util.String {
	return tx.To
}

// GetFrom gets the address of sender
func (tx *Transaction) GetFrom() util.String {
	return tx.From
}

// GetHash returns the hash of tx
func (tx *Transaction) GetHash() util.Hash {
	return tx.Hash
}

// SetHash sets the hash
func (tx *Transaction) SetHash(h util.Hash) {
	tx.Hash = h
}

// GetType gets the transaction type
func (tx *Transaction) GetType() int64 {
	return tx.Type
}

// GetBytesNoHashAndSig converts a transaction
// to bytes equivalent but omits the hash and
// signature in the result.
func (tx *Transaction) GetBytesNoHashAndSig() []byte {

	var invokeArgsBs []byte
	if tx.InvokeArgs != nil {
		invokeArgsBs = tx.InvokeArgs.Bytes()
	}

	data := []interface{}{
		tx.Fee,
		tx.From,
		invokeArgsBs,
		tx.Nonce,
		tx.SenderPubKey,
		tx.Timestamp,
		tx.To,
		tx.Type,
		tx.Value,
	}

	return getBytes(data)
}

// Bytes converts a transaction
// to bytes equivalent
func (tx *Transaction) Bytes() []byte {

	var invokeArgsBs []byte
	if tx.InvokeArgs != nil {
		invokeArgsBs = tx.InvokeArgs.Bytes()
	}

	data := []interface{}{
		tx.Fee,
		tx.From,
		tx.Hash,
		invokeArgsBs,
		tx.Nonce,
		tx.SenderPubKey,
		tx.Sig,
		tx.Timestamp,
		tx.To,
		tx.Type,
		tx.Value,
	}

	return getBytes(data)
}

// GetSizeNoFee returns the virtual size of the
// transaction by summing up the size of
// field contents except the `fee` field.
// The value does not represent the true size
// of the transaction
// on disk. It is main required for computing
// minimum fees etc.
func (tx *Transaction) GetSizeNoFee() int64 {

	var invokeArgsBs []byte
	if tx.InvokeArgs != nil {
		invokeArgsBs = tx.InvokeArgs.Bytes()
	}

	data := []interface{}{
		tx.From,
		tx.Hash,
		invokeArgsBs,
		tx.Nonce,
		tx.SenderPubKey,
		tx.Sig,
		tx.Timestamp,
		tx.To,
		tx.Type,
		tx.Value,
	}

	return int64(len(getBytes(data)))
}

// ComputeHash returns the SHA256
// hash of the transaction.
func (tx *Transaction) ComputeHash() util.Hash {
	bs := tx.GetBytesNoHashAndSig()
	hash := util.Blake2b256(bs)
	return util.BytesToHash(hash[:])
}

// GetID returns the hex representation of Hash()
func (tx *Transaction) GetID() string {
	return tx.ComputeHash().HexStr()
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

	pubKey, err := crypto.PubKeyFromBase58(string(tx.SenderPubKey))
	if err != nil {
		return fieldError("senderPubKey", err.Error())
	}

	valid, err := pubKey.Verify(tx.GetBytesNoHashAndSig(), tx.Sig)
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

	sig, err := pKey.Sign(tx.GetBytesNoHashAndSig())
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
