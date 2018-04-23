package addressmgr

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Addressmgr", func() {

	Describe(".NewAddress", func() {
		When("seed is 1", func() {
			It("multiple calls should return same private keys", func() {
				seed := int64(1)
				a1, err := NewAddress(&seed)
				Expect(err).To(BeNil())
				a2, err := NewAddress(&seed)
				Expect(a1).To(Equal(a2))
			})
		})

		When("with different seeds", func() {
			It("multiple calls should return same private keys", func() {
				seed := int64(1)
				a1, err := NewAddress(&seed)
				Expect(err).To(BeNil())
				seed = int64(2)
				a2, err := NewAddress(&seed)
				Expect(a1).NotTo(Equal(a2))
			})
		})
	})

	Describe(".PeerID", func() {
		It("should return 12D3KooWHHzSeKaY8xuZVzkLbKFfvNgPPeKhFBGrMbNzbm5akpqu", func() {
			seed := int64(1)
			a1, err := NewAddress(&seed)
			Expect(err).To(BeNil())
			Expect(a1.PeerID()).To(Equal("12D3KooWHHzSeKaY8xuZVzkLbKFfvNgPPeKhFBGrMbNzbm5akpqu"))
		})
	})

	Describe(".Addr", func() {
		It("should return 'eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad'", func() {
			seed := int64(1)
			a, err := NewAddress(&seed)
			Expect(err).To(BeNil())
			addr := a.Addr()
			Expect(addr).To(Equal("eGzzf1HtQL7M9Eh792iGHTvb6fsnnPipad"))
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
			a, err := NewAddress(&seed)
			bs, err := a.PubKey().Bytes()
			Expect(err).To(BeNil())
			expected := []byte{111, 21, 129, 112, 155, 183, 177, 239, 3, 13, 33, 13, 177, 142, 59, 11, 161, 199, 118, 251, 166, 93, 140, 218, 173, 5, 65, 81, 66, 209, 137, 248}
			Expect(bs).To(Equal(expected))
		})
	})

	Describe("PubKey.Base58", func() {
		It("should return 48d9u6L7tWpSVYmTE4zBDChMUasjP5pvoXE7kPw5HbJnXRnZBNC", func() {
			seed := int64(1)
			a, err := NewAddress(&seed)
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
			a, err := NewAddress(&seed)
			bs, err := a.PrivKey().Bytes()
			Expect(err).To(BeNil())
			expected := []byte{82, 253, 252, 7, 33, 130, 101, 79, 22, 63, 95, 15, 154, 98, 29, 114, 149, 102, 199, 77, 16, 3, 124, 77, 123, 187, 4, 7, 209, 226, 198, 73, 111, 21, 129, 112, 155, 183, 177, 239, 3, 13, 33, 13, 177, 142, 59, 11, 161, 199, 118, 251, 166, 93, 140, 218, 173, 5, 65, 81, 66, 209, 137, 248}
			Expect(bs).To(Equal(expected))
		})
	})

	Describe("Priv.Base58", func() {
		It("should return wU7ckbRBWevtkoT9QoET1adGCsABPRtyDx5T9EHZ4paP78EQ1w5sFM2sZg87fm1N2Np586c98GkYwywvtgy9d2gEpWbsbU", func() {
			seed := int64(1)
			a, err := NewAddress(&seed)
			Expect(err).To(BeNil())
			hx := a.PrivKey().Base58()
			Expect(hx).To(Equal("wU7ckbRBWevtkoT9QoET1adGCsABPRtyDx5T9EHZ4paP78EQ1w5sFM2sZg87fm1N2Np586c98GkYwywvtgy9d2gEpWbsbU"))
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

		It("should return err.Error(checksum error)", func() {
			err := IsValidAddr("E1juuqo9XEfKhGHSwExMxGry54h4JzoRkr")
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("invalid version"))
		})

		It("should return nil", func() {
			err := IsValidAddr("eDFPdimzRqfFKetEMSmsSLTLHCLSniZQwD")
			Expect(err).To(BeNil())
		})
	})
})
