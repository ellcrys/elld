package wire

import (
	"fmt"
	"time"

	"github.com/ellcrys/elld/crypto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Transaction", func() {

	Describe(".Bytes", func() {

		It("should return bytes = [48, 41, 2, 1, 1, 2, 1, 1, 12, 12, 115, 111, 109, 101, 95, 97, 100, 100, 114, 101, 115, 115, 12, 12, 115, 111, 109, 101, 95, 112, 117, 98, 95, 107, 101, 121, 12, 0, 12, 0, 2, 1, 0]", func() {
			tx := &Transaction{Type: 1, Nonce: 1, To: "some_address", SenderPubKey: "some_pub_key"}
			bs := tx.Bytes()
			Expect(bs).ToNot(BeEmpty())
			Expect(bs).To(Equal([]byte{48, 41, 2, 1, 1, 2, 1, 1, 12, 12, 115, 111, 109, 101, 95, 97, 100, 100, 114, 101, 115, 115, 12, 12, 115, 111, 109, 101, 95, 112, 117, 98, 95, 107, 101, 121, 12, 0, 12, 0, 2, 1, 0}))
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
			Expect(err).To(BeNil())
			Expect(sig).ToNot(BeEmpty())
			Expect(sig).To(Equal([]byte{131, 33, 128, 253, 155, 66, 21, 113, 42, 23, 114, 78, 5, 31, 15, 20, 203, 228, 36, 79, 207, 49, 248, 81, 24, 25, 138, 185, 69, 215, 52, 46, 120, 42, 207, 239, 111, 162, 242, 196, 218, 63, 79, 196, 180, 28, 146, 14, 71, 81, 62, 57, 242, 228, 113, 127, 104, 148, 138, 177, 220, 46, 74, 3}))
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
			Expect(err.Error()).To(Equal("field = senderPubKey, msg=sender public not set"))
		})

		It("should return err = 'signature not set' when signature is unset", func() {
			tx := &Transaction{Type: 1, Nonce: 1, SenderPubKey: "pub key", To: "some_address"}
			err := TxVerify(tx)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("field = sig, msg=signature not set"))
		})

		It("should verify successfully", func() {
			seed := int64(1)
			a, _ := crypto.NewKey(&seed)
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
			a, _ := crypto.NewKey(&seed)
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
		It("should return [35, 25, 58, 248, 4, 154, 18, 141, 250, 79, 195, 147, 216, 172, 66, 28, 119, 135, 234, 51, 111, 125, 163, 178, 177, 114, 247, 89, 141, 81, 111, 59]", func() {
			seed := int64(1)
			a, _ := crypto.NewKey(&seed)
			tx := &Transaction{Type: 1, Nonce: 1, To: "some_address", SenderPubKey: a.PubKey().Base58()}
			Expect(tx.Hash()).To(Equal([]byte{35, 25, 58, 248, 4, 154, 18, 141, 250, 79, 195, 147, 216, 172, 66, 28, 119, 135, 234, 51, 111, 125, 163, 178, 177, 114, 247, 89, 141, 81, 111, 59}))
		})
	})

	Describe(".ID", func() {
		It("should return '23193af8049a128dfa4fc393d8ac421c7787ea336f7da3b2b172f7598d516f3b'", func() {
			seed := int64(1)
			a, _ := crypto.NewKey(&seed)
			tx := &Transaction{Type: 1, Nonce: 1, To: "some_address", SenderPubKey: a.PubKey().Base58()}
			Expect(tx.ID()).To(Equal("23193af8049a128dfa4fc393d8ac421c7787ea336f7da3b2b172f7598d516f3b"))
		})
	})

	Describe(".TxValidate", func() {

		It("should include err = 'field = senderPubKey, msg=sender public key is required' when sender public key is not provided", func() {
			tx := &Transaction{}
			errs := TxValidate(tx)
			Expect(errs).ToNot(BeEmpty())
			Expect(errs).To(ContainElement(fmt.Errorf("field = senderPubKey, msg=sender public key is required")))
		})

		It("should include err = 'field = senderPubKey, msg=invalid format: version and/or checksum bytes missing' when sender public key is invalid", func() {
			tx := &Transaction{
				SenderPubKey: "48nCZsmoU7wvA3__invalid_fS8UhXQv4u43eny8qpT7ubdVxp3kus3eNZP",
			}
			errs := TxValidate(tx)
			Expect(errs).ToNot(BeEmpty())
			Expect(errs).To(ContainElement(fmt.Errorf("field = senderPubKey, msg=invalid format: version and/or checksum bytes missing")))
		})

		It("should include err = 'field = to, msg=recipient address is required' when recipient address is not provided", func() {
			tx := &Transaction{
				SenderPubKey: "48nCZsmoU7wvA3ULfS8UhXQv4u43eny8qpT7ubdVxp3kus3eNZH",
			}
			errs := TxValidate(tx)
			Expect(errs).ToNot(BeEmpty())
			Expect(errs).To(ContainElement(fmt.Errorf("field = to, msg=recipient address is required")))
		})

		It("should include err = 'field = to, msg=address is not valid' when recipient address is not provided", func() {
			tx := &Transaction{
				SenderPubKey: "48nCZsmoU7wvA3ULfS8UhXQv4u43eny8qpT7ubdVxp3kus3eNZH",
				To:           "e5a3zJReMgLJrNn4GsYcnKf1Qa6GQFimC4",
			}
			errs := TxValidate(tx)
			Expect(errs).ToNot(BeEmpty())
			Expect(errs).To(ContainElement(fmt.Errorf("field = to, msg=address is not valid")))
		})

		It("should include err = 'field = timestamp, msg=timestamp cannot be a future time' when timestamp is a future time", func() {
			tx := &Transaction{
				SenderPubKey: "48nCZsmoU7wvA3ULfS8UhXQv4u43eny8qpT7ubdVxp3kus3eNZH",
				Timestamp:    time.Now().Add(10 * time.Second).Unix(),
			}
			errs := TxValidate(tx)
			Expect(errs).ToNot(BeEmpty())
			Expect(errs).To(ContainElement(fmt.Errorf("field = timestamp, msg=timestamp cannot be a future time")))
		})

		It("should include err = 'field = timestamp, msg=timestamp cannot over 7 days ago'", func() {
			tx := &Transaction{
				SenderPubKey: "48nCZsmoU7wvA3ULfS8UhXQv4u43eny8qpT7ubdVxp3kus3eNZH",
				Timestamp:    time.Now().Add(-8 * 24 * time.Hour).Unix(),
			}
			errs := TxValidate(tx)
			Expect(errs).ToNot(BeEmpty())
			Expect(errs).To(ContainElement(fmt.Errorf("field = timestamp, msg=timestamp cannot over 7 days ago")))
		})

		It("should include err = 'field = sig, msg=signature is required'", func() {
			tx := &Transaction{
				SenderPubKey: "48nCZsmoU7wvA3ULfS8UhXQv4u43eny8qpT7ubdVxp3kus3eNZH",
				Timestamp:    time.Now().Unix(),
			}
			errs := TxValidate(tx)
			Expect(errs).ToNot(BeEmpty())
			Expect(errs).To(ContainElement(fmt.Errorf("field = sig, msg=signature is required")))
		})
	})
})
