package block

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto/sha3"
	"github.com/ethereum/go-ethereum/rlp"
	scribble "github.com/nanobox-io/golang-scribble"
)

type PartialBlock struct {
	header            int
	previousBlockHash common.Hash
}

// FullBlock Block  Construct
type FullBlock struct {
	Version        string
	HashPrevBlock  string
	TX             []string
	HashMerkleRoot string
	Time           string
	Nounce         uint64
	Difficulty     *big.Int
	Number         uint64
	PowHash        string
	PowResult      string
}

type JsonBlock struct {
	Version        string `json:"Version"`
	HashPrevBlock  string `json:"HashPrevBlock"`
	HashMerkleRoot string `json:"HashMerkleRoot"`
	Time           string `json:"Time"`
	Nounce         string `json:"Nounce"`
	Difficulty     string `json:"Difficulty"`
	Number         string `json:"Number"`
	PowHash        string `json:"PowHash"`
	PowResult      string `json:"PowResult"`
}

// HashNoNonce returns the hash which is used as input for the proof-of-work search.
func (h *FullBlock) HashNoNonce() common.Hash {
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

// GetTotalBlocks get total block in the blockchain
func GetTotalBlocks2() uint64 {
	return 5
}

// GetTotalBlocks2 get total block in the blockchain
func GetTotalBlocks() uint64 {
	db, err := scribble.New("scribleDB/", nil)
	if err != nil {
		fmt.Println("there is error in the db")
	}

	// Read all fish from the database, unmarshaling the response.
	records, _ := db.ReadAll("block")

	recordLength := len(records) - 1
	// fmt.Println("Total Block is ", len(records))
	return uint64(recordLength)
}

//get the genesis block difficulty
func GetGenesisDifficulty() *big.Int {
	return big.NewInt(500000)
}

func CheckBlock() {
	fmt.Println("This is a very good block")
}

func New() {

}
