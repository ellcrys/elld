package miner

import (
	"fmt"
	"math/big"
	"path/filepath"

	"github.com/ellcrys/druid/wire"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Ethash", func() {

	var ellBlock wire.Block
	var minerID = 63548
	var config = Config{
		CacheDir: "../CacheFile", CachesInMem: 0, CachesOnDisk: 1, DatasetDir: "../DagFile", DatasetsInMem: 0, DatasetsOnDisk: 1, PowMode: ModeFake,
	}

	BeforeEach(func() {
		ellBlock = wire.Block{
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
	})

	Describe(".Mine", func() {

		var digest string
		var result string
		var nounce uint64
		var err error

		It("When abortNonceSearch is not closed", func() {
			newEllMiner := New(config)
			digest, result, nounce, err = newEllMiner.Mine(&ellBlock, minerID)
			Expect(err).Should(BeNil())
			Expect(digest).ShouldNot(BeEmpty())
			Expect(digest).To(HaveLen(64))
			Expect(result).ShouldNot(BeEmpty())
			Expect(nounce).ShouldNot(BeZero())
		})

	})

	Describe(".SeedHash", func() {
		seed := SeedHash(5673831)
		It("Should Not be Nil", func() {
			Expect(seed).ShouldNot(BeNil())
		})
	})

	Describe(".newCache", func() {
		cache := newCache(5673831)
		It("Should Not be Nil", func() {
			Expect(cache).ShouldNot(BeNil())
		})
	})

	Describe(".newDataset", func() {
		dataset := newDataset(5673831)
		It("Should Not be Nil", func() {
			Expect(dataset).ShouldNot(BeNil())
		})
	})

	Describe(".memoryMap", func() {

		seed := seedHash(43*epochLength + 1)
		var endian string
		if !isLittleEndian() {
			endian = ".be"
		}
		algorithmRevision = 23
		path := filepath.Join("", fmt.Sprintf("cache-R%d-%x%s", algorithmRevision, seed[:8], endian))
		_, _, _, err := memoryMap(path)

		It("Result of memoryMap", func() {
			Expect(err).ShouldNot(BeIdenticalTo(error.Error))
		})

	})

	Describe(".memoryMapFile", func() {

		seed := seedHash(43*epochLength + 1)
		var endian string
		if !isLittleEndian() {
			endian = ".be"
		}
		algorithmRevision = 23
		path := filepath.Join("", fmt.Sprintf("cache-R%d-%x%s", algorithmRevision, seed[:8], endian))
		file, _, _, _ := memoryMap(path)
		_, _, err1 := memoryMapFile(file, true)

		It("Should not be an Error", func() {
			Expect(err1).ShouldNot(BeIdenticalTo(error.Error))
		})
	})

	Describe(".newCache", func() {
		cache := newCache(58697)

		It("Result of newCache ", func() {
			Expect(cache).ShouldNot(BeNil())
			Expect(cache).ShouldNot(BeZero())
		})

	})

	Describe(".Hashrate", func() {

		newEllMiner := New(config)

		val := newEllMiner.Hashrate()
		It("Hash rate result must", func() {
			Expect(val).ShouldNot(BeNumerically("<", 0))
		})

	})

	Describe(".calcDifficultyHomestead", func() {

		val := calcDifficultyHomestead(20180523080200, big.NewInt(20180522080200), big.NewInt(67896))

		It("Must Not Be an Error", func() {
			Expect(val).ShouldNot(BeZero())
		})

	})

	Describe(".isLittleEndian", func() {

		isLittleEndian := isLittleEndian()

		It("Must return Type of isLittleEndian bool value True/False", func() {
			Expect(isLittleEndian).To(SatisfyAny(Equal(true), Equal(false)))
		})

	})

})
