package wire

import (
	"crypto/sha256"
	"encoding/asn1"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

//Block  Construct
// type Block struct {
// 	Version        string
// 	HashPrevBlock  string
// 	TX             []string
// 	HashMerkleRoot string
// 	Time           string
// 	Nounce         uint64
// 	Difficulty     string
// 	Number         uint64
// 	PowHash        string
// 	PowResult      string
// }

type asn1BlockNoNonce struct {
	HashPrevBlock  string `asn1:"utf8" json:"hashPrevBlock"`
	HashMerkleRoot string `asn1:"utf8" json:"hashMerkleRoot"`
	Difficulty     string `asn1:"utf8" json:"difficulty"`
	Number         int64  `json:"number"`
	Time           string `json:"string"`
}

// HashNoNonce returns the hash which is used as input for the proof-of-work search.
func (h *Block) HashNoNonce() common.Hash {

	asn1B := asn1BlockNoNonce{
		HashPrevBlock:  h.HashPrevBlock,
		HashMerkleRoot: h.HashMerkleRoot,
		Difficulty:     h.Difficulty,
		// Number:         h.Number,
		Time: h.Time,
	}

	result, err := asn1.Marshal(asn1B)
	if err != nil {
		panic(err)
	}

	return sha256.Sum256(result)
}

//GetGenesisDifficulty get the genesis block difficulty
func (h *Block) GetGenesisDifficulty() *big.Int {
	return big.NewInt(500000)
}
