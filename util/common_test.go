package util

import (
	"math/big"

	"github.com/ellcrys/elld/wire"

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

	Describe(".Int64ToHex", func() {
		It("should return 0x3130", func() {
			Expect(Int64ToHex(10)).To(Equal("0x3130"))
		})
	})

	Describe(".HexToInt64", func() {
		It("should return 0x3130", func() {
			str, err := HexToInt64("0x3130")
			Expect(err).To(BeNil())
			Expect(str).To(Equal(int64(10)))
		})
	})

	Describe(".StrToHex", func() {
		It("should return 0x3130", func() {
			Expect(StrToHex("10")).To(Equal("0x3130"))
		})
	})

	Describe(".HexToStr", func() {
		It("should return '10'", func() {
			str, err := HexToStr("0x3130")
			Expect(err).To(BeNil())
			Expect(str).To(Equal("10"))
		})
	})

	Describe(".SerializeMsg", func() {
		It("should return [10, 12, 115, 111, 109, 101, 95, 97, 100, 100, 114, 101, 115, 115, 16, 244, 154, 144, 216, 5]", func() {
			o := &wire.Address{Timestamp: 1526992244, Address: "some_address"}
			bs := SerializeMsg(o)
			Expect(bs).To(Equal([]byte{10, 12, 115, 111, 109, 101, 95, 97, 100, 100, 114, 101, 115, 115, 16, 244, 154, 144, 216, 5}))
		})
	})

	Describe(".ToHex", func() {
		It("should return hex equivalent", func() {
			v := ToHex([]byte("abc"))
			Expect(v).To(Equal("0x616263"))
		})
	})

	Describe(".FromHex", func() {
		When("hex value begins with '0x'", func() {
			It("should return bytes equivalent of hex", func() {
				v, _ := FromHex("0x616263")
				Expect(v).To(Equal([]byte("abc")))
			})
		})

		When("hex value does not begin with '0x'", func() {
			It("should return bytes equivalent of hex", func() {
				v, _ := FromHex("616263")
				Expect(v).To(Equal([]byte("abc")))
			})
		})
	})
})
