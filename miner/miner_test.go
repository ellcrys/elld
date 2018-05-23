package miner

import (
	"errors"
	"fmt"
	"math/big"
	"path/filepath"
	"reflect"
	"testing"

	//miner "github.com/ellcrys/druid/miner"
	//"github.com/ellcrys/druid/miner"
	// "github.com/ellcrys/druid/miner"
	// "github.com/ellcrys/druid/miner"

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
		// blockDifficulty := CalcDifficulty("Homestead", currentTime, parentTime, parentDifficulty, parentBlockNumber)
		blockDifficulty := CalcDifficulty("Homestead", currentTime, parentTime, parentDifficulty, parentBlockNumber)
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

		config := Config{
			CacheDir: "../CacheFile", CachesInMem: 0, CachesOnDisk: 1, DatasetDir: "../DagFile", DatasetsInMem: 0, DatasetsOnDisk: 1, PowMode: ModeFake,
		}

		// Create a New Ethash Constructor
		newEllMiner := New(config)
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

	// send arbitrary parameters to New Function
	var _ = Describe(".New", func() {

		Context("When Empty Config struct is supplied as New Parameter", func() {

			// config := Config{
			// 	CacheDir: "../CacheFile", CachesInMem: 0, CachesOnDisk: 1, DatasetDir: "../DagFile", DatasetsInMem: 0, DatasetsOnDisk: 1, PowMode: ModeFake,
			// }
			config := Config{}
			newEllMiner := New(config)
			It("Must return Type of Etash Struct", func() {
				Expect(newEllMiner).Should(BeIdenticalTo(newEllMiner))
			})

			// It("Must Not be Empty", func() {
			// 	Expect(newEllMiner).ShouldNot(BeEmpty())
			// })

			It("Must Not be Nil", func() {
				Expect(newEllMiner).ShouldNot(BeNil())
			})

			It("Must Not be 0", func() {
				Expect(newEllMiner).ShouldNot(BeZero())
			})

		})
	})

	var _ = Describe(".memoryMap", func() {

		seed := seedHash(43*epochLength + 1)
		var endian string
		if !isLittleEndian() {
			endian = ".be"
		}

		algorithmRevision = 23

		path := filepath.Join("", fmt.Sprintf("cache-R%d-%x%s", algorithmRevision, seed[:8], endian))

		file, _, _, err := memoryMap(path)

		It("Should not be an Error", func() {
			Expect(err).ShouldNot(BeIdenticalTo(error.Error))
		})

		It("Should be Nil", func() {
			Expect(err).ShouldNot(BeNil())
		})

		_, _, err1 := memoryMapFile(file, true)

		It("Should not be an Error", func() {
			Expect(err1).ShouldNot(BeIdenticalTo(error.Error))
		})

		It("Should be Nil", func() {
			Expect(err1).ShouldNot(BeNil())
		})
	})

	var _ = Describe(".newCache", func() {
		cache := newCache(58697)

		It("Must Not Be Nil", func() {
			Expect(cache).ShouldNot(BeNil())
		})

		It("Must Not Be 0", func() {
			Expect(cache).ShouldNot(BeZero())
		})
	})

	var _ = Describe(".NewTester", func() {

		val := NewTester()

		It("Must Not Be 0", func() {
			Expect(val).ShouldNot(BeZero())
		})

		It("Must Not Be Nil", func() {
			Expect(val).ShouldNot(BeNil())
		})

		It("Must Be of type Ethash", func() {

			config := Config{}
			newEllMiner := New(config)
			Expect(reflect.TypeOf(val)).Should(Equal(reflect.TypeOf(newEllMiner)))
		})

	})

	var _ = Describe(".NewFaker", func() {

		val := NewFaker()

		It("Must Not Be 0", func() {
			Expect(val).ShouldNot(BeZero())
		})

		It("Must Not Be Nil", func() {
			Expect(val).ShouldNot(BeNil())
		})

		It("Must Be of type Ethash", func() {

			config := Config{}
			newEllMiner := New(config)
			Expect(reflect.TypeOf(val)).Should(Equal(reflect.TypeOf(newEllMiner)))
		})

	})

	var _ = Describe(".NewFakeFailer", func() {

		val := NewFakeFailer(88534)

		It("Must Not Be 0", func() {
			Expect(val).ShouldNot(BeZero())
		})

		It("Must Not Be Nil", func() {
			Expect(val).ShouldNot(BeNil())
		})

		It("Must Be of type Ethash", func() {

			config := Config{}
			newEllMiner := New(config)
			Expect(reflect.TypeOf(val)).Should(Equal(reflect.TypeOf(newEllMiner)))
		})

	})

	var _ = Describe(".NewFullFaker", func() {

		val := NewFullFaker()

		It("Must Not Be 0", func() {
			Expect(val).ShouldNot(BeZero())
		})

		It("Must Not Be Nil", func() {
			Expect(val).ShouldNot(BeNil())
		})

		It("Must Be of type Ethash", func() {

			config := Config{}
			newEllMiner := New(config)
			Expect(reflect.TypeOf(val)).Should(Equal(reflect.TypeOf(newEllMiner)))
		})

	})

	var _ = Describe(".NewShared", func() {

		val := NewShared()

		It("Must Not Be 0", func() {
			Expect(val).ShouldNot(BeZero())
		})

		It("Must Not Be Nil", func() {
			Expect(val).ShouldNot(BeNil())
		})

		It("Must Be of type Ethash", func() {

			config := Config{}
			newEllMiner := New(config)
			Expect(reflect.TypeOf(val)).Should(Equal(reflect.TypeOf(newEllMiner)))
		})

	})

	var _ = Describe(".NewFakeDelayer", func() {

		val := NewFakeDelayer(004046)

		It("Must Not Be 0", func() {
			Expect(val).ShouldNot(BeZero())
		})

		It("Must Not Be Nil", func() {
			Expect(val).ShouldNot(BeNil())
		})

		It("Must Be of type Ethash", func() {

			config := Config{}
			newEllMiner := New(config)
			Expect(reflect.TypeOf(val)).Should(Equal(reflect.TypeOf(newEllMiner)))
		})

	})

	var _ = Describe(".cache", func() {

		config := Config{}
		newEllMiner := New(config)

		val := newEllMiner.cache(30)

		It("Must Not Be 0", func() {
			Expect(val).ShouldNot(BeZero())
		})

		It("Must Not Be Nil", func() {
			Expect(val).ShouldNot(BeNil())
		})

		// It("Must Be of type Ethash", func() {
		// 	cacheTypeOf := reflect.TypeOf(cache{})
		// 	Expect(reflect.TypeOf(val)).Should(BeIdenticalTo(reflect.TypeOf(cacheTypeOf)))
		// })

	})

	//This will generate a new dag file wile running the test
	// var _ = Describe(".dataset", func() {

	// 	config := Config{}
	// 	newEllMiner := New(config)

	// 	val := newEllMiner.dataset(30)

	// 	It("Must Not Be 0", func() {
	// 		Expect(val).ShouldNot(BeZero())
	// 	})

	// 	It("Must Not Be Nil", func() {
	// 		Expect(val).ShouldNot(BeNil())
	// 	})

	// 	It("Must Be of type Ethash", func() {

	// 		datasetTypeOf := reflect.TypeOf(dataset{})
	// 		Expect(reflect.TypeOf(val)).Should(Equal(reflect.TypeOf(datasetTypeOf)))
	// 	})

	// })

	var _ = Describe(".Hashrate", func() {

		config := Config{}
		newEllMiner := New(config)

		val := newEllMiner.Hashrate()
		It("Must Not Be LESS TNAN 0", func() {
			Expect(val).ShouldNot(BeNumerically("<", 0))
		})

		It("Must Not Be an Error", func() {
			Expect(val).ShouldNot(BeIdenticalTo(error.Error))
		})

		It("Must Not Be Nil", func() {
			Expect(val).ShouldNot(BeNil())
		})

	})

	var _ = Describe(".cacheSize", func() {

		val := cacheSize(6753)

		It("Must Not Be LESS TNAN 0", func() {
			Expect(val).ShouldNot(BeNumerically("<", 0))
		})

		It("Must Not Be an Error", func() {
			Expect(val).ShouldNot(BeIdenticalTo(error.Error))
		})

		It("Must Not Be Nil", func() {
			Expect(val).ShouldNot(BeNil())
		})

		It("Must Not Be 0", func() {
			Expect(val).ShouldNot(BeZero())
		})

	})

	var _ = Describe(".calcCacheSize", func() {

		val := calcCacheSize(6753)

		It("Must Not Be LESS TNAN 0", func() {
			Expect(val).ShouldNot(BeNumerically("<", 0))
		})

		It("Must Not Be an Error", func() {
			Expect(val).ShouldNot(BeIdenticalTo(error.Error))
		})

		It("Must Not Be 0", func() {
			Expect(val).ShouldNot(BeZero())
		})

		It("Must Not Be Nil", func() {
			Expect(val).ShouldNot(BeNil())
		})

	})

	var _ = Describe(".datasetSize", func() {

		val := datasetSize(6753)

		It("Must Not Be LESS TNAN 0", func() {
			Expect(val).ShouldNot(BeNumerically("<", 0))
		})

		It("Must Not Be an Error", func() {
			Expect(val).ShouldNot(BeIdenticalTo(error.Error))
		})

		It("Must Not Be Nil", func() {
			Expect(val).ShouldNot(BeNil())
		})

		It("Must Not Be 0", func() {
			Expect(val).ShouldNot(BeZero())
		})

	})

	var _ = Describe(".calcDatasetSize", func() {

		val := calcDatasetSize(6753)

		It("Must Not Be LESS TNAN 0", func() {
			Expect(val).ShouldNot(BeNumerically("<", 0))
		})

		It("Must Not Be an Error", func() {
			Expect(val).ShouldNot(BeIdenticalTo(error.Error))
		})

		It("Must Not Be Nil", func() {
			Expect(val).ShouldNot(BeNil())
		})

		It("Must Not Be 0", func() {
			Expect(val).ShouldNot(BeZero())
		})

	})

	var _ = Describe(".seedHash", func() {

		val := seedHash(6753)

		It("Must Not Be an Error", func() {
			Expect(val).ShouldNot(BeIdenticalTo(error.Error))
		})

		It("Must Not Be Nil", func() {
			Expect(val).ShouldNot(BeNil())
		})

		It("Must Not Be 0", func() {
			Expect(val).ShouldNot(BeZero())
		})

		// if block number is greater than epoch length
		val1 := seedHash(30001)

		It("Must Not Be an Error", func() {
			Expect(val1).ShouldNot(BeIdenticalTo(error.Error))
		})

		It("Must Not Be Nil", func() {
			Expect(val1).ShouldNot(BeNil())
		})

	})

	var _ = Describe(".fnv", func() {

		val := fnv(6753, 8765)

		It("Must Not Be an Error", func() {
			Expect(val).ShouldNot(BeIdenticalTo(error.Error))
		})

		It("Must Not Be Nil", func() {
			Expect(val).ShouldNot(BeNil())
		})

		It("Must Not Be 0", func() {
			Expect(val).ShouldNot(BeZero())
		})

	})

	var _ = Describe(".calcDifficultyHomestead", func() {

		val := calcDifficultyHomestead(20180523080200, big.NewInt(20180522080200), big.NewInt(67896))

		It("Must Not Be an Error", func() {
			Expect(val).ShouldNot(BeIdenticalTo(error.Error))
		})

		It("Must Not Be Nil", func() {
			Expect(val).ShouldNot(BeNil())
		})

		It("Must Not Be 0", func() {
			Expect(val).ShouldNot(BeZero())
		})

	})

	var _ = Describe(".isLittleEndian", func() {

		isLittleEndian := isLittleEndian()

		It("Must return Type of isLittleEndian bool value True/False", func() {
			Expect(isLittleEndian).To(SatisfyAny(Equal(true), Equal(false)))
		})

		It("Must Not Be Nil", func() {
			Expect(isLittleEndian).ShouldNot(BeNil())
		})

		It("Must Not Be 0", func() {
			Expect(isLittleEndian).ShouldNot(BeZero())
		})

	})

	var _ = Describe("VerifyPOW", func() {

		// config struct to innitialize the miner package
		config := Config{
			CacheDir: "CacheFile", CachesInMem: 0, CachesOnDisk: 1, DatasetDir: "DagFile", DatasetsInMem: 0, DatasetsOnDisk: 1, PowMode: ModeFake,
		}

		//Create a New Ethash Constructor
		newEllMiner := New(config)

		ellBlock := DB.GetSingleBlock("20")
		ellBlock.Difficulty = "30678"

		errPow := newEllMiner.VerifyPOW(&ellBlock)

		It("It must not be 0", func() {
			Expect(errPow).ShouldNot(BeZero())
		})

		Context("Sending Arbitrary Value into VerifyPOW function", func() {

			var _ = Describe(".VerifyPOW Negative Block Difficulty", func() {
				ellBlock.Difficulty = "-30678"
				errPow := newEllMiner.VerifyPOW(&ellBlock)
				It("Should return non-positive difficulty error", func() {
					Expect(errPow).Should(Equal(errors.New("non-positive difficulty")))
				})

			})

			var _ = Describe(".VerifyPOW Negative Block Difficulty", func() {
				ellBlock.Number = 0
				ellBlock.Difficulty = "30678"
				errPow := newEllMiner.VerifyPOW(&ellBlock)
				It("Should return non Positive Block Number", func() {
					Expect(errPow).Should(Equal(errors.New("non Positive Block Number")))
				})

			})

			var _ = Describe(".VerifyPOW Invalid Mix digest", func() {
				ellBlock.Number = 8
				ellBlock.Difficulty = "30678"
				errPow := newEllMiner.VerifyPOW(&ellBlock)
				It("Should return invalid mix digest", func() {
					Expect(errPow).Should(Equal(errors.New("invalid mix digest")))
				})

			})

		})

	})

})
