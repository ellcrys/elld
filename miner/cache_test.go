package miner

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Cache", func() {

	Describe(".newCache", func() {
		It("It must not be Nil", func() {
			epoch := uint64(98)
			ds := newCache(epoch)
			Expect(ds).ShouldNot(BeNil())
		})
	})

})
