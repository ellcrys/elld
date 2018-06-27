package wire

import (
	bytes "bytes"
	"encoding/asn1"
	"encoding/hex"
	"fmt"
	"math/big"
	r "math/rand"
	"strings"
)

const (
	HashLength    = 32
	letterBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

// Hash represents a 32 bytes digest
type Hash [HashLength]byte

// Bytes gets the byte representation of the underlying hash.
func (h Hash) Bytes() []byte { return h[:] }

// Big converts a hash to a big integer.
func (h Hash) Big() *big.Int { return new(big.Int).SetBytes(h[:]) }

// Equal checks whether h is equal to o
func (h Hash) Equal(o Hash) bool {
	return bytes.Equal(h.Bytes(), o.Bytes())
}

// ToHex encodes value to hex with a '0x' prefix
func ToHex(value []byte) string {
	return fmt.Sprintf("0x%s", hex.EncodeToString(value))
}

// FromHex decodes hex value to bytes. If hex value is prefixed
// with '0x' it is trimmed before the decode operation.
func FromHex(hexValue string) ([]byte, error) {
	var _hexValue string
	parts := strings.Split(hexValue, "0x")
	if len(parts) == 1 {
		_hexValue = parts[0]
	} else {
		_hexValue = parts[1]
	}
	return hex.DecodeString(_hexValue)
}

// RandString is like RandBytes but returns string
func RandString(n int) string {
	return string(RandBytes(n))
}

// RandBytes gets random string of fixed length
func RandBytes(n int) []byte {
	b := make([]byte, n)
	for i, cache, remain := n-1, r.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = r.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return b
}

func fieldError(field, err string) error {
	return fmt.Errorf(fmt.Sprintf("field:%s, error:%s", field, err))
}

// getBytes return the ASN.1 marshalled representation of fields
func getBytes(fields []interface{}) []byte {

	result, err := asn1.Marshal(fields)
	if err != nil {
		panic(err)
	}

	return result
}
