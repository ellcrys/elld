package util

import (
	"math/big"
	"os"
	"path/filepath"

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

	Describe(".Untar", func() {

		var dest string

		BeforeEach(func() {
			dest = filepath.Join("./testdata", "untar")
			err := os.MkdirAll(dest, 0755)
			Expect(err).To(BeNil())
		})

		AfterEach(func() {
			err := os.RemoveAll(dest)
			Expect(err).To(BeNil())
		})

		Context("tar file with no root directory", func() {
			It("should successfully untar and return destination as root", func() {
				f, err := os.Open("./testdata/sample.tar")
				Expect(err).To(BeNil())
				defer f.Close()
				root, err := Untar(dest, f)
				Expect(err).To(BeNil())
				Expect(dest).To(Equal(root))
				_, err = os.Stat(filepath.Join(root, "sample.txt"))
				Expect(err).To(BeNil())
			})
		})

		It("should return root", func() {
			f, err := os.Open("./testdata/sampledir.tar")
			Expect(err).To(BeNil())
			defer f.Close()
			root, err := Untar(dest, f)
			Expect(err).To(BeNil())
			Expect(dest).ToNot(Equal(root))
			Expect(root).ToNot(Equal(filepath.Join(dest, "sampledata")))
		})
	})
})
