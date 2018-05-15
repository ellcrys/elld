package db

import (
	"fmt"

	scribble "github.com/nanobox-io/golang-scribble"
)

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
func AddBlockToChain(blockNumber string, mapData map[string]interface{}) {

	db, err := scribble.New("scribleDB/", nil)
	if err != nil {
		fmt.Println("there is error in the db")
	}

	db.Write("block", blockNumber, mapData)
}

// DeleteAllBlock should delete all block in blockchain
func DeleteAllBlock() {

	db, err := scribble.New("scribleDB/", nil)
	if err != nil {
		fmt.Println("there is error in the db")
	}

	if err := db.Delete("block", ""); err != nil {
		fmt.Println("Error", err)
	}
}
