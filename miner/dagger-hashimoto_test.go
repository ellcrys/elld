package miner

import (
	"encoding/binary"

	"github.com/ellcrys/elld/wire"
	"github.com/ellcrys/go-ethereum/crypto/sha3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DaggerHashimoto", func() {

	var ellBlock wire.Block

	BeforeEach(func() {
		ellBlock = wire.Block{
			Difficulty:     "70212",
			HashMerkleRoot: "8c4a16df5c399bb06a8752ff16f776901a714dfa3f4113be2c14be7c136ef582",
			HashPrevBlock:  "efae6c6522095e57bf885756af7ccc38483e6582b83a80adf588126f03134b78",
			Nounce:         1525873090716071984,
			Number:         2,
			PowHash:        "508f043b157f888e361b04e891423651c2baa87d70b03a5354bed24b2d3125bd",
			PowResult:      "000157eb1e61cc522aab7211e5afa77bd7d30b197827b63714ffbb543eeea833",
			Time:           "20180505144545",
			Version:        "1.0",
		}
	})

	var _ = Describe(".hashimoto", func() {

		keccak512 := makeHasher(sha3.NewKeccak512())
		cache := make([]uint32, 1000000)
		lookup := func(index uint32) []uint32 {
			rawData := generateDatasetItem(cache, index, keccak512)
			data := make([]uint32, len(rawData)/4)
			for i := 0; i < len(data); i++ {
				data[i] = binary.LittleEndian.Uint32(rawData[i*4:])
			}
			return data
		}

		hash := make([]byte, 30)
		digest, result := hashimoto(hash, ellBlock.Nounce, 1024, lookup)

		It("hashimoto Result ", func() {
			Expect(digest).ShouldNot(BeEmpty())
			Expect(result).ShouldNot(BeEmpty())
		})

	})
	var _ = Describe(".hashimotoLight", func() {

		cache := make([]uint32, 30)
		hash := make([]byte, 30)
		digest, result := hashimotoLight(1024, cache, hash, ellBlock.Nounce)
		It("hashimotoLight Result ", func() {
			Expect(digest).ShouldNot(BeEmpty())
			Expect(result).ShouldNot(BeEmpty())
		})

	})

	var _ = Describe(".hashimotoFull", func() {

		dataset := make([]uint32, 50)
		hash := make([]byte, 30)

		digest, result := hashimotoFull(dataset, hash, ellBlock.Nounce)

		It("hashimotoLight Result ", func() {
			Expect(digest).ShouldNot(BeEmpty())
			Expect(result).ShouldNot(BeEmpty())
		})
	})
})
