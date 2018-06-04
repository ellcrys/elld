// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"crypto/sha256"
	"fmt"
	"math/big"
	"strconv"
	"time"

	merkletree "github.com/cbergoon/merkletree"
	"github.com/ellcrys/druid/miner"
	ellBlock "github.com/ellcrys/druid/wire"

	DB "github.com/ellcrys/druid/scribleDB"
	"github.com/spf13/cobra"
)

// minerCmd represents the miner command
var minerCmd = &cobra.Command{
	Use:   "miner",
	Short: "Mining Algorithm for proof of work",
	Long: `An Ethash proof of work Algorith based on formerly Dagger-Hashimoto algorith
	It uses Dag file to speed up mining process
	go run main.go miner to run this package.`,

	Run: func(cmd *cobra.Command, args []string) {

		miner.GetLogger().SetToDebug()
		fmt.Println("************************************************************************************************************************************************************ ")

		//get current time stamp
		currentUTCTime := time.Now().Format("20060102150405")
		currentUTCTimeUint, _ := strconv.ParseUint(currentUTCTime, 10, 64)

		selectedTransaction := MemPool(15)

		var list []merkletree.Content

		for _, tx := range selectedTransaction {
			list = append(list, Transaction{tx_hash: tx})
		}

		// Create a New Merkle tree from the list of Transaction
		t, _ := merkletree.NewTree(list)

		//Get the Merkle Root of the tree
		merkleRoot := t.MerkleRoot()
		merkleRootString := fmt.Sprintf("%x", merkleRoot)

		// currentBlockNumber to get total blocks and add 1 to it
		currentProcessingBlockNumber := DB.GetTotalBlocks() + 1

		block := ellBlock.Block{
			Version:        "1.0",
			HashMerkleRoot: merkleRootString,
			Time:           currentUTCTime,
			Number:         uint64(currentProcessingBlockNumber),
			TX:             selectedTransaction,
		}

		// create a config struct to be passed as our parameter
		config := miner.Config{
			CacheDir: "CacheFile", CachesInMem: 0, CachesOnDisk: 1, DatasetDir: "DagFile", DatasetsInMem: 0, DatasetsOnDisk: 1, PowMode: miner.ModeFake,
		}

		// Create a New Ethash Constructor
		newEllMiner := miner.New(config)

		//ID of the Miner
		minerID := 67927

		//check if block is a genesuis block
		totalBlockNumber := DB.GetTotalBlocks()

		//if the block is a genesis bloc
		if totalBlockNumber == 0 {
			block.HashPrevBlock = ""

			bd := block.GetGenesisDifficulty().String()
			block.Difficulty = bd

		} else {

			parentBlock := DB.GetSingleBlock(strconv.Itoa(int(totalBlockNumber)))
			block.HashPrevBlock = parentBlock.PowHash

			parentBlockTime, err1 := new(big.Int).SetString(parentBlock.Time, 10)
			if err1 == false {
				fmt.Println("Error converting parent blockTime to BigInt")
			}

			ParentDifficulty, err2 := new(big.Int).SetString(parentBlock.Difficulty, 10)
			if err2 == false {
				fmt.Println("Error converting ParentDifficulty to BigInt", err2, parentBlock.Difficulty)
			}

			parentBlockNumber := new(big.Int).SetUint64(parentBlock.Number)

			BlockDifficulty := newEllMiner.CalcDifficulty("Homestead", currentUTCTimeUint, parentBlockTime, ParentDifficulty, parentBlockNumber)

			// convert homestead block difficulty to string
			BlockDifficultyString := BlockDifficulty.String()
			block.Difficulty = BlockDifficultyString
		}

		// Create a runner and the multiple search threads it directs
		outputDigest, outputResult, outputNonce, err := newEllMiner.Mine(&block, minerID)

		if err != nil {
			fmt.Println("Mining was cancelled by the Miner")
		}

		if outputDigest != "" {
			block.Nounce = outputNonce
			block.PowHash = outputDigest
			block.PowResult = outputResult

			// store block into Database
			//ADD block to block chain
			DB.AddBlockToChain(strconv.Itoa(int(block.Number)), block)
			fmt.Println("Block ", block.Number, " Successfully Mined")

			fmt.Println("************************************************************************************************************************************************************ ")

			fmt.Println("+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
			fmt.Println("Verifying the Proof of work")
			errSeal := newEllMiner.VerifyPOW(&block)
			if errSeal != nil {
				fmt.Println("Invalid Proof of work verification")
			} else {
				fmt.Println("Proof of work verification Successful")
			}
			fmt.Println("+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")

		} else {
			fmt.Println("Error Mining Block ")
		}

	},
}

// MemPool generates random transactions based on max parameter
func MemPool(maxTx int) []string {

	var tx_array []string

	//rand.Seed(time.Now().UTC().UnixNano())
	for i := 1; i <= maxTx; i++ {

		//i_byte := []byte(strconv.Itoa(i))
		iByte := int(time.Now().UTC().UnixNano())
		transaction := sha256.Sum256([]byte(strconv.Itoa(iByte)))

		// hash := fmt.Printf("%x", transaction)
		hash := fmt.Sprintf("%x", transaction)

		// fmt.Println(hash)
		tx_array = append(tx_array, hash)
	}

	return tx_array
}

// Transaction implements the Content interface provided by
// merkletree and represents the content stored in the tree.
type Transaction struct {
	tx_hash string
}

//CalculateHash hashes the values of a Transaction
func (t Transaction) CalculateHash() []byte {
	h := sha256.New()
	h.Write([]byte(t.tx_hash))
	return h.Sum(nil)
}

//Equals tests for equality of two Transaction
func (t Transaction) Equals(other merkletree.Content) bool {
	return t.tx_hash == other.(Transaction).tx_hash
}

func init() {
	rootCmd.AddCommand(minerCmd)

}
