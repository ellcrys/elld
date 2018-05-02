package wire

import (
	"github.com/ellcrys/druid/crypto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Transaction", func() {

	Describe(".Bytes", func() {

		It("should return bytes = [48, 39, 2, 1, 1, 2, 1, 1, 12, 12, 115, 111, 109, 101, 95, 97, 100, 100, 114, 101, 115, 115, 12, 12, 115, 111, 109, 101, 95, 112, 117, 98, 95, 107, 101, 121, 2, 1, 0, 19, 0]", func() {
			tx := &Transaction{Type: 1, Nonce: 1, To: "some_address", SenderPubKey: "some_pub_key"}
			bs := tx.Bytes()
			Expect(bs).ToNot(BeEmpty())
			Expect(bs).To(Equal([]byte{48, 39, 2, 1, 1, 2, 1, 1, 12, 12, 115, 111, 109, 101, 95, 97, 100, 100, 114, 101, 115, 115, 12, 12, 115, 111, 109, 101, 95, 112, 117, 98, 95, 107, 101, 121, 2, 1, 0, 19, 0}))
		})
	})

	Describe(".TxSign", func() {
		It("should return error = 'nil tx' when tx is nil", func() {
			_, err := TxSign(nil, "private key")
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("nil tx"))
		})

		It("should return error = when tx is nil", func() {
			seed := int64(1)
			a, _ := crypto.NewAddress(&seed)
			tx := &Transaction{Type: 1, Nonce: 1, To: "some_address", SenderPubKey: "some_pub_key"}
			sig, err := TxSign(tx, a.PrivKey().Base58())
			Expect(err).To(BeNil())
			Expect(sig).ToNot(BeEmpty())
			Expect(sig).To(Equal([]byte{244, 14, 92, 225, 125, 58, 24, 155, 26, 38, 189, 239, 224, 252, 82, 128, 16, 188, 149, 0, 119, 192, 36, 50, 247, 79, 190, 133, 197, 62, 89, 221, 239, 93, 0, 93, 138, 108, 139, 171, 171, 57, 1, 250, 2, 202, 125, 180, 88, 195, 105, 103, 20, 159, 48, 184, 61, 9, 210, 141, 120, 47, 143, 6}))
		})
	})

	Describe(".TxVerify", func() {
		It("should return error = 'nil tx' when nil is passed", func() {
			err := TxVerify(nil)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("nil tx"))
		})

		It("should return err = 'sender public not set' when sender public key is not set on the transaction", func() {
			tx := &Transaction{Type: 1, Nonce: 1, To: "some_address"}
			err := TxVerify(tx)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("sender public not set"))
		})

		It("should return err = 'signature not set' when signature is unset", func() {
			tx := &Transaction{Type: 1, Nonce: 1, SenderPubKey: "pub key", To: "some_address"}
			err := TxVerify(tx)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("signature not set"))
		})

		It("should verify successfully", func() {
			seed := int64(1)
			a, _ := crypto.NewAddress(&seed)
			tx := &Transaction{Type: 1, Nonce: 1, To: "some_address", SenderPubKey: a.PubKey().Base58()}
			sig, err := TxSign(tx, a.PrivKey().Base58())
			Expect(err).To(BeNil())
			Expect(sig).ToNot(BeEmpty())

			tx.Sig = sig
			err = TxVerify(tx)
			Expect(err).To(BeNil())
		})

		It("should return err = 'verify failed' when verification failed", func() {
			seed := int64(1)
			a, _ := crypto.NewAddress(&seed)
			tx := &Transaction{Type: 1, Nonce: 1, To: "some_address", SenderPubKey: a.PubKey().Base58()}
			sig, err := TxSign(tx, a.PrivKey().Base58())
			Expect(err).To(BeNil())
			Expect(sig).ToNot(BeEmpty())

			tx.Sig = sig
			tx.To = "altered_address"
			err = TxVerify(tx)
			Expect(err).ToNot(BeNil())
			Expect(err).To(Equal(crypto.ErrVerifyFailed))
		})
	})

	Describe(".Hash", func() {
		It("should return", func() {
			seed := int64(1)
			a, _ := crypto.NewAddress(&seed)
			tx := &Transaction{Type: 1, Nonce: 1, To: "some_address", SenderPubKey: a.PubKey().Base58()}
			Expect(tx.Hash()).To(Equal([]byte{227, 139, 22, 55, 152, 201, 142, 46, 247, 17, 34, 50, 97, 163, 255, 214, 25, 33, 109, 222, 80, 78, 113, 12, 192, 170, 95, 175, 85, 136, 57, 221}))
		})
	})
})
