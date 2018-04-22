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
		It("should return 'elf781f1e102d4a31a810e715ccc3da0de8bb773d9'", func() {
			seed := int64(1)
			a, err := NewAddress(&seed)
			Expect(err).To(BeNil())
			addr := a.Addr()
			Expect(addr).To(Equal("elf781f1e102d4a31a810e715ccc3da0de8bb773d9"))
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

	Describe("Priv.Hex", func() {
		It("should return ", func() {
			seed := int64(1)
			a, err := NewAddress(&seed)
			Expect(err).To(BeNil())
			hx := a.PrivKey().Hex()
			Expect(hx).To(Equal("52fdfc072182654f163f5f0f9a621d729566c74d10037c4d7bbb0407d1e2c6496f1581709bb7b1ef030d210db18e3b0ba1c776fba65d8cdaad05415142d189f8"))
		})
	})
})
