package addressmgr

import (
	"encoding/hex"
	"fmt"
	mrand "math/rand"
	"time"

	"golang.org/x/crypto/ripemd160"

	"github.com/btcsuite/btcutil/base58"
	mh "github.com/multiformats/go-multihash"

	peer "github.com/libp2p/go-libp2p-peer"

	crypto "github.com/ellcrys/go-libp2p-crypto"
	"golang.org/x/crypto/sha3"
)

// AddressVersion is the base58 encode version adopted
var AddressVersion byte = 92

// PublicKeyVersion is the base58 encode version adopted for public keys
var PublicKeyVersion byte = 93

// PrivateKeyVersion is the base58 encode version adopted for private keys
var PrivateKeyVersion byte = 94

// Address represents an address
type Address struct {
	privKey *PrivKey
}

// PubKey represents a public key
type PubKey struct {
	pubKey crypto.PubKey
}

// PrivKey represents a private key
type PrivKey struct {
	privKey crypto.PrivKey
}

// NewAddress creates a new address
func NewAddress(seed *int64) (*Address, error) {

	var r = mrand.New(mrand.NewSource(time.Now().UnixNano()))

	if seed != nil {
		r = mrand.New(mrand.NewSource(*seed))
	}

	priv, _, err := crypto.GenerateEd25519Key(r)
	if err != nil {
		return nil, err
	}

	addr := &Address{
		privKey: &PrivKey{
			privKey: priv,
		},
	}

	return addr, nil
}

func idFromPublicKey(pk crypto.PubKey) (string, error) {
	b, err := pk.Bytes()
	if err != nil {
		return "", err
	}
	var alg uint64 = mh.SHA2_256
	if len(b) <= peer.MaxInlineKeyLength {
		alg = mh.ID
	}
	hash, _ := mh.Sum(b, alg, -1)
	return hash.B58String(), nil
}

// PeerID returns the IPFS compatible peer ID for the corresponding public key
func (a *Address) PeerID() string {
	pid, _ := idFromPublicKey(a.PubKey().pubKey)
	return pid
}

// Addr returns the address corresponding to the public key
func (a *Address) Addr() string {
	pkHex := a.PubKey().Hex()
	pubSha256 := sha3.Sum256([]byte(pkHex))
	r := ripemd160.New()
	r.Write(pubSha256[:])
	addr := r.Sum(nil)
	return base58.CheckEncode(addr, AddressVersion)
}

// PubKey returns the public key
func (a *Address) PubKey() *PubKey {
	return &PubKey{
		pubKey: a.privKey.privKey.GetPublic(),
	}
}

// PrivKey returns the private key
func (a *Address) PrivKey() *PrivKey {
	return a.privKey
}

// Bytes returns the byte equivalent of the public key
func (p *PubKey) Bytes() ([]byte, error) {
	if p.pubKey == nil {
		return nil, fmt.Errorf("public key is nil")
	}
	bs := p.pubKey.(*crypto.Ed25519PublicKey).BytesU()
	return bs[:], nil
}

// Hex returns the public key in hex encoding
func (p *PubKey) Hex() string {
	bs, _ := p.Bytes()
	return hex.EncodeToString(bs[:])
}

// Base58 returns the public key in base58 encoding
func (p *PubKey) Base58() string {
	bs, _ := p.Bytes()
	return base58.CheckEncode(bs[:], PublicKeyVersion)
}

// Bytes returns the byte equivalent of the public key
func (p *PrivKey) Bytes() ([]byte, error) {
	if p.privKey == nil {
		return nil, fmt.Errorf("private key is nil")
	}
	bs := p.privKey.(*crypto.Ed25519PrivateKey).BytesU()
	return bs[:], nil
}

// Base58 returns the public key in base58 encoding
func (p *PrivKey) Base58() string {
	bs, _ := p.Bytes()
	return base58.CheckEncode(bs, PrivateKeyVersion)
}

// IsValidAddr checks whether an address is valid
func IsValidAddr(addr string) error {
	if addr == "" {
		return fmt.Errorf("empty address")
	}

	_, v, err := base58.CheckDecode(addr)
	if err != nil {
		return err
	}

	if v != AddressVersion {
		return fmt.Errorf("invalid version")
	}

	return nil
}
