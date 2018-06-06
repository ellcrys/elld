package miner

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Cache", func() {

	It("newCache must not be Nil", func() {
		epoch := uint64(98)
		ds := newCache(epoch)
		Expect(ds).ShouldNot(BeNil())
	})

})
