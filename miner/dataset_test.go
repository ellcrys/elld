package miner

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Dataset", func() {

	It("newDataset must not be Nil", func() {
		epoch := uint64(98)
		ds := newDataset(epoch)
		Expect(ds).ShouldNot(BeNil())
	})

})
