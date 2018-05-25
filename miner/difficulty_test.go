package miner

import (
	"math/big"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe(".CalcDifficulty", func() {

	currentTime := uint64(20180515105500)
	parentTime := big.NewInt(20180514105500)
	parentDifficulty := big.NewInt(31083)
	parentBlockNumber := big.NewInt(30)

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
