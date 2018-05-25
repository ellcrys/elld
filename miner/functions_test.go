package miner

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Functions", func() {

	var _ = Describe(".cacheSize", func() {

		val := cacheSize(6753)

		It("Must Not Be LESS TNAN 0", func() {
			Expect(val).ShouldNot(BeZero())
		})
	})

	var _ = Describe(".calcCacheSize", func() {

		val := calcCacheSize(6753)

		It("Must Not Be LESS TNAN 0", func() {
			Expect(val).ShouldNot(BeZero())
		})

	})

	var _ = Describe(".calcDatasetSize", func() {

		val := calcDatasetSize(6753)

		It("Must Not Be LESS TNAN 0", func() {
			Expect(val).ShouldNot(BeZero())
		})

	})

	var _ = Describe(".seedHash", func() {

		val := seedHash(6753)

		It("Must Not Be  0", func() {
			Expect(val).ShouldNot(BeZero())
		})

	})

	var _ = Describe(".datasetSize", func() {

		val := datasetSize(6753)

		It("Must Not Be  0", func() {
			Expect(val).ShouldNot(BeZero())
		})

	})

	var _ = Describe(".fnv", func() {

		val := fnv(6753, 8765)

		It("Must Not Be  0", func() {
			Expect(val).ShouldNot(BeZero())
		})

	})

})
