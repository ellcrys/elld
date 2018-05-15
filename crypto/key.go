package crypto

import (
	"encoding/hex"
	"fmt"
	mrand "math/rand"
	"time"

	"github.com/gogo/protobuf/proto"
	pb "github.com/libp2p/go-libp2p-crypto/pb"

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

// Key represents an address
type Key struct {
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

// NewKey creates a new key
func NewKey(seed *int64) (*Key, error) {

	var r = mrand.New(mrand.NewSource(time.Now().UnixNano()))

	if seed != nil {
		r = mrand.New(mrand.NewSource(*seed))
	}

	priv, _, err := crypto.GenerateEd25519Key(r)
	if err != nil {
		return nil, err
	}

	key := &Key{
		privKey: &PrivKey{
			privKey: priv,
		},
	}

	return key, nil
}

// NewKeyFromIntSeed is like NewKey but accepts seed of type Int and casts to Int64
// Error is not returned.
// Intended to be used in test
func NewKeyFromIntSeed(seed int) *Key {
	int64Seed := int64(seed)
	addr, _ := NewKey(&int64Seed)
	return addr
}

// NewKeyFromPrivKey creates a new address from a private key
func NewKeyFromPrivKey(sk *PrivKey) *Key {
	return &Key{
		privKey: sk,
	}
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
func (k *Key) PeerID() string {
	pid, _ := idFromPublicKey(k.PubKey().pubKey)
	return pid
}

// Addr returns the address corresponding to the public key
func (k *Key) Addr() string {
	pkHex := k.PubKey().Hex()
	pubSha256 := sha3.Sum256([]byte(pkHex))
	r := ripemd160.New()
	r.Write(pubSha256[:])
	addr := r.Sum(nil)
	return base58.CheckEncode(addr, AddressVersion)
}

// PubKey returns the public key
func (k *Key) PubKey() *PubKey {
	return &PubKey{
		pubKey: k.privKey.privKey.GetPublic(),
	}
}

// PrivKey returns the private key
func (k *Key) PrivKey() *PrivKey {
	return k.privKey
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

// Verify verifies a signature
func (p *PubKey) Verify(data, sig []byte) (bool, error) {
	return p.pubKey.Verify(data, sig)
}

// Bytes returns the byte equivalent of the public key
func (p *PrivKey) Bytes() ([]byte, error) {
	if p.privKey == nil {
		return nil, fmt.Errorf("private key is nil")
	}
	bs := p.privKey.(*crypto.Ed25519PrivateKey).BytesU()
	return bs[:], nil
}

// Marshal encodes the private key using protocol buffer
func (p *PrivKey) Marshal() ([]byte, error) {
	pbmes := new(pb.PrivateKey)
	typ := pb.KeyType_Ed25519
	pbmes.Type = &typ

	sk, _ := p.Bytes()
	pk := p.privKey.GetPublic().(*crypto.Ed25519PublicKey).BytesU()

	buf := make([]byte, 96)
	copy(buf, sk[:])
	copy(buf[64:], pk[:])
	pbmes.Data = buf
	return proto.Marshal(pbmes)
}

// Base58 returns the public key in base58 encoding
func (p *PrivKey) Base58() string {
	bs, _ := p.Bytes()
	return base58.CheckEncode(bs, PrivateKeyVersion)
}

// Sign signs a message
func (p *PrivKey) Sign(data []byte) ([]byte, error) {
	return p.privKey.Sign(data)
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

// IsValidPubKey checks whether a public key is valid
func IsValidPubKey(pubKey string) error {

	if pubKey == "" {
		return fmt.Errorf("empty pub key")
	}

	_, v, err := base58.CheckDecode(pubKey)
	if err != nil {
		return err
	}

	if v != PublicKeyVersion {
		return fmt.Errorf("invalid version")
	}

	return nil
}

// IsValidPrivKey checks whether a private key is valid
func IsValidPrivKey(privKey string) error {

	if privKey == "" {
		return fmt.Errorf("empty priv key")
	}

	_, v, err := base58.CheckDecode(privKey)
	if err != nil {
		return err
	}

	if v != PrivateKeyVersion {
		return fmt.Errorf("invalid version")
	}

	return nil
}

// PubKeyFromBase58 decodes a base58 encoded public key
func PubKeyFromBase58(pk string) (*PubKey, error) {

	if err := IsValidPubKey(pk); err != nil {
		return nil, err
	}

	decPubKey, _, _ := base58.CheckDecode(pk)
	pubKey, err := crypto.UnmarshalEd25519PublicKey(decPubKey)
	if err != nil {
		return nil, err
	}

	return &PubKey{
		pubKey: pubKey,
	}, nil
}

// PrivKeyFromBase58 decodes a base58 encoded private key
func PrivKeyFromBase58(pk string) (*PrivKey, error) {

	if err := IsValidPrivKey(pk); err != nil {
		return nil, err
	}

	decPrivKey, _, _ := base58.CheckDecode(pk)
	var sk [64]byte
	copy(sk[:], decPrivKey)
	privKey := crypto.Ed25519PrivateKeyFromPrivKey(sk)

	return &PrivKey{
		privKey: privKey,
	}, nil
}
