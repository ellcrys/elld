package miner

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Util", func() {

	Describe(".isLittleEndian", func() {
		It("It must be Boolean", func() {
			response := isLittleEndian()
			Expect(response).Should(BeTrue())
		})
	})

	Describe(".cacheSize", func() {
		It("It will use cacheSizes[epoch] & must not be Zero", func() {
			blockNumber := uint64(60)
			size := cacheSize(blockNumber)
			Expect(size).ShouldNot(BeZero())
		})

		It("It will use calcCacheSize(epoch) & must not be Zero", func() {
			blockNumber := uint64(99999999999999999)
			size := cacheSize(blockNumber)
			Expect(size).ShouldNot(BeZero())
		})
	})

	Describe(".calcCacheSize", func() {
		It("It must not be Zero", func() {
			epoch := 1000
			size := calcCacheSize(epoch)
			Expect(size).ShouldNot(BeZero())
		})
	})

	Describe(".datasetSize", func() {
		It("It must use datasetSizes[epoch] & must not be Zero", func() {
			blockNumber := uint64(60)
			size := datasetSize(blockNumber)
			Expect(size).ShouldNot(BeZero())
		})

		It("datasetSize will use calcDatasetSize(epoch) & must not be Zero", func() {
			blockNumber := uint64(99999999999999999)
			size := datasetSize(blockNumber)
			Expect(size).ShouldNot(BeZero())
		})
	})

	Describe(".calcDatasetSize", func() {
		It("It must not be Zero", func() {
			epoch := 1000
			size := calcDatasetSize(epoch)
			Expect(size).ShouldNot(BeZero())
		})
	})

	Describe(".SeedHash", func() {
		It("It must not be Zero", func() {
			block := uint64(98)
			seed := SeedHash(block)
			Expect(len(seed)).ShouldNot(BeZero())
		})
	})

	Describe(".seedHash", func() {
		It("It must not be Zero", func() {
			block := uint64(98)
			seed := seedHash(block)
			Expect(len(seed)).ShouldNot(BeZero())
		})
	})

	Describe(".fnv", func() {
		It("It must not be Zero", func() {
			a := uint32(98)
			b := uint32(8)
			res := fnv(a, b)
			Expect(res).ShouldNot(BeZero())
		})
	})

	Describe(".rHash256", func() {
		It("It must not be Zero", func() {
			res := rHash256([]byte("ellcrys"))
			Expect(len(res)).ShouldNot(BeZero())
		})
	})

	Describe(".rHash512", func() {
		It("It must not be Zero", func() {
			res := rHash512([]byte("ellcrys"))
			Expect(len(res)).ShouldNot(BeZero())
		})
	})

})
