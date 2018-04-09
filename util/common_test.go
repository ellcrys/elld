package util_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/ellcrys/druid/util"
)

var _ = Describe("Common", func() {
	Describe(".StructToBytes", func() {
		It("should return []byte{123, 34, 78, 97, 109, 101, 34, 58, 34, 98, 101, 110, 34, 125} bytes", func() {
			s := struct{ Name string }{Name: "ben"}
			expected := []byte{123, 34, 78, 97, 109, 101, 34, 58, 34, 98, 101, 110, 34, 125}
			bs := StructToBytes(s)
			Expect(bs).To(Equal(expected))
		})
	})
})
