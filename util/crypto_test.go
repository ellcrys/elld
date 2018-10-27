package util

import (
	"testing"

	. "github.com/ncodes/goblin"
	. "github.com/onsi/gomega"
)

func TestCrypto(t *testing.T) {
	g := Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("Crypto", func() {
		g.Describe(".Encrypt", func() {
			g.It("should return err='crypto/aes: invalid key size 12' when key size is less than 32 bytes", func() {
				msg := []byte("hello")
				key := []byte("not-32-bytes")
				_, err := Encrypt(msg, key)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("crypto/aes: invalid key size 12"))
			})

			g.It("should successfully encrypt", func() {
				msg := []byte("hello")
				key := []byte("abcdefghijklmnopqrstuvwxyzabcdef")
				enc, err := Encrypt(msg, key)
				Expect(err).To(BeNil())
				Expect(enc).ToNot(BeNil())
			})
		})

		g.Describe(".Decrypt", func() {
			g.It("should successfully decrypt", func() {
				msg := []byte("hello")
				key := []byte("abcdefghijklmnopqrstuvwxyzabcdef")
				enc, err := Encrypt(msg, key)
				Expect(err).To(BeNil())
				Expect(enc).ToNot(BeNil())

				dec, err := Decrypt(enc, key)
				Expect(err).To(BeNil())
				Expect(dec).To(Equal(msg))
			})
		})

		g.Describe(".HexToHash", func() {
			g.It("", func() {
				hash := StrToHash("something")
				hex := hash.HexStr()
				result, err := HexToHash(hex)
				Expect(err).To(BeNil())
				Expect(result.Equal(hash)).To(BeTrue())
			})
		})
	})
}
