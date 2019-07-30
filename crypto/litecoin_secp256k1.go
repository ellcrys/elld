package crypto

import (
	"crypto/ecdsa"
	"crypto/rand"
	mrand "math/rand"

	"github.com/ellcrys/elld/ltcsuite/ltcd/chaincfg"

	"github.com/ellcrys/elld/ltcsuite/ltcd/btcec"
	"github.com/ellcrys/elld/ltcsuite/ltcutil"
)

// Secp256k1Key represents a secp256k1 elliptic curve key
type Secp256k1Key struct {
	sk         *ecdsa.PrivateKey
	testnet    bool
	compressed bool
}

// NewSecp256k1 creates an instance of Secp256k1Key.
// The seed is optional; If not provided, a random
// seed will be generated.
// The testnet argument indicates that the chain config to use
// when generating a WIF address should be for the testnet,
// otherwise the mainnet chain config is used. The compressed
// argument indicates that the WIF public key is to be serialized
// compressed or uncompressed.
func NewSecp256k1(seed *int64, testnet, compressed bool) (*Secp256k1Key, error) {
	var r = rand.Reader

	// Set the seed; If seed is provided,
	// use it as a source of randomness
	if seed != nil {
		r = mrand.New(mrand.NewSource(*seed))
	}

	// Generate a secp256k1 key
	sk, err := ecdsa.GenerateKey(btcec.S256(), r)
	if err != nil {
		return nil, err
	}

	return &Secp256k1Key{
		sk:         sk,
		testnet:    testnet,
		compressed: compressed,
	}, nil
}

// PrivateKey returns the ecdsa.PrivateKey instance
func (k *Secp256k1Key) PrivateKey() *ecdsa.PrivateKey {
	return k.sk
}

// WIF returns a ltcutil.WIF structure.
func (k *Secp256k1Key) WIF() (*ltcutil.WIF, error) {

	chainCfg := chaincfg.MainNetParams
	if k.testnet {
		chainCfg = chaincfg.TestNet4Params
	}

	wrappedSK := btcec.PrivateKey(*k.sk)
	wif, err := ltcutil.NewWIF(&wrappedSK, &chainCfg, k.compressed)
	if err != nil {
		return nil, err
	}

	return wif, nil
}

// Addr returns the Litecoin pay-to-pubkey address.
// Panics if unable to derive pay-to-pubkey address.
func (k *Secp256k1Key) Addr() string {
	wif, err := k.WIF()
	if err != nil {
		panic(err)
	}

	netCfg := &chaincfg.TestNet4Params
	if !wif.IsForNet(&chaincfg.TestNet4Params) {
		netCfg = &chaincfg.MainNetParams
	}

	pubKey, err := ltcutil.NewAddressPubKey(wif.SerializePubKey(), netCfg)
	if err != nil {
		panic(err)
	}

	return pubKey.EncodeAddress()
}

// ForTestnet checks whether the key is for the testnet
func (k *Secp256k1Key) ForTestnet() bool {
	return k.testnet
}
