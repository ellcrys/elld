package miner_test

import (
	"math/big"
	"testing"

	miner "github.com/ellcrys/druid/miner"
	DB "github.com/ellcrys/druid/scribleDB"
	"github.com/ellcrys/druid/wire"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestBlock(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Block Suite")
}

var _ = Describe("Miner", func() {

	//Test For CalcDifficulty
	var _ = Describe(".CalcDifficulty", func() {

		currentTime := uint64(20180515105500)
		parentTime := big.NewInt(20180514105500)
		parentDifficulty := big.NewInt(31083)
		parentBlockNumber := big.NewInt(30)
		blockDifficulty := miner.CalcDifficulty("Homestead", currentTime, parentTime, parentDifficulty, parentBlockNumber)

		Context("calculate Difficulty for next block", func() {
			It("It must not be 0", func() {
				Expect(blockDifficulty).ShouldNot(BeZero())
			})

			It("It must not be nil", func() {
				Expect(blockDifficulty).ShouldNot(BeNil())
			})
		})

	})

	//Test For Mine - PROOF OF WORK
	var _ = Describe(".Mine", func() {

		ellBlock := wire.Block{
			Difficulty:     "70212",
			HashMerkleRoot: "8c4a16df5c399bb06a8752ff16f776901a714dfa3f4113be2c14be7c136ef582",
			HashPrevBlock:  "efae6c6522095e57bf885756af7ccc38483e6582b83a80adf588126f03134b78",
			Nounce:         1525873090716071984,
			Number:         2,
			PowHash:        "508f043b157f888e361b04e891423651c2baa87d70b03a5354bed24b2d3125bd",
			PowResult:      "000157eb1e61cc522aab7211e5afa77bd7d30b197827b63714ffbb543eeea833",
			Time:           "2018-05-05 14:45:45",
			Version:        "1.0",
		}

		minerID := 63548

		config := miner.Config{
			CacheDir: "../CacheFile", CachesInMem: 0, CachesOnDisk: 1, DatasetDir: "../DagFile", DatasetsInMem: 0, DatasetsOnDisk: 1, PowMode: miner.ModeFake,
		}

		// Create a New Ethash Constructor
		newEllMiner := miner.New(config)
		digest, result, nounce := newEllMiner.Mine(&ellBlock, minerID)

		Context("calculate POW for a block", func() {

			// Describe the digest from a Mine Function
			var _ = Describe("digest", func() {

				It("It must not be 0", func() {
					Expect(digest).ShouldNot(BeZero())
				})

				It("It must not be nil", func() {
					Expect(digest).ShouldNot(BeNil())
				})
				It("It must not be empty", func() {
					Expect(digest).ShouldNot(BeEmpty())
				})

				It("It must be 64 character in length", func() {
					Expect(digest).To(HaveLen(64))
				})

			})

			// Describe the result from a Mine Function
			var _ = Describe("result", func() {

				It("It must not be 0", func() {
					Expect(result).ShouldNot(BeZero())
				})

				It("It must not be nil", func() {
					Expect(result).ShouldNot(BeNil())
				})
				It("It must not be empty", func() {
					Expect(result).ShouldNot(BeEmpty())
				})

				It("It must be 64 character in length", func() {
					Expect(result).To(HaveLen(64))
				})

			})

			// Describe the nounce from a Mine Function
			var _ = Describe("nounce", func() {

				It("It must not be 0", func() {
					Expect(nounce).ShouldNot(BeZero())
				})

				It("It must not be nil", func() {
					Expect(nounce).ShouldNot(BeNil())
				})

				It("It must be 64 character in length", func() {
					Expect(digest).NotTo(HaveLen(0))
				})

			})

		})

	})

	var _ = Describe("VerifyPOW", func() {

		// config struct to innitialize the miner package
		config := miner.Config{
			CacheDir: "CacheFile", CachesInMem: 0, CachesOnDisk: 1, DatasetDir: "DagFile", DatasetsInMem: 0, DatasetsOnDisk: 1, PowMode: miner.ModeFake,
		}

		//Create a New Ethash Constructor
		newEllMiner := miner.New(config)

		ellBlock := DB.GetSingleBlock("20")
		ellBlock.Difficulty = "30678"

		errPow := newEllMiner.VerifyPOW(&ellBlock)

		It("It must not be 0", func() {
			Expect(errPow).ShouldNot(BeZero())
		})

	})

})
