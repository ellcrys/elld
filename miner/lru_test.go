package miner

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Lru", func() {

	var _ = Describe(".newlru", func() {

		lruRes := newlru("cache", 0, newCache)

		It("Should be type of lru struct", func() {
			Expect(lruRes).ShouldNot(BeIdenticalTo(lru{}))
		})
	})

	var _ = Describe(".get", func() {

		lru := newlru("cache", 0, newCache)
		item, future := lru.get(6789)

		It("Resuls Should Not be Nil", func() {
			Expect(item).ShouldNot(BeNil())
			Expect(future).Should(BeNil())
		})

	})
})
