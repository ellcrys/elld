package util

import (
	"io/ioutil"
	"math/big"
	"os"
	"testing"

	. "github.com/ncodes/goblin"
	. "github.com/onsi/gomega"
)

func TestCommon(t *testing.T) {
	g := Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("Common", func() {
		g.Describe(".ObjectsToBytes", func() {
			g.It("should return encode", func() {
				s := struct{ Name string }{Name: "ben"}
				expected := []uint8{
					0x81, 0xa4, 0x4e, 0x61, 0x6d, 0x65, 0xa3, 0x62, 0x65, 0x6e,
				}
				bs := ObjectToBytes(s)
				Expect(bs).To(Equal(expected))
			})
		})

		g.Describe(".BytesToObject", func() {

			var bs []byte
			var m = map[string]interface{}{"stuff": int8(10)}

			g.BeforeEach(func() {
				bs = ObjectToBytes(m)
				Expect(bs).ToNot(BeEmpty())
			})

			g.It("should decode to expected value", func() {
				var actual map[string]interface{}
				err := BytesToObject(bs, &actual)
				Expect(err).To(BeNil())
				Expect(actual).To(Equal(m))
			})
		})

		g.Describe(".IsPathOk", func() {

			g.BeforeEach(func() {
				err := os.Mkdir("a_dir_here", 0655)
				Expect(err).To(BeNil())
			})

			g.AfterEach(func() {
				err := os.Remove("a_dir_here")
				Expect(err).To(BeNil())
			})

			g.It("should return true when path exists", func() {
				Expect(IsPathOk("./a_dir_here")).To(BeTrue())
			})

			g.It("should return false when path does not exists", func() {
				Expect(IsPathOk("./abcxyz")).To(BeFalse())
			})
		})

		g.Describe(".IsFileOk", func() {

			g.BeforeEach(func() {
				err := os.Mkdir("a_dir_here", 0700)
				Expect(err).To(BeNil())
				err = ioutil.WriteFile("./a_dir_here/a_file", []byte("abc"), 0700)
				Expect(err).To(BeNil())
			})

			g.AfterEach(func() {
				err := os.RemoveAll("a_dir_here")
				Expect(err).To(BeNil())
			})

			g.It("should return true when path exists", func() {
				Expect(IsFileOk("./a_dir_here/a_file")).To(BeTrue())
			})

			g.It("should return false when path does not exists", func() {
				Expect(IsFileOk("./abcxyz")).To(BeFalse())
			})
		})

		g.Describe(".NonZeroOrDefIn64", func() {
			g.It("should return 3 when v=0", func() {
				Expect(NonZeroOrDefIn64(0, 3)).To(Equal(int64(3)))
			})

			g.It("should return 2 when v=2", func() {
				Expect(NonZeroOrDefIn64(2, 3)).To(Equal(int64(2)))
			})
		})

		g.Describe(".StrToDec", func() {
			g.It("should panic if value is not numeric", func() {
				val := "129.1a"
				Expect(func() {
					StrToDec(val)
				}).To(Panic())
			})
		})

		g.Describe(".Int64ToHex", func() {
			g.It("should return 0x3130", func() {
				Expect(Int64ToHex(10)).To(Equal("0x3130"))
			})
		})

		g.Describe(".HexToInt64", func() {
			g.It("should return 0x3130", func() {
				str, err := HexToInt64("0x3130")
				Expect(err).To(BeNil())
				Expect(str).To(Equal(int64(10)))
			})
		})

		g.Describe(".StrToHex", func() {
			g.It("should return 0x3130", func() {
				Expect(StrToHex("10")).To(Equal("0x3130"))
			})
		})

		g.Describe(".HexToStr", func() {
			g.It("should return '10'", func() {
				str, err := HexToStr("0x3130")
				Expect(err).To(BeNil())
				Expect(str).To(Equal("10"))
			})
		})

		g.Describe(".SerializeMsg", func() {
			g.It("should successfully serialize object", func() {
				o := []interface{}{1, 2, 3}
				bs := SerializeMsg(o)
				Expect(bs).To(Equal([]byte{147, 1, 2, 3}))
			})
		})

		g.Describe(".ToHex", func() {
			g.It("should return hex equivalent", func() {
				v := ToHex([]byte("abc"))
				Expect(v).To(Equal("0x616263"))
			})
		})

		g.Describe(".FromHex", func() {
			g.When("hex value begins with '0x'", func() {
				g.It("should return bytes equivalent of hex", func() {
					v, _ := FromHex("0x616263")
					Expect(v).To(Equal([]byte("abc")))
				})
			})

			g.When("hex value does not begin with '0x'", func() {
				g.It("should return bytes equivalent of hex", func() {
					v, _ := FromHex("616263")
					Expect(v).To(Equal([]byte("abc")))
				})
			})
		})

		g.Describe(".MustFromHex", func() {
			g.When("hex value begins with '0x'", func() {
				g.It("should return bytes equivalent of hex", func() {
					v := MustFromHex("0x616263")
					Expect(v).To(Equal([]byte("abc")))
				})
			})

			g.When("hex value is not valid'", func() {
				g.It("should panic", func() {
					Expect(func() {
						MustFromHex("sa&616263")
					}).To(Panic())
				})
			})
		})

		g.Describe("String", func() {
			g.Describe(".Bytes", func() {
				g.It("should return expected bytes value", func() {
					s := String("abc")
					Expect(s.Bytes()).To(Equal([]uint8{0x61, 0x62, 0x63}))
				})
			})

			g.Describe(".Equal", func() {
				g.It("should equal b", func() {
					a := String("abc")
					b := String("abc")
					Expect(a.Equal(b)).To(BeTrue())
				})

				g.It("should not equal b", func() {
					a := String("abc")
					b := String("xyz")
					Expect(a.Equal(b)).ToNot(BeTrue())
				})
			})

			g.Describe(".SS", func() {
				g.Context("when string is greater than 32 characters", func() {
					g.It("should return short form", func() {
						s := String("abcdefghijklmnopqrstuvwxyz12345678")
						Expect(s.SS()).To(Equal("abcdefghij...yz12345678"))
					})
				})

				g.Context("when string is less than 32 characters", func() {
					g.It("should return unchanged", func() {
						s := String("abcdef")
						Expect(s.SS()).To(Equal("abcdef"))
					})
				})
			})

			g.Describe(".Decimal", func() {
				g.It("should return decimal", func() {
					d := String("12.50").Decimal()
					Expect(d.String()).To(Equal("12.5"))
				})

				g.It("should panic if string is not convertible to decimal", func() {
					Expect(func() {
						String("12a50").Decimal()
					}).To(Panic())
				})
			})

			g.Describe(".IsDecimal", func() {
				g.It("should return true if convertible to decimal", func() {
					actual := String("12.50").IsDecimal()
					Expect(actual).To(BeTrue())
				})

				g.It("should return false if not convertible to decimal", func() {
					actual := String("12a50").IsDecimal()
					Expect(actual).To(BeFalse())
				})
			})
		})

		g.Describe(".StructToMap", func() {

			type testStruct struct {
				Name string
			}

			g.It("should return correct map equivalent", func() {
				s := testStruct{Name: "odion"}
				expected := map[string]interface{}{"Name": "odion"}
				Expect(StructToMap(s)).To(Equal(expected))
			})

		})

		g.Describe(".GetPtrAddr", func() {
			g.It("should get numeric pointer address", func() {
				name := "xyz"
				ptrAddr := GetPtrAddr(name)
				Expect(ptrAddr.Cmp(Big0)).To(Equal(1))
			})
		})

		g.Describe(".MapDecode", func() {

			type testStruct struct {
				Name string
			}

			g.It("should decode map to struct", func() {
				var m = map[string]interface{}{"Name": "abc"}
				var s testStruct
				err := MapDecode(m, &s)
				Expect(err).To(BeNil())
				Expect(s.Name).To(Equal(m["Name"]))
			})
		})

		g.Describe(".EncodeNumber", func() {
			g.It("should encode number to expected byte", func() {
				encVal := EncodeNumber(100)
				Expect(encVal).To(Equal([]uint8{
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x64,
				}))
			})
		})

		g.Describe(".DecodeNumber", func() {
			g.It("should decode bytes value to 100", func() {
				decVal := DecodeNumber([]uint8{
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x64,
				})
				Expect(decVal).To(Equal(uint64(100)))
			})

			g.It("should panic if unable to decode", func() {
				Expect(func() {
					DecodeNumber([]byte("n10a"))
				}).To(Panic())
			})
		})

		g.Describe(".MayDecodeNumber", func() {
			g.It("should decode bytes value to 100", func() {
				decVal, err := MayDecodeNumber([]uint8{
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x64,
				})
				Expect(err).To(BeNil())
				Expect(decVal).To(Equal(uint64(100)))
			})

			g.It("should return error if unable to decode", func() {
				_, err := MayDecodeNumber([]byte("n10a"))
				Expect(err).ToNot(BeNil())
			})
		})

		g.Describe(".ToJSFriendlyMap", func() {

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

			g.It("should return expected output", func() {
				t1 := test1{
					Name: "fred",
					Desc: []byte("i love games"),
				}
				result := ToJSFriendlyMap(t1)
				Expect(result).To(Equal(map[string]interface{}{"Name": "fred",
					"Desc": "0x69206c6f76652067616d6573",
				}))

				t2 := test2{
					Age:    20,
					Others: t1,
				}
				result = ToJSFriendlyMap(t2)
				Expect(result.(map[string]interface{})["Others"]).To(Equal(map[string]interface{}{
					"Name": "fred",
					"Desc": "0x69206c6f76652067616d6573",
				}))

				t3 := test2{
					Age:    20,
					Others: t1,
					More:   []interface{}{t1},
				}
				result = ToJSFriendlyMap(t3)
				Expect(result.(map[string]interface{})["More"]).To(Equal([]interface{}{
					map[string]interface{}{"Name": "fred",
						"Desc": "0x69206c6f76652067616d6573",
					},
				}))

				t4 := test3{
					Sig: StrToHash("fred"),
				}
				result = ToJSFriendlyMap(t4)
				Expect(result).To(Equal(map[string]interface{}{"Sig": "0x6672656400000000000000000000000000000000000000000000000000000000"}))

				t5 := test4{
					Num: new(big.Int).SetInt64(10),
				}
				result = ToJSFriendlyMap(t5)
				Expect(result).To(Equal(map[string]interface{}{"Num": "0xa"}))

				t6 := test5{
					Num: EncodeNonce(10),
				}
				result = ToJSFriendlyMap(t6)
				Expect(result).To(Equal(map[string]interface{}{"Num": "0x000000000000000a"}))
			})
		})
	})
}
