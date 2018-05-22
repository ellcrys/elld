package node

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("HistoryCache", func() {

	var err error
	var hc *HistoryCache

	BeforeEach(func() {
		hc, err = NewHistoryCache(10)
		Expect(err).To(BeNil())
	})

	Describe(".Add", func() {
		It("should add without error", func() {
			err := hc.Add([]interface{}{1, int64(10), "james"})
			Expect(err).To(BeNil())
		})
	})

	Describe(".Has", func() {
		It("should add without error", func() {
			key := []interface{}{1, int64(10), "james"}
			err := hc.Add(key)
			Expect(err).To(BeNil())
			has := hc.Has(key)
			Expect(has).To(BeTrue())
		})
	})
})
