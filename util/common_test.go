package util

import (
	"io/ioutil"
	"math/big"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Common", func() {

	Describe(".ObjectsToBytes", func() {
		It("should return encode", func() {
			s := struct{ Name string }{Name: "ben"}
			expected := []uint8{
				0x81, 0xa4, 0x4e, 0x61, 0x6d, 0x65, 0xa3, 0x62, 0x65, 0x6e,
			}
			bs := ObjectToBytes(s)
			Expect(bs).To(Equal(expected))
		})
	})

	Describe(".BytesToObject", func() {

		var bs []byte
		var m = map[string]interface{}{"stuff": int8(10)}

		BeforeEach(func() {
			bs = ObjectToBytes(m)
			Expect(bs).ToNot(BeEmpty())
		})

		It("should decode to expected value", func() {
			var actual map[string]interface{}
			err := BytesToObject(bs, &actual)
			Expect(err).To(BeNil())
			Expect(actual).To(Equal(m))
		})
	})

	Describe(".IsPathOk", func() {

		BeforeEach(func() {
			err := os.Mkdir("a_dir_here", 0655)
			Expect(err).To(BeNil())
		})

		AfterEach(func() {
			err := os.Remove("a_dir_here")
			Expect(err).To(BeNil())
		})

		It("should return true when path exists", func() {
			Expect(IsPathOk("./a_dir_here")).To(BeTrue())
		})

		It("should return false when path does not exists", func() {
			Expect(IsPathOk("./abcxyz")).To(BeFalse())
		})
	})

	Describe(".IsFileOk", func() {

		BeforeEach(func() {
			err := os.Mkdir("a_dir_here", 0700)
			Expect(err).To(BeNil())
			err = ioutil.WriteFile("./a_dir_here/a_file", []byte("abc"), 0700)
			Expect(err).To(BeNil())
		})

		AfterEach(func() {
			err := os.RemoveAll("a_dir_here")
			Expect(err).To(BeNil())
		})

		It("should return true when path exists", func() {
			Expect(IsFileOk("./a_dir_here/a_file")).To(BeTrue())
		})

		It("should return false when path does not exists", func() {
			Expect(IsFileOk("./abcxyz")).To(BeFalse())
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

	Describe(".MustFromHex", func() {
		When("hex value begins with '0x'", func() {
			It("should return bytes equivalent of hex", func() {
				v := MustFromHex("0x616263")
				Expect(v).To(Equal([]byte("abc")))
			})
		})

		When("hex value is not valid'", func() {
			It("should panic", func() {
				Expect(func() {
					MustFromHex("sa&616263")
				}).To(Panic())
			})
		})
	})

	Describe("String", func() {
		Describe(".Bytes", func() {
			It("should return expected bytes value", func() {
				s := String("abc")
				Expect(s.Bytes()).To(Equal([]uint8{0x61, 0x62, 0x63}))
			})
		})

		Describe(".Equal", func() {
			It("should equal b", func() {
				a := String("abc")
				b := String("abc")
				Expect(a.Equal(b)).To(BeTrue())
			})

			It("should not equal b", func() {
				a := String("abc")
				b := String("xyz")
				Expect(a.Equal(b)).ToNot(BeTrue())
			})
		})

		Describe(".SS", func() {
			Context("when string is greater than 32 characters", func() {
				It("should return short form", func() {
					s := String("abcdefghijklmnopqrstuvwxyz12345678")
					Expect(s.SS()).To(Equal("abcdefghij...yz12345678"))
				})
			})

			Context("when string is less than 32 characters", func() {
				It("should return unchanged", func() {
					s := String("abcdef")
					Expect(s.SS()).To(Equal("abcdef"))
				})
			})
		})

		Describe(".Decimal", func() {
			It("should return decimal", func() {
				d := String("12.50").Decimal()
				Expect(d.String()).To(Equal("12.5"))
			})

			It("should panic if string is not convertible to decimal", func() {
				Expect(func() {
					String("12a50").Decimal()
				}).To(Panic())
			})
		})

		Describe(".IsDecimal", func() {
			It("should return true if convertible to decimal", func() {
				actual := String("12.50").IsDecimal()
				Expect(actual).To(BeTrue())
			})

			It("should return false if not convertible to decimal", func() {
				actual := String("12a50").IsDecimal()
				Expect(actual).To(BeFalse())
			})
		})
	})

	Describe(".StructToMap", func() {

		type testStruct struct {
			Name string
		}

		It("should return correct map equivalent", func() {
			s := testStruct{Name: "odion"}
			expected := map[string]interface{}{"Name": "odion"}
			Expect(StructToMap(s)).To(Equal(expected))
		})

	})

	Describe(".GetPtrAddr", func() {
		It("should get numeric pointer address", func() {
			name := "xyz"
			ptrAddr := GetPtrAddr(name)
			Expect(ptrAddr.Cmp(Big0)).To(Equal(1))
		})
	})

	Describe(".MapDecode", func() {

		type testStruct struct {
			Name string
		}

		It("should decode map to struct", func() {
			var m = map[string]interface{}{"Name": "abc"}
			var s testStruct
			err := MapDecode(m, &s)
			Expect(err).To(BeNil())
			Expect(s.Name).To(Equal(m["Name"]))
		})
	})

	Describe(".EncodeNumber", func() {
		It("should encode number to expected byte", func() {
			encVal := EncodeNumber(100)
			Expect(encVal).To(Equal([]uint8{
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x64,
			}))
		})
	})

	Describe(".DecodeNumber", func() {
		It("should decode bytes value to 100", func() {
			decVal := DecodeNumber([]uint8{
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x64,
			})
			Expect(decVal).To(Equal(uint64(100)))
		})

		It("should panic if unable to decode", func() {
			Expect(func() {
				DecodeNumber([]byte("n10a"))
			}).To(Panic())
		})
	})

	Describe(".MayDecodeNumber", func() {
		It("should decode bytes value to 100", func() {
			decVal, err := MayDecodeNumber([]uint8{
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x64,
			})
			Expect(err).To(BeNil())
			Expect(decVal).To(Equal(uint64(100)))
		})

		It("should return error if unable to decode", func() {
			_, err := MayDecodeNumber([]byte("n10a"))
			Expect(err).ToNot(BeNil())
		})
	})

	Describe(".ToJSFriendlyMap", func() {

		type test1 struct {
			Name string
			Desc []byte
		}

		type test2 struct {
			Age    int
			Others test1
			More   []interface{}
		}

		type test3 struct {
			Sig Hash
		}

		type test4 struct {
			Num *big.Int
		}

		type test5 struct {
			Num BlockNonce
		}

		It("should return expected output", func() {
			t1 := test1{
				Name: "fred",
				Desc: []byte("i love games"),
			}
			result := EncodeForJS(t1)
			Expect(result).To(Equal(map[string]interface{}{"Name": "fred",
				"Desc": "0x69206c6f76652067616d6573",
			}))

			t2 := test2{
				Age:    20,
				Others: t1,
			}
			result = EncodeForJS(t2)
			Expect(result.(map[string]interface{})["Others"]).To(Equal(map[string]interface{}{
				"Name": "fred",
				"Desc": "0x69206c6f76652067616d6573",
			}))

			t3 := test2{
				Age:    20,
				Others: t1,
				More:   []interface{}{t1},
			}
			result = EncodeForJS(t3)
			Expect(result.(map[string]interface{})["More"]).To(Equal([]interface{}{
				map[string]interface{}{"Name": "fred",
					"Desc": "0x69206c6f76652067616d6573",
				},
			}))

			t4 := test3{
				Sig: StrToHash("fred"),
			}
			result = EncodeForJS(t4)
			Expect(result).To(Equal(map[string]interface{}{"Sig": "0x6672656400000000000000000000000000000000000000000000000000000000"}))

			t5 := test4{
				Num: new(big.Int).SetInt64(10),
			}
			result = EncodeForJS(t5)
			Expect(result).To(Equal(map[string]interface{}{"Num": "0xa"}))

			t6 := test5{
				Num: EncodeNonce(10),
			}
			result = EncodeForJS(t6)
			Expect(result).To(Equal(map[string]interface{}{"Num": "0x000000000000000a"}))
		})

		Context("With ignoreField specified", func() {
			t1 := test2{Age: 30, Others: test1{Desc: []byte("i love games")}}

			BeforeEach(func() {
				result := EncodeForJS(t1)
				Expect(result.(map[string]interface{})["Age"]).To(Equal("0x1e"))
			})

			It("should not modify field", func() {
				result := EncodeForJS(t1, "Age")
				Expect(result.(map[string]interface{})["Age"]).To(Equal(30))
			})
		})
	})

	Describe(".IsBoolChanClosed", func() {
		It("should return false if not closed", func() {
			c := make(chan bool)
			Expect(IsBoolChanClosed(c)).To(BeFalse())
		})

		It("should return true if closed", func() {
			c := make(chan bool)
			close(c)
			Expect(IsBoolChanClosed(c)).To(BeTrue())
		})
	})
})
