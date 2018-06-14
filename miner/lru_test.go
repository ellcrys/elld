package miner

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Lru", func() {

	Describe(".newlru", func() {
		It("It Should be of type lru struct", func() {
			lruRes := newlru("cache", 0, newCache)
			Expect(lruRes).ShouldNot(BeIdenticalTo(lru{}))
		})
	})
})
