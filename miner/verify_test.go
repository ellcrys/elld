package miner

import (
	"errors"

	"github.com/ellcrys/elld/wire"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Verify", func() {

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

	config := Config{
		CacheDir: "CacheFile", CachesInMem: 0, CachesOnDisk: 1, DatasetDir: "DagFile", DatasetsInMem: 0, DatasetsOnDisk: 1, PowMode: ModeFake,
	}
	newEllMiner := New(config)

	It("Parsing a normal block; It must not be an error", func() {
		errPow := newEllMiner.VerifyPOW(&ellBlock)
		Expect(errPow).ShouldNot(BeIdenticalTo(error.Error))
	})

	It("Parsing Negative Block Number ; Should return non-positive difficulty error", func() {
		ellBlock.Difficulty = "-30678"
		errPow := newEllMiner.VerifyPOW(&ellBlock)
		Expect(errPow).Should(Equal(errors.New("non-positive difficulty")))
	})

	It("Parsing 0 as Block Dificulty; Should return non Positive Block Number", func() {
		ellBlock.Number = 0
		errPow := newEllMiner.VerifyPOW(&ellBlock)
		Expect(errPow).Should(Equal(errors.New("non Positive Block Number")))
	})

	It("Arbitrary Block info;  Should return invalid mix digest", func() {
		ellBlock.Number = 8
		ellBlock.Difficulty = "30678"
		errPow := newEllMiner.VerifyPOW(&ellBlock)
		Expect(errPow).Should(Equal(errors.New("invalid mix digest")))
	})

})
