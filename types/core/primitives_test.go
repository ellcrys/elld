package core

import (
	"math/big"

	"github.com/ellcrys/elld/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Primitives", func() {
	Describe(".MapByteFieldsToHex", func() {

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
			Sig util.Hash
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
			result := MapFieldsToHex(t1)
			Expect(result).To(Equal(map[string]interface{}{"Name": "fred",
				"Desc": "0x69206c6f76652067616d6573",
			}))

			t2 := test2{
				Age:    20,
				Others: t1,
			}
			result = MapFieldsToHex(t2)
			Expect(result.(map[string]interface{})["Others"]).To(Equal(map[string]interface{}{
				"Name": "fred",
				"Desc": "0x69206c6f76652067616d6573",
			}))

			t3 := test2{
				Age:    20,
				Others: t1,
				More:   []interface{}{t1},
			}
			result = MapFieldsToHex(t3)
			Expect(result.(map[string]interface{})["More"]).To(Equal([]interface{}{
				map[string]interface{}{"Name": "fred",
					"Desc": "0x69206c6f76652067616d6573",
				},
			}))

			t4 := test3{
				Sig: util.StrToHash("fred"),
			}
			result = MapFieldsToHex(t4)
			Expect(result).To(Equal(map[string]interface{}{"Sig": "0x6672656400000000000000000000000000000000000000000000000000000000"}))

			t5 := test4{
				Num: new(big.Int).SetInt64(10),
			}
			result = MapFieldsToHex(t5)
			Expect(result).To(Equal(map[string]interface{}{"Num": "0xa"}))

			t6 := test5{
				Num: EncodeNonce(10),
			}
			result = MapFieldsToHex(t6)
			Expect(result).To(Equal(map[string]interface{}{"Num": "0x000000000000000a"}))
		})
	})
})
