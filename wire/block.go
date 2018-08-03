package wire

import (
	"crypto/sha256"
	"encoding/asn1"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"

	"github.com/ellcrys/elld/crypto"
	"gopkg.in/asaskevich/govalidator.v4"
)

// Block represents a block
type Block struct {
	Header       *Header        `json:"header" msgpack:"header"`
	Transactions []*Transaction `json:"transactions" msgpack:"transactions"`
	Hash         string         `json:"hash" msgpack:"hash"`
	Sig          string         `json:"sig" msgpack:"sig"`
}

// Header represents the header of a block
type Header struct {
	ParentHash       string `json:"ParentHash" msgpack:"ParentHash"`
	CreatorPubKey    string `json:"creatorPubKey" msgpack:"creatorPubKey"`
	Number           uint64 `json:"number" msgpack:"number"`
	StateRoot        string `json:"stateRoot" msgpack:"stateRoot"`
	TransactionsRoot string `json:"transactionsRoot" msgpack:"transactionsRoot"`
	Nonce            uint64 `json:"nonce" msgpack:"nonce"`
	MixHash          string `json:"mixHash" msgpack:"mixHash"`
	Difficulty       string `json:"difficulty" msgpack:"difficulty"`
	Timestamp        int64  `json:"timestamp" msgpack:"timestamp"`
}

// GetNumber returns the header number which is the block number
func (h *Header) GetNumber() uint64 {
	return h.Number
}

// GetHash returns the block's hash
func (b *Block) GetHash() string {
	return b.Hash
}

// BlockFromString unmarshal a json string into a Block
func BlockFromString(str string) (*Block, error) {
	var block Block
	var err error
	err = json.Unmarshal([]byte(str), &block)
	return &block, err
}

// HashNoNonce returns the hash which is used as input for the proof-of-work search.
func (h *Header) HashNoNonce() Hash {

	asn1Data := []interface{}{
		h.ParentHash,
		strconv.FormatUint(h.Number, 10),
		h.TransactionsRoot,
		h.StateRoot,
		h.Difficulty,
		h.Timestamp,
	}

	result, err := asn1.Marshal(asn1Data)
	if err != nil {
		panic(err)
	}

	return sha256.Sum256(result)
}

// Validate the header
func (h *Header) Validate() error {

	if h.Number != 1 && len(h.ParentHash) < 66 {
		return fieldError("parentHash", "expected 66 characters")
	}

	if h.Number == 1 && len(h.ParentHash) > 0 {
		return fieldError("parentHash", "should be empty since block number is 1")
	}

	if h.Number < 1 {
		return fieldError("number", "number must be greater or equal to 1")
	}

	if len(h.CreatorPubKey) == 0 {
		return fieldError("creatorPubKey", "creator public key is required")
	}

	if _, err := crypto.PubKeyFromBase58(h.CreatorPubKey); err != nil {
		return fieldError("creatorPubKey", err.Error())
	}

	if len(h.TransactionsRoot) != 66 {
		return fieldError("transactionsRoot", "expected 66 characters")
	}

	if len(h.StateRoot) != 66 {
		return fieldError("stateRoot", "expected 66 characters")
	}

	if len(h.StateRoot) != 66 {
		return fieldError("stateRoot", "expected 66 characters")
	}

	if h.Nonce == 0 {
		return fieldError("nonce", "must not be zero")
	}

	if len(h.MixHash) != 32 {
		return fieldError("mixHash", "expected 32 characters")
	}

	if !govalidator.IsNumeric(h.Difficulty) {
		return fieldError("difficulty", "must be numeric")
	}

	if diff, _ := strconv.ParseInt(h.Difficulty, 10, 64); diff <= 0 {
		return fieldError("difficulty", "must be non-zero and non-negative")
	}

	if h.Timestamp <= 0 {
		return fieldError("timestamp", "must not be empty or a negative value")
	}

	return nil
}

// Bytes return the bytes representation of the header
func (h *Header) Bytes() []byte {
	return getBytes([]interface{}{
		h.ParentHash,
		strconv.FormatUint(h.Number, 10),
		h.TransactionsRoot,
		h.StateRoot,
		strconv.FormatUint(h.Nonce, 10),
		h.MixHash,
		h.Difficulty,
		h.Timestamp,
	})
}

// ComputeHash returns the SHA256 hash of the header as a hex string
// prefixed by '0x'
func (h *Header) ComputeHash() string {
	bs := h.Bytes()
	hash := sha256.Sum256(bs)
	return ToHex(hash[:])
}

// Bytes returns the ANS1 bytes equivalent of the block data.
// The block signature and hash are not included in this computation.
func (b *Block) Bytes() []byte {

	var txBytes [][]byte
	for _, tx := range b.Transactions {
		txBytes = append(txBytes, tx.Bytes())
	}

	return getBytes([]interface{}{
		b.Header.Bytes(),
		txBytes,
	})
}

// ComputeHash returns the SHA256 hash of the header as a hex string
// prefixed by '0x'
func (b *Block) ComputeHash() string {
	bs := b.Bytes()
	hash := sha256.Sum256(bs)
	return ToHex(hash[:])
}

// BlockSign signs a block.
// Expects private key in base58Check encoding
func BlockSign(b *Block, privKey string) ([]byte, error) {

	if b == nil {
		return nil, fmt.Errorf("nil block")
	}

	pKey, err := crypto.PrivKeyFromBase58(privKey)
	if err != nil {
		return nil, err
	}

	sig, err := pKey.Sign(b.Bytes())
	if err != nil {
		return nil, err
	}

	return sig, nil
}

// BlockVerify checks whether a block's signature is valid.
// Expect block.Header.CreatorPubKey and block.Sig to be set.
func BlockVerify(block *Block) error {

	if block == nil {
		return fmt.Errorf("nil block")
	}

	if block.Header != nil && block.Header.CreatorPubKey == "" {
		return fieldError("header.creatorPubKey", "creator public not set")
	}

	if len(block.Sig) == 0 {
		return fieldError("sig", "signature not set")
	}

	pubKey, err := crypto.PubKeyFromBase58(block.Header.CreatorPubKey)
	if err != nil {
		return fieldError("senderPubKey", err.Error())
	}

	decSig, _ := FromHex(block.Sig)
	valid, err := pubKey.Verify(block.Bytes(), decSig)
	if err != nil {
		return fieldError("sig", err.Error())
	}

	if !valid {
		return crypto.ErrBlockVerificationFailed
	}

	return nil
}

//GetGenesisDifficulty get the genesis block difficulty
func (b *Block) GetGenesisDifficulty() *big.Int {
	return big.NewInt(500000)
}

// GetNumber returns the number
func (b *Block) GetNumber() uint64 {
	return b.Header.Number
}
