package wire

import (
	"math/big"

	"github.com/ellcrys/elld/types/core/objects"

	"github.com/ellcrys/elld/util"
	"github.com/vmihailenco/msgpack"
)

// Handshake represents the first message between peers
type Handshake struct {
	SubVersion               string    `json:"subversion" msgpack:"subversion"`
	BestBlockHash            util.Hash `json:"bestBlockHash" msgpack:"bestBlockHash"`
	BestBlockTotalDifficulty *big.Int  `json:"bestBlockTD" msgpack:"bestBlockTD"`
	BestBlockNumber          uint64    `json:"bestBlockNumber" msgpack:"bestBlockNumber"`
}

// EncodeMsgpack implements msgpack.CustomEncoder
func (h *Handshake) EncodeMsgpack(enc *msgpack.Encoder) error {
	tdStr := h.BestBlockTotalDifficulty.String()
	return enc.Encode(h.SubVersion, h.BestBlockHash, h.BestBlockNumber, tdStr)
}

// DecodeMsgpack implements msgpack.CustomDecoder
func (h *Handshake) DecodeMsgpack(dec *msgpack.Decoder) error {
	var tdStr string
	if err := dec.Decode(&h.SubVersion, &h.BestBlockHash, &h.BestBlockNumber, &tdStr); err != nil {
		return err
	}
	h.BestBlockTotalDifficulty, _ = new(big.Int).SetString(tdStr, 10)
	return nil
}

// GetAddr is used to request for peer addresses from other peers
type GetAddr struct {
}

// Addr is used to send peer addresses in response to a GetAddr
type Addr struct {
	Addresses []*Address `json:"addresses" msgpack:"addresses"`
}

// Address represents a peer address
type Address struct {
	Address   string `json:"address" msgpack:"address"`
	Timestamp int64  `json:"timestamp" msgpack:"timestamp"`
}

// Ping represents a ping message
type Ping struct {
	BestBlockHash            util.Hash `json:"bestBlockHash" msgpack:"bestBlockHash"`
	BestBlockTotalDifficulty *big.Int  `json:"bestBlockTD" msgpack:"bestBlockTD"`
	BestBlockNumber          uint64    `json:"bestBlockNumber" msgpack:"bestBlockNumber"`
}

// EncodeMsgpack implements msgpack.CustomEncoder
func (p *Ping) EncodeMsgpack(enc *msgpack.Encoder) error {
	tdStr := p.BestBlockTotalDifficulty.String()
	return enc.Encode(p.BestBlockHash, p.BestBlockNumber, tdStr)
}

// DecodeMsgpack implements msgpack.CustomDecoder
func (p *Ping) DecodeMsgpack(dec *msgpack.Decoder) error {
	var tdStr string
	if err := dec.Decode(&p.BestBlockHash, &p.BestBlockNumber, &tdStr); err != nil {
		return err
	}
	p.BestBlockTotalDifficulty, _ = new(big.Int).SetString(tdStr, 10)
	return nil
}

// Pong represents a pong message
type Pong struct {
	BestBlockHash            util.Hash `json:"bestBlockHash" msgpack:"bestBlockHash"`
	BestBlockTotalDifficulty *big.Int  `json:"bestBlockTD" msgpack:"bestBlockTD"`
	BestBlockNumber          uint64    `json:"bestBlockNumber" msgpack:"bestBlockNumber"`
}

// EncodeMsgpack implements msgpack.CustomEncoder
func (p *Pong) EncodeMsgpack(enc *msgpack.Encoder) error {
	tdStr := p.BestBlockTotalDifficulty.String()
	return enc.Encode(p.BestBlockHash, p.BestBlockNumber, tdStr)
}

// DecodeMsgpack implements msgpack.CustomDecoder
func (p *Pong) DecodeMsgpack(dec *msgpack.Decoder) error {
	var tdStr string
	if err := dec.Decode(&p.BestBlockHash, &p.BestBlockNumber, &tdStr); err != nil {
		return err
	}
	p.BestBlockTotalDifficulty, _ = new(big.Int).SetString(tdStr, 10)
	return nil
}

// Reject defines information about a rejected action
type Reject struct {
	Message   string `json:"message" msgpack:"message"`
	Code      int32  `json:"code" msgpack:"code"`
	Reason    string `json:"reason" msgpack:"reason"`
	ExtraData []byte `json:"extraData" msgpack:"extraData"`
}

// RequestBlock represents a message requesting for a block
type RequestBlock struct {
	Hash   string `json:"hash" msgpack:"hash"`
	Number uint64 `json:"number" msgpack:"number"`
}

// GetBlockHashes represents a message requesting
// for headers of blocks after the provided hash
type GetBlockHashes struct {
	Hash      util.Hash `json:"hash" msgpack:"hash"`
	MaxBlocks int64     `json:"maxBlocks" msgpack:"maxBlocks"`
}

// BlockHashes represents a message containing
// block hashes as a response to GetBlockHeaders
type BlockHashes struct {
	Hashes []util.Hash
}

// BlockBody represents the body of a block
type BlockBody struct {
	Header       *objects.Header        `json:"header" msgpack:"header"`
	Transactions []*objects.Transaction `json:"transactions" msgpack:"transactions"`
	Hash         util.Hash              `json:"hash" msgpack:"hash"`
	Sig          []byte                 `json:"sig" msgpack:"sig"`
}

// BlockBodies represents a collection of block bodies
type BlockBodies struct {
	Blocks []*BlockBody
}

// GetBlockBodies represents a message to fetch block bodies
// belonging to the given hashes
type GetBlockBodies struct {
	Hashes []util.Hash
}
