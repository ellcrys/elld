package util

import (
	"math/big"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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

	Describe(".AscOrderBigIntMeta", func() {
		It("should return slice in this order [1,2,3]", func() {
			v := []*BigIntWithMeta{
				{Int: big.NewInt(3)},
				{Int: big.NewInt(1)},
				{Int: big.NewInt(2)},
			}
			AscOrderBigIntMeta(v)
			Expect(v[0].Int.Int64()).To(Equal(int64(1)))
			Expect(v[1].Int.Int64()).To(Equal(int64(2)))
			Expect(v[2].Int.Int64()).To(Equal(int64(3)))
		})
	})

	Describe(".NonZeroOrDefIn64", func() {
		It("should return 3 when v=0", func() {
			Expect(NonZeroOrDefIn64(0, 3)).To(Equal(int64(3)))
		})

		It("should return 2 when v=2", func() {
			Expect(NonZeroOrDefIn64(2, 3)).To(Equal(int64(2)))
		})
	})

	Describe(".StrToDec", func() {
		It("should panic if value is not numeric", func() {
			val := "129.1a"
			Expect(func() {
				StrToDec(val)
			}).To(Panic())
		})
	})
})
