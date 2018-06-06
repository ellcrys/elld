package wire

import (
	"crypto/sha256"
	"encoding/asn1"
	"math/big"
	"strconv"

	"github.com/ellcrys/elld/util"
)

type asn1BlockNoNonce struct {
	HashPrevBlock  string `asn1:"utf8" json:"hashPrevBlock"`
	HashMerkleRoot string `asn1:"utf8" json:"hashMerkleRoot"`
	Difficulty     string `asn1:"utf8" json:"difficulty"`
	Number         string `json:"number"`
	Timestamp      string `json:"timestamp"`
}

// HashNoNonce returns the hash which is used as input for the proof-of-work search.
func (h *Header) HashNoNonce() util.Hash {

	asn1Data := []interface{}{
		h.ParentHash,
		strconv.FormatUint(h.Number, 10),
		h.Root,
		h.Difficulty,
		h.Timestamp,
		h.Root,
	}

	result, err := asn1.Marshal(asn1Data)
	if err != nil {
		panic(err)
	}

	return sha256.Sum256(result)
}

//GetGenesisDifficulty get the genesis block difficulty
func (b *Block) GetGenesisDifficulty() *big.Int {
	return big.NewInt(500000)
}
