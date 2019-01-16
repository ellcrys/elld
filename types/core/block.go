package core

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/vmihailenco/msgpack"

	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/types"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/elld/util/math"
)

var mtx sync.RWMutex

// Header represents the header of a block
type Header struct {
	Number           uint64          `json:"number" msgpack:"number"`
	Nonce            util.BlockNonce `json:"nonce" msgpack:"nonce"`
	Timestamp        int64           `json:"timestamp" msgpack:"timestamp"`
	CreatorPubKey    util.String     `json:"creatorPubKey" msgpack:"creatorPubKey"`
	ParentHash       util.Hash       `json:"parentHash" msgpack:"parentHash"`
	StateRoot        util.Hash       `json:"stateRoot" msgpack:"stateRoot"`
	TransactionsRoot util.Hash       `json:"transactionsRoot" msgpack:"transactionsRoot"`
	Difficulty       *big.Int        `json:"difficulty" msgpack:"difficulty"`
	TotalDifficulty  *big.Int        `json:"totalDifficulty" msgpack:"totalDifficulty"`
	Extra            []byte          `json:"extra" msgpack:"extra"`
}

// GetTransactionsRoot gets the transaction root
func (h *Header) GetTransactionsRoot() util.Hash {
	return h.TransactionsRoot
}

// SetTransactionsRoot sets the transaction root
func (h *Header) SetTransactionsRoot(txRoot util.Hash) {
	h.TransactionsRoot = txRoot
}

// SetStateRoot sets the state root
func (h *Header) SetStateRoot(sr util.Hash) {
	h.StateRoot = sr
}

// GetStateRoot gets the state root
func (h *Header) GetStateRoot() util.Hash {
	return h.StateRoot
}

// SetDifficulty sets the difficulty
func (h *Header) SetDifficulty(diff *big.Int) {
	h.Difficulty = diff
}

// SetParentHash sets parent hash
func (h *Header) SetParentHash(hash util.Hash) {
	h.ParentHash = hash
}

// SetNumber sets the number
func (h *Header) SetNumber(n uint64) {
	h.Number = n
}

// SetNonce sets the nonce
func (h *Header) SetNonce(nonce util.BlockNonce) {
	h.Nonce = nonce
}

// SetTimestamp sets the timestamp
func (h *Header) SetTimestamp(timestamp int64) {
	h.Timestamp = timestamp
}

// GetCreatorPubKey gets the public key of the creator
func (h *Header) GetCreatorPubKey() util.String {
	return h.CreatorPubKey
}

// GetParentHash gets the parent hash
func (h *Header) GetParentHash() util.Hash {
	return h.ParentHash
}

// GetNonce gets the nonce
func (h *Header) GetNonce() util.BlockNonce {
	return h.Nonce
}

// GetDifficulty gets the difficulty
func (h *Header) GetDifficulty() *big.Int {
	return h.Difficulty
}

// GetTotalDifficulty gets the total difficulty
func (h *Header) GetTotalDifficulty() *big.Int {
	return h.TotalDifficulty
}

// SetTotalDifficulty sets the total difficulty
func (h *Header) SetTotalDifficulty(td *big.Int) {
	h.TotalDifficulty = td
}

// GetTimestamp gets the time stamp
func (h *Header) GetTimestamp() int64 {
	return h.Timestamp
}

// GetExtra gets the extra data
func (h *Header) GetExtra() []byte {
	return h.Extra
}

// GetNumber returns the header number which is the block number
func (h *Header) GetNumber() uint64 {
	return h.Number
}

// Copy creates a copy of the header
func (h *Header) Copy() types.Header {
	var newH = *h
	if newH.Difficulty = new(big.Int); h.Difficulty != nil {
		newH.Difficulty.Set(h.Difficulty)
	}
	if newH.TotalDifficulty = new(big.Int); h.TotalDifficulty != nil {
		newH.TotalDifficulty.Set(h.TotalDifficulty)
	}
	return &newH
}

// EncodeMsgpack implements msgpack.CustomEncoder
func (h *Header) EncodeMsgpack(enc *msgpack.Encoder) error {
	difficultyStr := h.Difficulty.String()
	tdStr := h.TotalDifficulty.String()
	return enc.Encode(h.Number, h.Nonce, h.Timestamp, h.CreatorPubKey,
		h.ParentHash, h.StateRoot, h.TransactionsRoot, h.Extra, difficultyStr, tdStr)
}

// GetBytes return the bytes representation of the header
func (h *Header) GetBytes() []byte {
	return getBytes([]interface{}{
		h.CreatorPubKey,
		math.SetBigInt(new(big.Int), h.Difficulty).Bytes(),
		h.Extra,
		h.Nonce,
		h.Number,
		h.ParentHash,
		h.StateRoot,
		h.Timestamp,
		math.SetBigInt(new(big.Int), h.TotalDifficulty).Bytes(),
		h.TransactionsRoot,
	})
}

// SetCreatorPubKey sets the creator's public key
func (h *Header) SetCreatorPubKey(key util.String) {
	h.CreatorPubKey = key
}

// ComputeHash returns the SHA256 hash of the header
func (h *Header) ComputeHash() util.Hash {
	bs := h.GetBytes()
	hash := util.Blake2b256(bs)
	return util.BytesToHash(hash[:])
}

// DecodeMsgpack implements msgpack.CustomDecoder
func (h *Header) DecodeMsgpack(dec *msgpack.Decoder) error {
	var difficultyStr, tdStr string
	if err := dec.Decode(&h.Number, &h.Nonce, &h.Timestamp, &h.CreatorPubKey,
		&h.ParentHash, &h.StateRoot, &h.TransactionsRoot, &h.Extra, &difficultyStr, &tdStr); err != nil {
		return err
	}
	h.Difficulty, _ = new(big.Int).SetString(difficultyStr, 10)
	h.TotalDifficulty, _ = new(big.Int).SetString(tdStr, 10)
	return nil
}

// Block represents a block
type Block struct {
	Header       *Header        `json:"header" msgpack:"header"`
	Transactions []*Transaction `json:"transactions" msgpack:"transactions"`
	Hash         util.Hash      `json:"hash" msgpack:"hash"`
	Sig          []byte         `json:"sig" msgpack:"sig"`

	// Broadcaster is the peer responsible
	// for sending this block.
	Broadcaster Engine `json:"-" msgpack:"-"`

	// types.ValidationContext can be used to alter
	// the way the block is validated
	ValidationContexts []types.ValidationContext `json:"-" msgpack:"-"`
}

// SetBroadcaster sets the originator
func (b *Block) SetBroadcaster(o Engine) {
	mtx.Lock()
	defer mtx.Unlock()
	b.Broadcaster = o
}

// GetBroadcaster gets the originator
func (b *Block) GetBroadcaster() Engine {
	mtx.RLock()
	defer mtx.RUnlock()
	return b.Broadcaster
}

// GetHash returns the block's hash
func (b *Block) GetHash() util.Hash {
	return b.Hash
}

// GetHashAsHex returns the block's hex equivalent of its hash
// preceded by 0x
func (b *Block) GetHashAsHex() string {
	return b.GetHash().HexStr()
}

// EncodeMsgpack implements msgpack.CustomEncoder
func (b *Block) EncodeMsgpack(enc *msgpack.Encoder) error {
	return enc.Encode(b.Hash, b.Sig, b.Header, b.Transactions)
}

// DecodeMsgpack implements msgpack.CustomDecoder
func (b *Block) DecodeMsgpack(dec *msgpack.Decoder) error {
	return dec.Decode(&b.Hash, &b.Sig, &b.Header, &b.Transactions)
}

// GetHashNoNonce gets the hash of the header
// without the nonce included in the computation
func (h *Header) GetHashNoNonce() util.Hash {
	result := getBytes([]interface{}{
		h.CreatorPubKey,
		math.SetBigInt(new(big.Int), h.Difficulty).Bytes(),
		h.Extra,
		h.Number,
		h.ParentHash,
		h.StateRoot,
		h.Timestamp,
		math.SetBigInt(new(big.Int), h.TotalDifficulty).Bytes(),
		h.TransactionsRoot,
	})
	return util.BytesToHash(util.Blake2b256(result[:]))
}

// GetTransactions gets the transactions
func (b *Block) GetTransactions() (txs []types.Transaction) {
	for _, tx := range b.Transactions {
		txs = append(txs, tx)
	}
	return
}

// GetValidationContexts gets the validation contexts
func (b *Block) GetValidationContexts() []types.ValidationContext {
	return b.ValidationContexts
}

// SetValidationContexts sets the validation contexts
func (b *Block) SetValidationContexts(ctxs ...types.ValidationContext) {
	b.ValidationContexts = ctxs
}

// GetHeader gets the block's header
func (b *Block) GetHeader() types.Header { return b.Header }

// SetHeader sets the block header
func (b *Block) SetHeader(h types.Header) {
	if h == nil {
		b.Header = nil
		return
	}
	b.Header = h.(*Header)
}

// ComputeHash returns the SHA256 hash of the header as a hex string
// prefixed by '0x'
func (b *Block) ComputeHash() util.Hash {
	bs := b.GetBytesNoHashSig()
	hash := util.Blake2b256(bs)
	return util.BytesToHash(hash[:])
}

// GetSignature gets the signature
func (b *Block) GetSignature() []byte {
	return b.Sig
}

// SetHash sets the hash
func (b *Block) SetHash(h util.Hash) {
	b.Hash = h
}

// SetSignature sets the signature
func (b *Block) SetSignature(sig []byte) {
	b.Sig = sig
}

// ReplaceHeader creates a copy of h and
// sets as the blocks header
func (b *Block) ReplaceHeader(h types.Header) types.Block {
	cpy := h.Copy()
	return &Block{
		Header:       cpy.(*Header),
		Transactions: b.Transactions,
	}
}

// SetSig sets the signature
func (b *Block) SetSig(sig []byte) {
	b.Sig = sig
}

// GetBytes gets the bytes representation
// of the block.
func (b *Block) GetBytes() []byte {

	txsBytes := []byte{}
	for _, tx := range b.Transactions {
		txsBytes = append(txsBytes, tx.Bytes()...)
	}

	data := []interface{}{
		b.Hash,
		b.Header.GetBytes(),
		b.Sig,
		txsBytes,
	}

	return getBytes(data)
}

// GetBytesNoTxs gets the bytes representation
// of the block without the adding the
// transactions' bytes
func (b *Block) GetBytesNoTxs() []byte {

	data := []interface{}{
		b.Hash,
		b.Header.GetBytes(),
		b.Sig,
	}

	return getBytes(data)
}

// GetSize gets the byte size
func (b *Block) GetSize() int64 {
	return int64(len(b.GetBytes()))
}

// GetSizeNoTxs gets the byte size
func (b *Block) GetSizeNoTxs() int64 {
	return int64(len(b.GetBytesNoTxs()))
}

// GetBytesNoHashSig gets the bytes representation
// of the block without the hash and signature
// included.
func (b *Block) GetBytesNoHashSig() []byte {

	txsBytes := []byte{}
	for _, tx := range b.Transactions {
		txsBytes = append(txsBytes, tx.Bytes()...)
	}

	data := []interface{}{
		b.Header.GetBytes(),
		txsBytes,
	}

	return getBytes(data)
}

// BlockSign signs a block.
// Expects private key in base58Check encoding
func BlockSign(b types.Block, privKey string) ([]byte, error) {

	if b == nil {
		return nil, fmt.Errorf("nil block")
	}

	pKey, err := crypto.PrivKeyFromBase58(privKey)
	if err != nil {
		return nil, err
	}

	sig, err := pKey.Sign(b.GetBytesNoHashSig())
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

	if block.Header != nil && block.Header.GetCreatorPubKey() == "" {
		return fieldError("header.creatorPubKey", "creator public not set")
	}

	if len(block.Sig) == 0 {
		return fieldError("sig", "signature not set")
	}

	pubKey, err := crypto.PubKeyFromBase58(block.Header.GetCreatorPubKey().String())
	if err != nil {
		return fieldError("header.creatorPubKey", err.Error())
	}

	valid, err := pubKey.Verify(block.GetBytesNoHashSig(), block.Sig)
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
	return b.Header.GetNumber()
}
