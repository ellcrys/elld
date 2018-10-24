package util

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Crypto", func() {

	Describe(".Encrypt", func() {

		It("should return err='crypto/aes: invalid key size 12' when key size is less than 32 bytes", func() {
			msg := []byte("hello")
			key := []byte("not-32-bytes")
			_, err := Encrypt(msg, key)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("crypto/aes: invalid key size 12"))
		})

		It("should successfully encrypt", func() {
			msg := []byte("hello")
			key := []byte("abcdefghijklmnopqrstuvwxyzabcdef")
			enc, err := Encrypt(msg, key)
			Expect(err).To(BeNil())
			Expect(enc).ToNot(BeNil())
		})
	})

	Describe(".Decrypt", func() {
		It("should successfully decrypt", func() {
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

	Describe(".HexToHash", func() {
		It("", func() {
			hash := StrToHash("something")
			hex := hash.HexStr()
			result, err := HexToHash(hex)
			Expect(err).To(BeNil())
			Expect(result.Equal(hash)).To(BeTrue())
		})
	})

})
