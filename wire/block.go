package wire

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/util"
)

// A BlockNonce is a 64-bit hash which proves (combined with the
// mix-hash) that a sufficient amount of computation has been carried
// out on a block.
type BlockNonce [8]byte

// EmptyBlockNonce is a BlockNonce with no values
var EmptyBlockNonce = BlockNonce([8]byte{})

// EncodeNonce converts the given integer to a block nonce.
func EncodeNonce(i uint64) BlockNonce {
	var n BlockNonce
	binary.BigEndian.PutUint64(n[:], i)
	return n
}

// Uint64 returns the integer value of a block nonce.
func (n BlockNonce) Uint64() uint64 {
	return binary.BigEndian.Uint64(n[:])
}

// MarshalText encodes n as a hex string with 0x prefix.
func (n BlockNonce) MarshalText() string {
	return util.ToHex(n[:])
}

// Block represents a block
type Block struct {
	Header       *Header        `json:"header" msgpack:"header"`
	Transactions []*Transaction `json:"transactions" msgpack:"transactions"`
	Hash         util.Hash      `json:"hash" msgpack:"hash"`
	Sig          []byte         `json:"sig" msgpack:"sig"`
}

// Header represents the header of a block
type Header struct {
	Number           uint64     `json:"number" msgpack:"number"`
	Nonce            BlockNonce `json:"nonce" msgpack:"nonce"`
	MixHash          util.Hash  `json:"mixHash" msgpack:"mixHash"`
	Timestamp        int64      `json:"timestamp" msgpack:"timestamp"`
	CreatorPubKey    string     `json:"creatorPubKey" msgpack:"creatorPubKey"`
	ParentHash       util.Hash  `json:"ParentHash" msgpack:"ParentHash"`
	StateRoot        util.Hash  `json:"stateRoot" msgpack:"stateRoot"`
	TransactionsRoot util.Hash  `json:"transactionsRoot" msgpack:"transactionsRoot"`
	Difficulty       *big.Int   `json:"difficulty" msgpack:"difficulty"`
	Extra            []byte     `json:"extra" msgpack:"extra"`
}

// GetNumber returns the header number which is the block number
func (h *Header) GetNumber() uint64 {
	return h.Number
}

// GetHash returns the block's hash
func (b *Block) GetHash() util.Hash {
	return b.Hash
}

// HashToHex returns the block's hex equivalent of its hash
// preceded by 0x
func (b *Block) HashToHex() string {
	return b.GetHash().HexStr()
}

// BlockFromString unmarshal a json string into a Block
func BlockFromString(str string) (*Block, error) {
	var block Block
	var err error
	err = json.Unmarshal([]byte(str), &block)
	return &block, err
}

// HashNoNonce returns the hash which is used as input for the proof-of-work search.
func (h *Header) HashNoNonce() util.Hash {
	result := getBytes([]interface{}{
		h.ParentHash,
		h.Number,
		h.CreatorPubKey,
		h.TransactionsRoot,
		h.StateRoot,
		h.Difficulty,
		h.Timestamp,
		h.Extra,
	})
	return sha256.Sum256(result)
}

// Bytes return the bytes representation of the header
func (h *Header) Bytes() []byte {
	return getBytes([]interface{}{
		h.ParentHash,
		h.Number,
		h.CreatorPubKey,
		h.TransactionsRoot,
		h.StateRoot,
		h.MixHash,
		h.Difficulty,
		h.Timestamp,
		h.Nonce,
		h.Extra,
	})
}

// ComputeHash returns the SHA256 hash of the header
func (h *Header) ComputeHash() util.Hash {
	bs := h.Bytes()
	hash := sha256.Sum256(bs)
	return util.BytesToHash(hash[:])
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

// GetHeader gets the block's header
func (b *Block) GetHeader() *Header { return b.Header }

// SetHeader sets the block header
func (b *Block) SetHeader(h *Header) { b.Header = h }

// ComputeHash returns the SHA256 hash of the header as a hex string
// prefixed by '0x'
func (b *Block) ComputeHash() util.Hash {
	bs := b.Bytes()
	hash := sha256.Sum256(bs)
	return util.BytesToHash(hash[:])
}

// CopyHeader creates a copy of a block header
func CopyHeader(h *Header) *Header {
	newH := *h
	if newH.Difficulty = new(big.Int); h.Difficulty != nil {
		newH.Difficulty.Set(h.Difficulty)
	}
	return &newH
}

// WithSeal returns a new block with the data from b but the header replaced with
// the sealed one.
func (b *Block) WithSeal(header *Header) *Block {
	cpy := *header

	return &Block{
		Header:       &cpy,
		Transactions: b.Transactions,
	}
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
		return fieldError("header.creatorPubKey", err.Error())
	}

	valid, err := pubKey.Verify(block.Bytes(), block.Sig)
	if err != nil {
		return fieldError("sig", err.Error())
	}

	if !valid {
		return crypto.ErrBlockVerificationFailed
	}

	return nil
}

// GetNumber returns the number
func (b *Block) GetNumber() uint64 {
	return b.Header.Number
}
