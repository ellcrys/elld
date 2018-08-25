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
}

// EncodeMsgpack implements msgpack.CustomEncoder
func (h *Handshake) EncodeMsgpack(enc *msgpack.Encoder) error {
	tdStr := h.BestBlockTotalDifficulty.String()
	return enc.Encode(h.SubVersion, h.BestBlockHash, tdStr)
}

// DecodeMsgpack implements msgpack.CustomDecoder
func (h *Handshake) DecodeMsgpack(dec *msgpack.Decoder) error {
	var tdStr string
	if err := dec.Decode(&h.SubVersion, &h.BestBlockHash, &tdStr); err != nil {
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
}

// Pong represents a pong message
type Pong struct {
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

// GetBlockHeaders represents a message requesting
// for headers of blocks after the provided hash
type GetBlockHeaders struct {
	Hash      util.Hash `json:"hash" msgpack:"hash"`
	MaxBlocks int64     `json:"maxBlocks" msgpack:"maxBlocks"`
}

// BlockHeaders represents a message containing
// block headers as a response to GetBlockHeaders
type BlockHeaders struct {
	Headers []objects.Header
}
