package util

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	r "math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"

	"github.com/mitchellh/mapstructure"

	"github.com/vmihailenco/msgpack"

	"github.com/fatih/structs"

	"github.com/shopspring/decimal"
)

const (
	letterBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

// Big0 represents a zero value big.Int
var Big0 = new(big.Int).SetInt64(0)

func init() {
	r.Seed(time.Now().UnixNano())
}

// String represents a custom string
type String string

// Bytes returns the bytes equivalent of the string
func (s String) Bytes() []byte {
	return []byte(s)
}

// Equal check whether s and o are the same
func (s String) Equal(o String) bool {
	return s.String() == o.String()
}

func (s String) String() string {
	return string(s)
}

// SS returns a short version of String() with the middle
// characters truncated when length is at least 32
func (s String) SS() string {
	if len(s) >= 32 {
		return fmt.Sprintf("%s...%s", string(s)[0:10], string(s)[len(s)-10:])
	}
	return string(s)
}

// Decimal returns the decimal representation of the string.
// Panics if string failed to be converted to decimal.
func (s String) Decimal() decimal.Decimal {
	return StrToDec(s.String())
}

// IsDecimal checks whether the string
// can be converted to decimal
func (s String) IsDecimal() bool {
	defer func() {
		recover()
	}()
	s.Decimal()
	return true
}

// ObjectToBytes returns msgpack encoded
// representation of an object
func ObjectToBytes(s interface{}) []byte {
	b, _ := msgpack.Marshal(s)
	return b
}

// BytesToObject decodes bytes produced
// by BytesToObject to the given dest object
func BytesToObject(bs []byte, dest interface{}) error {
	return msgpack.Unmarshal(bs, dest)
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

// NonZeroOrDefIn64 checks if v is 0 so it returns def, otherwise returns v
func NonZeroOrDefIn64(v int64, def int64) int64 {
	if v == 0 {
		return def
	}
	return v
}

// StrToDec converts a numeric string to decimal.
// Panics if val could not be converted to decimal.
func StrToDec(val string) decimal.Decimal {
	d, err := decimal.NewFromString(val)
	if err != nil {
		panic(err)
	}
	return d
}

// IsPathOk checks if a path exist and whether
// there are no permission errors
func IsPathOk(path string) bool {
	_, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}

// IsFileOk checks if a path exists and it is a file
func IsFileOk(path string) bool {
	s, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return !s.IsDir()
}

// Int64ToHex converts an Int64 value to hex string.
// The resulting hex is prefixed by '0x'
func Int64ToHex(intVal int64) string {
	intValStr := strconv.FormatInt(intVal, 10)
	return "0x" + hex.EncodeToString([]byte(intValStr))
}

// HexToInt64 attempts to convert an hex string to Int64.
// Expects the hex string to begin with '0x'.
func HexToInt64(hexVal string) (int64, error) {
	hexStr, err := hex.DecodeString(hexVal[2:])
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(string(hexStr), 10, 64)
}

// StrToHex converts a string to hex. T
// The resulting hex is prefixed by '0x'
func StrToHex(str string) string {
	return "0x" + hex.EncodeToString([]byte(str))
}

// HexToStr decodes an hex string to string.
// Expects hexStr to begin with '0x'
func HexToStr(hexStr string) (string, error) {
	bs, err := hex.DecodeString(hexStr[2:])
	if err != nil {
		return "", err
	}
	return string(bs), nil
}

// SerializeMsg serializes an object using msgpack.
// Panics if an error is encountered
func SerializeMsg(o interface{}) []byte {
	bs, err := msgpack.Marshal(o)
	if err != nil {
		panic(err)
	}
	return bs
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

// MustFromHex is like FromHex except it panics if an error occurs
func MustFromHex(hexValue string) []byte {
	v, err := FromHex(hexValue)
	if err != nil {
		panic(err)
	}
	return v
}

// LogAlert wraps pp.Println with a read [Alert] prefix.
func LogAlert(format string, args ...interface{}) {
	fmt.Println(color.RedString("[Alert] "+format, args...))
}

// StructToMap returns a map containing fields from the s.
// Map fields are named after their json tags on the struct
func StructToMap(s interface{}) map[string]interface{} {
	_s := structs.New(s)
	_s.TagName = "json"
	return _s.Map()
}

// GetPtrAddr takes a pointer and returns the address
func GetPtrAddr(ptrAddr interface{}) *big.Int {
	ptrAddrInt, ok := new(big.Int).SetString(fmt.Sprintf("%d", &ptrAddr), 10)
	if !ok {
		panic("could not convert pointer address to big.Int")
	}
	return ptrAddrInt
}

// MapDecode decodes a map to a struct.
// It uses mapstructure.Decode internally but
// with 'json' TagName.
func MapDecode(m interface{}, rawVal interface{}) error {
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Metadata: nil,
		Result:   rawVal,
		TagName:  "json",
	})
	if err != nil {
		return err
	}

	return decoder.Decode(m)
}

// EncodeNumber serializes a number to BigEndian
func EncodeNumber(n uint64) []byte {
	var b = make([]byte, 8)
	binary.BigEndian.PutUint64(b, n)
	return b
}

// DecodeNumber deserialize a number from BigEndian
func DecodeNumber(encNum []byte) uint64 {
	return binary.BigEndian.Uint64(encNum)
}

// MayDecodeNumber is like DecodeNumber but returns
// an error instead of panicking
func MayDecodeNumber(encNum []byte) (r uint64, err error) {
	defer func() {
		if rcv, ok := recover().(error); ok {
			err = rcv
		}
	}()
	r = DecodeNumber(encNum)
	return
}

// BlockNonce is a 64-bit hash which proves (combined with the
// mix-hash) that a sufficient amount of computation has been carried
// out on a block.
type BlockNonce [8]byte

// EmptyBlockNonce is a BlockNonce with no values
var EmptyBlockNonce = BlockNonce([8]byte{})

// EncodeNonce converts the given integer to a block nonce.
func EncodeNonce(i uint64) BlockNonce {
	var n BlockNonce
	binary.BigEndian.PutUint64(n[:], i)
	return n
}

// Uint64 returns the integer value of a block nonce.
func (n BlockNonce) Uint64() uint64 {
	return binary.BigEndian.Uint64(n[:])
}

// MarshalText encodes n as a hex string with 0x prefix.
func (n BlockNonce) MarshalText() string {
	return ToHex(n[:])
}

// ToJSFriendlyMap takes a struct and converts
// selected types to values that are compatible in the
// JS environment. It returns a map and will panic
// if obj is not a map/struct.
func ToJSFriendlyMap(obj interface{}) interface{} {

	if obj == nil {
		return obj
	}

	var m map[string]interface{}

	// if not struct, we assume it is a map
	if structs.IsStruct(obj) {
		s := structs.New(obj)
		s.TagName = "json"
		m = s.Map()
	} else {
		m = obj.(map[string]interface{})
	}

	for k, v := range m {
		switch _v := v.(type) {
		case BlockNonce:
			m[k] = ToHex(_v[:])
		case Hash:
			m[k] = _v.HexStr()
		case *big.Int, int8, int, int64, uint64, []byte:
			m[k] = fmt.Sprintf("0x%x", _v)
		case map[string]interface{}:
			m[k] = ToJSFriendlyMap(_v)
		case []interface{}:
			for i, item := range _v {
				_v[i] = ToJSFriendlyMap(item)
			}
		}
	}

	return m
}
