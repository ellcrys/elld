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
// type FullBlock struct {
// 	Version        string
// 	HashPrevBlock  string
// 	TX             []string
// 	HashMerkleRoot string
// 	Time           *big.Int
// 	Nounce         uint64
// 	Difficulty     *big.Int
// 	Number         uint64
// 	PowHash        string
// 	PowResult      string
// }

// FullBlock Block  Construct
type FullBlock struct {
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

//JsonBlock consume previous block details
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
func GetTotalBlocks() int {
	db, err := scribble.New("scribleDB/", nil)
	if err != nil {
		fmt.Println("there is error in the db")
	}

	// Read all fish from the database, unmarshaling the response.
	records, _ := db.ReadAll("block")

	recordLength := len(records)
	return recordLength

}

//AddBlockToChain Add blocks to the Chain
func (h *FullBlock) AddBlockToChain(blockNumber string, mapData map[string]interface{}) {

	db, err := scribble.New("scribleDB/", nil)
	if err != nil {
		fmt.Println("there is error in the db")
	}

	db.Write("block", blockNumber, mapData)
	fmt.Println(blockNumber, " Added to Chain")
}

// DeleteAllBlock should delete all block in blockchain
func DeleteAllBlock() {

	db, err := scribble.New("scribleDB/", nil)
	if err != nil {
		fmt.Println("there is error in the db")
	}

	if err := db.Delete("block", ""); err != nil {
		fmt.Println("Error", err)
	} else {
		fmt.Println("All blocks successfully Deleted")
	}
}

//GetGenesisDifficulty get the genesis block difficulty
func (h *FullBlock) GetGenesisDifficulty() *big.Int {
	return big.NewInt(500000)
}

func CheckBlockInBlockChain() {
	fmt.Println("This is a very good block")
}
