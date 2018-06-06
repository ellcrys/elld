package miner

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Util", func() {

	It("isLittleEndian must be Boolean", func() {
		response := isLittleEndian()
		Expect(response).Should(BeTrue())
	})

	It("cacheSize must use cacheSizes[epoch] & must not be Zero", func() {
		blockNumber := uint64(60)
		size := cacheSize(blockNumber)
		Expect(size).ShouldNot(BeZero())
	})

	It("cacheSize will use calcCacheSize(epoch) & must not be Zero", func() {
		blockNumber := uint64(99999999999999999)
		size := cacheSize(blockNumber)
		Expect(size).ShouldNot(BeZero())
	})

	It("calcCacheSize must not be Zero", func() {
		epoch := 1000
		size := calcCacheSize(epoch)
		Expect(size).ShouldNot(BeZero())
	})

	It("datasetSize must use datasetSizes[epoch] & must not be Zero", func() {
		blockNumber := uint64(60)
		size := datasetSize(blockNumber)
		Expect(size).ShouldNot(BeZero())
	})

	It("datasetSize will use calcDatasetSize(epoch) & must not be Zero", func() {
		blockNumber := uint64(99999999999999999)
		size := datasetSize(blockNumber)
		Expect(size).ShouldNot(BeZero())
	})

	It("calcDatasetSize must not be Zero", func() {
		epoch := 1000
		size := calcDatasetSize(epoch)
		Expect(size).ShouldNot(BeZero())
	})

	It("SeedHash length must not be Zero", func() {
		block := uint64(98)
		seed := SeedHash(block)
		Expect(len(seed)).ShouldNot(BeZero())
	})

})
