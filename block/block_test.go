package block_test

import (
	"math/big"
	"strconv"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	block "github.com/ellcrys/druid/block"
)

func TestBlock(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Block Suite")
}

var _ = Describe("Block", func() {

	// GetTotalBlocks Test
	var _ = Describe(".GetTotalBlocks", func() {

		//ellBlock := block.FullBlock{}

		//Delete all existing block in the Chain
		BeforeEach(func() {
			block.DeleteAllBlock()
		})

		//When the blockchain is empty
		Context("When blockchain has 0 blocks", func() {
			//test total block in block chain must be 0
			It("Must have 0 block in block chain", func() {
				Expect(block.GetTotalBlocks()).Should(BeZero())
			})

			// test total block should not be nil
			It("Must have 0 block in block chain", func() {
				Expect(block.GetTotalBlocks()).ShouldNot(BeNil())
			})
		})

		Context("When a block is being added", func() {

			ellBlock := block.FullBlock{
				Difficulty:     "30212",
				HashMerkleRoot: "8c4a16df5c399bb06a8752ff16f776901a714dfa3f4113be2c14be7c136ef582",
				HashPrevBlock:  "efae6c6522095e57bf885756af7ccc38483e6582b83a80adf588126f03134b78",
				Nounce:         1525873090716071984,
				Number:         2,
				PowHash:        "508f043b157f888e361b04e891423651c2baa87d70b03a5354bed24b2d3125bd",
				PowResult:      "000157eb1e61cc522aab7211e5afa77bd7d30b197827b63714ffbb543eeea833",
				Time:           "2018-05-05 14:45:45",
				Version:        "1.0",
			}

			BeforeEach(func() {

				mapD := map[string]interface{}{"Number": strconv.Itoa(int(ellBlock.Number)), "Version": ellBlock.Version, "HashPrevBlock": ellBlock.HashPrevBlock, "HashMerkleRoot": ellBlock.HashMerkleRoot, "Time": ellBlock.Time, "Nounce": strconv.Itoa(int(ellBlock.Nounce))}
				//ADD block to block chain
				ellBlock.AddBlockToChain(strconv.Itoa(int(ellBlock.Number)), mapD)
			})

			It("Must have 1 block in blockchain", func() {
				Expect(block.GetTotalBlocks()).ShouldNot(BeZero())
			})

			It("Must have 1 block in blockchain", func() {
				Expect(block.GetTotalBlocks()).ShouldNot(BeNil())
			})

			It("Must have 1 block in blockchain", func() {
				Expect(block.GetTotalBlocks()).To(Equal(1))
			})

		})

	})

	// GetTotalBlocks Test
	var _ = Describe(".GetGenesisDifficulty", func() {

		ellBlock := block.FullBlock{}

		//Delete all existing block in the Chain
		BeforeEach(func() {
			block.DeleteAllBlock()
		})

		//When the blockchain is empty
		Context("When blockchain has 0 blocks", func() {

			//test Genesis block difficulty should not be Nil
			It("Genesis Block Diificulty which must not be Nil ", func() {
				Expect(ellBlock.GetGenesisDifficulty()).ShouldNot(BeNil())
			})

			//test Genesis block difficulty must not be zero
			It("Genesis Block Diificulty should not be 0", func() {
				Expect(ellBlock.GetGenesisDifficulty()).ShouldNot(BeZero())
			})

			//test Genesis block difficulty to equal to 500000
			It("Genesis Block Dificulty of 500000", func() {
				Expect(ellBlock.GetGenesisDifficulty()).To(Equal(big.NewInt(500000)))
			})

		})

	})

	//HashNoNonce for Hash of Proof of work
	var _ = Describe(".HashNoNonce", func() {

		Context("When a block proof of work is being calculated", func() {

			ellBlock := block.FullBlock{
				Difficulty:     "30212",
				HashMerkleRoot: "8c4a16df5c399bb06a8752ff16f776901a714dfa3f4113be2c14be7c136ef582",
				HashPrevBlock:  "efae6c6522095e57bf885756af7ccc38483e6582b83a80adf588126f03134b78",
				Nounce:         1525873090716071984,
				Number:         2,
				PowHash:        "508f043b157f888e361b04e891423651c2baa87d70b03a5354bed24b2d3125bd",
				PowResult:      "000157eb1e61cc522aab7211e5afa77bd7d30b197827b63714ffbb543eeea833",
				Time:           "2018-05-05 14:45:45",
				Version:        "1.0",
			}

			//Test HashNoNonce for Proof of work  should not be Nil
			It("Genesis Block Proof of work must not be Nil ", func() {
				Expect(ellBlock.HashNoNonce()).ShouldNot(BeNil())
			})

			//Test HashNoNonce for Proof of work  must not be zero
			It("Genesis Block Proof of work not be 0", func() {
				Expect(ellBlock.HashNoNonce()).ShouldNot(BeZero())
			})

		})

	})

})
