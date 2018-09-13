package util

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Common", func() {
	Describe(".StructToBytes", func() {
		It("should return []byte{123, 34, 78, 97, 109, 101, 34, 58, 34, 98, 101, 110, 34, 125} bytes", func() {
			s := struct{ Name string }{Name: "ben"}
			expected := []uint8{
				0x81, 0xa4, 0x4e, 0x61, 0x6d, 0x65, 0xa3, 0x62, 0x65, 0x6e,
			}
			bs := ObjectToBytes(s)
			Expect(bs).To(Equal(expected))
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
		It("should successfully serialize object", func() {
			o := []interface{}{1, 2, 3}
			bs := SerializeMsg(o)
			Expect(bs).To(Equal([]byte{147, 1, 2, 3}))
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
