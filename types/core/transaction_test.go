package core

import (
	"testing"
	"time"

	"github.com/ellcrys/elld/crypto"
	"github.com/ellcrys/elld/util"
	. "github.com/ncodes/goblin"
	. "github.com/onsi/gomega"
)

func TestTransaction(t *testing.T) {
	g := Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("Transaction", func() {

		var address = crypto.NewKeyFromIntSeed(1)

		g.Describe(".NewTx", func() {
			g.It("should successfully create and sign a new transaction", func() {
				Expect(func() {
					NewTx(TxTypeBalance, 0, "recipient_addr", address, "10", "0.1", time.Now().Unix())
				}).ToNot(Panic())
			})
		})

		g.Describe(".Bytes", func() {

			g.It("should return expected bytes", func() {
				tx := &Transaction{Type: 1, Nonce: 1, To: "some_address", SenderPubKey: "some_pub_key"}
				bs := tx.GetBytesNoHashAndSig()
				expected := []byte{153, 1, 1, 172, 115, 111, 109, 101, 95, 97, 100, 100, 114, 101, 115, 115, 172, 115, 111, 109, 101, 95, 112, 117, 98, 95, 107, 101, 121, 160, 160, 160, 0, 192}
				Expect(bs).ToNot(BeEmpty())
				Expect(bs).To(Equal(expected))
			})
		})

		g.Describe(".TxSign", func() {
			g.It("should return error = 'nil tx' when tx is nil", func() {
				_, err := TxSign(nil, "private key")
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("nil tx"))
			})

			g.It("should expected signature", func() {
				seed := int64(1)
				a, _ := crypto.NewKey(&seed)
				tx := &Transaction{Type: 1, Nonce: 1, To: "some_address", SenderPubKey: "some_pub_key"}
				sig, err := TxSign(tx, a.PrivKey().Base58())
				expected := []byte{187, 27, 160, 206, 120, 165, 81, 92, 246, 142, 5, 12, 12, 67, 206, 178, 44, 84, 30, 208, 38, 158, 38, 54, 247, 198, 217, 212, 20, 181, 131, 219, 212, 102, 159, 199, 77, 32, 95, 143, 17, 36, 245, 16, 110, 87, 235, 61, 165, 111, 246, 46, 52, 223, 83, 113, 230, 236, 5, 110, 255, 154, 54, 12}
				Expect(err).To(BeNil())
				Expect(sig).ToNot(BeEmpty())
				Expect(sig).To(Equal(expected))
				Expect(sig).To(HaveLen(64))
			})
		})

		g.Describe(".TxVerify", func() {
			g.It("should return error = 'nil tx' when nil is passed", func() {
				err := TxVerify(nil)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("nil tx"))
			})

			g.It("should return err = 'sender public not set' when sender public key is not set on the transaction", func() {
				tx := &Transaction{Type: 1, Nonce: 1, To: "some_address"}
				err := TxVerify(tx)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("field:senderPubKey, error:sender public not set"))
			})

			g.It("should return err = 'signature not set' when signature is unset", func() {
				tx := &Transaction{Type: 1, Nonce: 1, SenderPubKey: "pub key", To: "some_address"}
				err := TxVerify(tx)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("field:sig, error:signature not set"))
			})

			g.It("should return error if senderPubKey is invalid", func() {
				tx := &Transaction{Type: 1, Nonce: 1, SenderPubKey: "pub key", To: "some_address", Sig: []byte("some_sig")}
				err := TxVerify(tx)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("field:senderPubKey, error:invalid format: version and/or checksum bytes missing"))
			})

			g.It("should verify successfully", func() {
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

			g.It("should return err = 'verify failed' when verification failed", func() {
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

		g.Describe(".ComputeHash", func() {
			g.It("should successfully return hash", func() {
				seed := int64(1)
				a, _ := crypto.NewKey(&seed)
				tx := &Transaction{Type: 1, Nonce: 1, To: "some_address", SenderPubKey: util.String(a.PubKey().Base58())}
				expected := util.BytesToHash([]byte{227, 245, 150, 232, 102, 136, 13, 163, 21, 114, 180, 15, 61, 191, 52, 60, 9, 50, 212, 196, 115, 235, 1, 181, 50, 212, 232, 44, 193, 175, 239, 61})
				Expect(tx.ComputeHash()).To(Equal(expected))
			})
		})

		g.Describe(".ID", func() {
			g.It("should return '0xe3f596e866880da31572b40f3dbf343c0932d4c473eb01b532d4e82cc1afef3d'", func() {
				seed := int64(1)
				a, _ := crypto.NewKey(&seed)
				tx := &Transaction{Type: 1, Nonce: 1, To: "some_address", SenderPubKey: util.String(a.PubKey().Base58())}
				Expect(tx.GetID()).To(Equal("0xe3f596e866880da31572b40f3dbf343c0932d4c473eb01b532d4e82cc1afef3d"))
			})
		})
	})
}
