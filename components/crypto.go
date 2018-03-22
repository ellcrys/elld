package components

import (
	"io"
	crypto "github.com/libp2p/go-libp2p-crypto"
)

// GenerateKeyPair generates private and public keys
func GenerateKeyPair(r io.Reader) (crypto.PrivKey, crypto.PubKey, error) {
	return crypto.GenerateEd25519Key(r)
}
