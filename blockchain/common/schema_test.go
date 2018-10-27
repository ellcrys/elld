package common

import (
	"testing"

	. "github.com/ncodes/goblin"
	. "github.com/onsi/gomega"
)

func TestSchema(t *testing.T) {
	g := Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("Schema", func() {

		g.Describe(".MakeAccountKey", func() {
			g.It("should return expected key", func() {
				k := MakeKeyAccount(10, []byte("chainA"), []byte("some_addr"))
				Expect(k).To(Equal([]uint8{
					0x63, 0x3a, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x41, 0x3a, 0x61, 0x3a, 0x73, 0x6f, 0x6d, 0x65, 0x5f,
					0x61, 0x64, 0x64, 0x72, 0x40, 0x40, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0a,
				}))
			})
		})

		g.Describe(".QueryAccountKey", func() {
			g.It("should return expected key", func() {
				k := MakeQueryKeyAccount([]byte("chainA"), []byte("some_addr"))
				Expect(k).To(Equal([]uint8{
					0x63, 0x3a, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x41, 0x3a, 0x61, 0x3a, 0x73, 0x6f, 0x6d, 0x65, 0x5f,
					0x61, 0x64, 0x64, 0x72,
				}))
			})
		})

		g.Describe(".MakeBlockKey", func() {
			g.It("should return expected key", func() {
				k := MakeKeyBlock([]byte("chainA"), 10)
				Expect(k).To(Equal([]uint8{
					0x63, 0x3a, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x41, 0x3a, 0x62, 0x40, 0x40, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x0a,
				}))
			})
		})

		g.Describe(".MakeBlocksQueryKey", func() {
			g.It("should return expected key", func() {
				k := MakeQueryKeyBlocks([]byte("chainA"))
				Expect(k).To(Equal([]uint8{
					0x63, 0x3a, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x41, 0x3a, 0x62,
				}))
			})
		})

		g.Describe(".MakeChainKey", func() {
			g.It("should return expected key", func() {
				k := MakeKeyChain([]byte("chainA"))
				Expect(k).To(Equal([]uint8{
					0x69, 0x3a, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x41,
				}))
			})
		})

		g.Describe(".MakeChainsQueryKey", func() {
			g.It("should return expected key", func() {
				k := MakeQueryKeyChains()
				Expect(k).To(Equal([]uint8{
					0x69,
				}))
			})
		})

		g.Describe(".MakeTxKey", func() {
			g.It("should return expected key", func() {
				k := MakeKeyTransaction([]byte("chainA"), 221, []byte("tx123"))
				Expect(k).To(Equal([]uint8{
					0x63, 0x3a, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x41, 0x3a, 0x74, 0x3a, 0x74, 0x78, 0x31, 0x32, 0x33,
					0x40, 0x40, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xdd,
				}))
			})
		})

		g.Describe(".MakeTxQueryKey", func() {
			g.It("should return expected key", func() {
				k := MakeQueryKeyTransaction([]byte("chainA"), []byte("tx123"))
				Expect(k).To(Equal([]uint8{
					0x63, 0x3a, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x41, 0x3a, 0x74, 0x3a, 0x74, 0x78, 0x31, 0x32, 0x33,
				}))
			})
		})

		g.Describe(".MakeTreeKey", func() {
			g.It("should return expected key", func() {
				k := MakeTreeKey(10, TagAccount)
				Expect(k).To(Equal([]uint8{
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0a, 0x61,
				}))
			})
		})
	})
}
