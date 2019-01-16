package core

import (
	"time"

	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Transaction", func() {

	var address = crypto.NewKeyFromIntSeed(1)

	Describe(".NewTx", func() {
		It("should successfully create and sign a new transaction", func() {
			Expect(func() {
				NewTx(TxTypeBalance, 0, "recipient_addr", address, "10", "0.1", time.Now().Unix())
			}).ToNot(Panic())
		})
	})

	Describe(".Bytes", func() {

		It("should return expected bytes", func() {
			tx := &Transaction{Type: 1, Nonce: 1, To: "some_address", SenderPubKey: "some_pub_key"}
			bs := tx.GetBytesNoHashAndSig()
			expected := []byte{153, 160, 160, 192, 1, 172, 115, 111, 109, 101, 95, 112, 117, 98, 95, 107, 101, 121, 0, 172, 115, 111, 109, 101, 95, 97, 100, 100, 114, 101, 115, 115, 1, 160}
			Expect(bs).ToNot(BeEmpty())
			Expect(bs).To(Equal(expected))
		})
	})

	Describe(".TxSign", func() {
		It("should return error = 'nil tx' when tx is nil", func() {
			_, err := TxSign(nil, "private key")
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("nil tx"))
		})

		It("should expected signature", func() {
			seed := int64(1)
			a, _ := crypto.NewKey(&seed)
			tx := &Transaction{Type: 1, Nonce: 1, To: "some_address", SenderPubKey: "some_pub_key"}
			sig, err := TxSign(tx, a.PrivKey().Base58())
			expected := []byte{182, 67, 103, 153, 34, 21, 108, 220, 73, 148, 192, 180, 118, 79, 211, 8, 78, 57, 14, 73, 52, 251, 118, 23, 239, 35, 109, 203, 12, 140, 219, 203, 232, 69, 231, 60, 176, 154, 236, 7, 127, 9, 27, 220, 178, 72, 14, 147, 90, 175, 179, 160, 185, 35, 219, 109, 248, 36, 224, 228, 105, 240, 32, 7}
			Expect(err).To(BeNil())
			Expect(sig).ToNot(BeEmpty())
			Expect(sig).To(Equal(expected))
			Expect(sig).To(HaveLen(64))
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
			Expect(err.Error()).To(Equal("field:senderPubKey, error:sender public not set"))
		})

		It("should return err = 'signature not set' when signature is unset", func() {
			tx := &Transaction{Type: 1, Nonce: 1, SenderPubKey: "pub key", To: "some_address"}
			err := TxVerify(tx)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("field:sig, error:signature not set"))
		})

		It("should return error if senderPubKey is invalid", func() {
			tx := &Transaction{Type: 1, Nonce: 1, SenderPubKey: "pub key", To: "some_address", Sig: []byte("some_sig")}
			err := TxVerify(tx)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("field:senderPubKey, error:invalid format: version and/or checksum bytes missing"))
		})

		It("should verify successfully", func() {
			seed := int64(1)
			a, _ := crypto.NewKey(&seed)
			tx := &Transaction{Type: 1, Nonce: 1, To: "some_address", SenderPubKey: util.String(a.PubKey().Base58())}
			sig, err := TxSign(tx, a.PrivKey().Base58())
			Expect(err).To(BeNil())
			Expect(sig).ToNot(BeEmpty())

			tx.Sig = sig
			err = TxVerify(tx)
			Expect(err).To(BeNil())
		})

		It("should return err = 'verify failed' when verification failed", func() {
			seed := int64(1)
			a, _ := crypto.NewKey(&seed)
			tx := &Transaction{Type: 1, Nonce: 1, To: "some_address", SenderPubKey: util.String(a.PubKey().Base58())}
			sig, err := TxSign(tx, a.PrivKey().Base58())
			Expect(err).To(BeNil())
			Expect(sig).ToNot(BeEmpty())

			tx.Sig = sig
			tx.To = "altered_address"
			err = TxVerify(tx)
			Expect(err).ToNot(BeNil())
			Expect(err).To(Equal(crypto.ErrTxVerificationFailed))
		})
	})

	Describe(".ComputeHash", func() {
		It("should successfully return hash", func() {
			seed := int64(1)
			a, _ := crypto.NewKey(&seed)
			tx := &Transaction{Type: 1, Nonce: 1, To: "some_address", SenderPubKey: util.String(a.PubKey().Base58())}
			expected := util.BytesToHash([]byte{97, 63, 219, 173, 209, 80, 110, 25, 172, 79, 166, 171, 69, 81, 152, 53, 253, 36, 35, 195, 76, 235, 151, 70, 74, 74, 203, 178, 38, 102, 187, 74})
			Expect(tx.ComputeHash()).To(Equal(expected))
		})
	})

	Describe(".ID", func() {
		It("should return expected transaction ID", func() {
			seed := int64(1)
			a, _ := crypto.NewKey(&seed)
			tx := &Transaction{Type: 1, Nonce: 1, To: "some_address", SenderPubKey: util.String(a.PubKey().Base58())}
			Expect(tx.GetID()).To(Equal("0x613fdbadd1506e19ac4fa6ab45519835fd2423c34ceb97464a4acbb22666bb4a"))
		})
	})

})
