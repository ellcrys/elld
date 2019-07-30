package crypto

import (
	"crypto/ecdsa"

	"github.com/ellcrys/elld/ltcsuite/ltcd/chaincfg"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = FDescribe("Secp256k1", func() {
	Describe(".NewSecp256k1", func() {
		When("different seeds are used to derive two keys", func() {
			var sk, sk2 *Secp256k1Key

			BeforeEach(func() {
				var err error
				sk, err = NewSecp256k1(nil, true, false)
				Expect(err).To(BeNil())
				sk2, err = NewSecp256k1(nil, true, false)
				Expect(err).To(BeNil())
			})

			Specify("that both keys are not the same", func() {
				Expect(sk.sk.D.Cmp(sk2.sk.D)).ToNot(BeZero())
			})
		})

		When("same seed is used to derive two keys", func() {
			var sk, sk2 *Secp256k1Key

			BeforeEach(func() {
				var err error
				seed := int64(1234)
				sk, err = NewSecp256k1(&seed, true, false)
				Expect(err).To(BeNil())
				sk2, err = NewSecp256k1(&seed, true, false)
				Expect(err).To(BeNil())
			})

			Specify("that both keys are the same", func() {
				Expect(sk.sk.D.Cmp(sk2.sk.D)).To(BeZero())
			})
		})
	})

	Describe(".PrivateKey", func() {
		It("should return the ecdsa.PrivateKey", func() {
			sk, _ := NewSecp256k1(nil, true, false)
			Expect(sk.PrivateKey()).ToNot(BeAssignableToTypeOf(ecdsa.PrivateKey{}))
		})
	})

	Describe(".WIF", func() {
		It("should return the ltcutil.WIF instance", func() {
			sk, _ := NewSecp256k1(nil, true, false)
			wif, err := sk.WIF()
			Expect(err).To(BeNil())
			Expect(wif.IsForNet(&chaincfg.TestNet4Params)).To(BeTrue())
			Expect(wif.CompressPubKey).To(BeFalse())
		})
	})

	Describe(".Addr", func() {
		It("should return an address", func() {
			sk, _ := NewSecp256k1(nil, true, false)
			addr := sk.Addr()
			Expect(addr).ToNot(Equal(""))
		})
	})

	Describe(".ForTestnet", func() {
		It("should return true", func() {
			sk, _ := NewSecp256k1(nil, true, false)
			testnet := sk.ForTestnet()
			Expect(testnet).To(BeTrue())
		})

		It("should return false", func() {
			sk, _ := NewSecp256k1(nil, false, false)
			testnet := sk.ForTestnet()
			Expect(testnet).To(BeFalse())
		})
	})
})
