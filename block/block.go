package block

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto/sha3"
	"github.com/ethereum/go-ethereum/rlp"
)

//Block  Construct
type Block struct {
	Version        string
	HashPrevBlock  string
	TX             []string
	HashMerkleRoot string
	Time           string
	Nounce         uint64
	Difficulty     string
	Number         uint64
	PowHash        string
	PowResult      string
}

// HashNoNonce returns the hash which is used as input for the proof-of-work search.
func (h *Block) HashNoNonce() common.Hash {
	return rlpHash([]interface{}{
		h.HashPrevBlock,
		h.HashMerkleRoot,
		h.Difficulty,
		// h.Number,
		h.Time,
		// h.Extra,
	})
}

func rlpHash(x interface{}) (h common.Hash) {
	hw := sha3.NewKeccak256()
	rlp.Encode(hw, x)
	hw.Sum(h[:0])
	return h
}

//GetGenesisDifficulty get the genesis block difficulty
func (h *Block) GetGenesisDifficulty() *big.Int {
	return big.NewInt(500000)
}
