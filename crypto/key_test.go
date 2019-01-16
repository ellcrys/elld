package crypto

import (
	"github.com/btcsuite/btcutil/base58"
	"github.com/ellcrys/elld/util"
	"github.com/ellcrys/go-libp2p-crypto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Key", func() {
	Describe(".NewKey", func() {
		When("seeds are '1'", func() {
			It("multiple calls should return same private keys", func() {
				seed := int64(1)
				a1, err := NewKey(&seed)
				Expect(err).To(BeNil())
				a2, err := NewKey(&seed)
				Expect(a1).To(Equal(a2))
			})
		})

		When("seeds are nil", func() {
			It("should return random key on each call", func() {
				a1, err := NewKey(nil)
				Expect(err).To(BeNil())
				a2, err := NewKey(nil)
				Expect(err).To(BeNil())
				Expect(a1).NotTo(Equal(a2))
			})
		})

		When("seeds are different", func() {
			It("multiple calls should not return same private keys", func() {
				seed := int64(1)
				a1, err := NewKey(&seed)
				Expect(err).To(BeNil())
				seed = int64(2)
				a2, err := NewKey(&seed)
				Expect(a1).NotTo(Equal(a2))
			})
		})
	})

	Describe(".PeerID", func() {
		It("should return 12D3KooWHHzSeKaY8xuZVzkLbKFfvNgPPeKhFBGrMbNzbm5akpqu", func() {
			seed := int64(1)
			a1, err := NewKey(&seed)
			Expect(err).To(BeNil())
			Expect(a1.PeerID()).To(Equal("12D3KooWHHzSeKaY8xuZVzkLbKFfvNgPPeKhFBGrMbNzbm5akpqu"))
		})
	})

	Describe(".Addr", func() {
		It("should return 'eLPbkwui7eymQFLo8GRa7jgTJrhrXS6a8c'", func() {
			seed := int64(1)
			a, err := NewKey(&seed)
			Expect(err).To(BeNil())
			addr := a.Addr()
			Expect(addr).To(Equal(util.String("eLPbkwui7eymQFLo8GRa7jgTJrhrXS6a8c")))
		})
	})

	Describe("PubKey.Bytes", func() {
		It("should return err.Error('public key is nil')", func() {
			a := PubKey{}
			_, err := a.Bytes()
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(Equal("public key is nil"))
		})

		It("should return []byte{111, 21, 129, 112, 155, 183, 177, 239, 3, 13, 33, 13, 177, 142, 59, 11, 161, 199, 118, 251, 166, 93, 140, 218, 173, 5, 65, 81, 66, 209, 137, 248}", func() {
			seed := int64(1)
			a, err := NewKey(&seed)
			bs, err := a.PubKey().Bytes()
			Expect(err).To(BeNil())
			expected := []byte{111, 21, 129, 112, 155, 183, 177, 239, 3, 13, 33, 13, 177, 142, 59, 11, 161, 199, 118, 251, 166, 93, 140, 218, 173, 5, 65, 81, 66, 209, 137, 248}
			Expect(bs).To(Equal(expected))
		})
	})

	Describe("PubKey.Base58", func() {
		It("should return 48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC", func() {
			seed := int64(1)
			a, err := NewKey(&seed)
			Expect(err).To(BeNil())
			hx := a.PubKey().Base58()
			Expect(hx).To(Equal("48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC"))
		})
	})

	Describe("Priv.Bytes", func() {
		It("should return err.Error('private key is nil')", func() {
			a := PrivKey{}
			_, err := a.Bytes()
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(Equal("private key is nil"))
		})

		It("should return []byte{82, 253, 252, 7, 33, 130, 101, 79, 22, 63, 95, 15, 154, 98, 29, 114, 149, 102, 199, 77, 16, 3, 124, 77, 123, 187, 4, 7, 209, 226, 198, 73, 111, 21, 129, 112, 155, 183, 177, 239, 3, 13, 33, 13, 177, 142, 59, 11, 161, 199, 118, 251, 166, 93, 140, 218, 173, 5, 65, 81, 66, 209, 137, 248}", func() {
			seed := int64(1)
			a, err := NewKey(&seed)
			bs, err := a.PrivKey().Bytes()
			Expect(err).To(BeNil())
			expected := []byte{82, 253, 252, 7, 33, 130, 101, 79, 22, 63, 95, 15, 154, 98, 29, 114, 149, 102, 199, 77, 16, 3, 124, 77, 123, 187, 4, 7, 209, 226, 198, 73, 111, 21, 129, 112, 155, 183, 177, 239, 3, 13, 33, 13, 177, 142, 59, 11, 161, 199, 118, 251, 166, 93, 140, 218, 173, 5, 65, 81, 66, 209, 137, 248}
			Expect(bs).To(Equal(expected))
		})
	})

	Describe("Priv.Base58", func() {
		It("should return wU7ckbRBWevtkoT9QoET1adGCsABPRtyDx5T9EHZ4paP78EQ1w5sFM2sZg87fm1N2Np586c98GkYwywvtgy9d2gEpWbsbU", func() {
			seed := int64(1)
			a, err := NewKey(&seed)
			Expect(err).To(BeNil())
			hx := a.PrivKey().Base58()
			Expect(hx).To(Equal("wU7ckbRBWevtkoT9QoET1adGCsABPRtyDx5T9EHZ4paP78EQ1w5sFM2sZg87fm1N2Np586c98GkYwywvtgy9d2gEpWbsbU"))
		})
	})

	Describe("Priv.Sign", func() {
		It("should sign the word 'hello'", func() {
			seed := int64(1)
			a, err := NewKey(&seed)
			Expect(err).To(BeNil())
			sig, err := a.PrivKey().Sign([]byte("hello"))
			Expect(err).To(BeNil())
			Expect(sig).ToNot(BeEmpty())
			Expect(sig).To(Equal([]byte{158, 13, 68, 26, 41, 83, 26, 181, 43, 77, 192, 150, 115, 117, 175, 47, 207, 26, 118, 217, 101, 179, 49, 206, 126, 203, 37, 152, 3, 68, 75, 1, 141, 65, 141, 7, 87, 247, 160, 35, 94, 34, 137, 101, 185, 75, 228, 85, 240, 182, 166, 71, 94, 88, 208, 108, 189, 55, 174, 220, 119, 184, 128, 15}))
		})
	})

	Describe("Priv.Marshal", func() {
		It("should marshal and unmarshal correctly", func() {
			seed := int64(1)
			a, err := NewKey(&seed)
			Expect(err).To(BeNil())

			bs, err := a.PrivKey().Marshal()
			Expect(err).To(BeNil())
			Expect(bs).ToNot(BeEmpty())

			_, err = crypto.UnmarshalPrivateKey(bs)
			Expect(err).To(BeNil())
		})
	})

	Describe("Pub.Verify", func() {

		It("should return false when signature is incorrect", func() {
			seed := int64(1)
			a, err := NewKey(&seed)
			Expect(err).To(BeNil())
			sig, err := a.PrivKey().Sign([]byte("hello"))
			Expect(err).To(BeNil())

			valid, err := a.PubKey().Verify([]byte("hello friend"), sig)
			Expect(err).To(BeNil())
			Expect(valid).To(BeFalse())
		})

		It("should return true when the signature is correct", func() {
			seed := int64(1)
			a, err := NewKey(&seed)
			Expect(err).To(BeNil())
			sig, err := a.PrivKey().Sign([]byte("hello"))
			Expect(err).To(BeNil())
			Expect(sig).ToNot(BeEmpty())
			Expect(sig).To(Equal([]byte{158, 13, 68, 26, 41, 83, 26, 181, 43, 77, 192, 150, 115, 117, 175, 47, 207, 26, 118, 217, 101, 179, 49, 206, 126, 203, 37, 152, 3, 68, 75, 1, 141, 65, 141, 7, 87, 247, 160, 35, 94, 34, 137, 101, 185, 75, 228, 85, 240, 182, 166, 71, 94, 88, 208, 108, 189, 55, 174, 220, 119, 184, 128, 15}))

			valid, err := a.PubKey().Verify([]byte("hello"), sig)
			Expect(err).To(BeNil())
			Expect(valid).To(BeTrue())
		})
	})

	Describe(".IsValidAddr", func() {
		It("should return error.Error(empty address)", func() {
			err := IsValidAddr("")
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("empty address"))
		})

		It("should return err.Error(checksum error)", func() {
			err := IsValidAddr("hh23887dhhw88su")
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("checksum error"))
		})

		It("should return err.Error(invalid version)", func() {
			err := IsValidAddr("E1juuqo9XEfKhGHSwExMxGry54h4JzoRkr")
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("invalid version"))
		})

		It("should return err.Error(invalid version)", func() {
			invalidAddress := base58.CheckEncode([]byte{2, 3, 5}, AddressVersion)
			err := IsValidAddr(invalidAddress)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("invalid address size"))
		})

		It("should return nil", func() {
			err := IsValidAddr("eDFPdimzRqfFKetEMSmsSLTLHCLSniZQwD")
			Expect(err).To(BeNil())
		})
	})

	Describe(".IsValidPubKey", func() {
		It("should return error.Error(empty pub key)", func() {
			err := IsValidPubKey("")
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("empty pub key"))
		})

		It("should return err.Error(checksum error)", func() {
			err := IsValidPubKey("hh23887dhhw88su")
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("checksum error"))
		})

		It("should return err.Error(invalid version)", func() {
			err := IsValidPubKey("E1juuqo9XEfKhGHSwExMxGry54h4JzoRkr")
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("invalid version"))
		})

		It("should return nil", func() {
			err := IsValidPubKey("48s9G48LD5eo5YMjJWmRjPaoDZJRNTuiscHMov6zDGMEqUg4vbG")
			Expect(err).To(BeNil())
		})
	})

	Describe(".FromBase58PubKey", func() {
		It("should return err.Error(checksum error)", func() {
			_, err := PubKeyFromBase58("hh23887dhhw88su")
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("checksum error"))
		})

		It("should return err.Error(invalid version)", func() {
			_, err := PubKeyFromBase58("E1juuqo9XEfKhGHSwExMxGry54h4JzoRkr")
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("invalid version"))
		})

		It("should return err = nil", func() {
			pk, err := PubKeyFromBase58("48s9G48LD5eo5YMjJWmRjPaoDZJRNTuiscHMov6zDGMEqUg4vbG")
			Expect(err).To(BeNil())
			Expect(pk).ToNot(BeNil())
		})
	})

	Describe(".IsValidPrivKey", func() {
		It("should return error.Error(empty priv key)", func() {
			err := IsValidPrivKey("")
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("empty priv key"))
		})

		It("should return err.Error(checksum error)", func() {
			err := IsValidPrivKey("hh23887dhhw88su")
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("checksum error"))
		})

		It("should return err.Error(invalid version)", func() {
			err := IsValidPrivKey("E1juuqo9XEfKhGHSwExMxGry54h4JzoRkr")
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("invalid version"))
		})

		It("should return nil", func() {
			err := IsValidPrivKey("waS1jBBgdyYgpNtTjKbt6MZbDTYweLtzkNxueyyEc6ss33kPG58VcJNmpDK82BwuX8LAoqZuBCdaoXbxHPM99k8HFvqueW")
			Expect(err).To(BeNil())
		})
	})

	Describe(".PrivKeyFromBase58", func() {
		It("should return err.Error(checksum error)", func() {
			_, err := PrivKeyFromBase58("hh23887dhhw88su")
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("checksum error"))
		})

		It("should return err.Error(invalid version)", func() {
			_, err := PrivKeyFromBase58("E1juuqo9XEfKhGHSwExMxGry54h4JzoRkr")
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("invalid version"))
		})

		It("should return err = nil", func() {
			pk, err := PrivKeyFromBase58("waS1jBBgdyYgpNtTjKbt6MZbDTYweLtzkNxueyyEc6ss33kPG58VcJNmpDK82BwuX8LAoqZuBCdaoXbxHPM99k8HFvqueW")
			Expect(err).To(BeNil())
			Expect(pk).ToNot(BeNil())
		})
	})

})
