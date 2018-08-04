package util

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	r "math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

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

// ObjectToBytes returns json encoded representation of an object
func ObjectToBytes(s interface{}) []byte {
	b, _ := json.Marshal(s)
	return b
}

// BytesToObject converts byte slice to an object
func BytesToObject(bs []byte, dest interface{}) error {
	return json.Unmarshal(bs, dest)
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

// BigIntWithMeta represents a big integer with an arbitrary meta object attached
type BigIntWithMeta struct {
	Int  *big.Int
	Meta interface{}
}

type byBigIntWithMeta []*BigIntWithMeta

func (s byBigIntWithMeta) Len() int {
	return len(s)
}

func (s byBigIntWithMeta) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s byBigIntWithMeta) Less(i, j int) bool {
	return s[i].Int.Cmp(s[j].Int) == -1
}

// AscOrderBigIntMeta sorts a slice of BigIntWithMeta in ascending order
func AscOrderBigIntMeta(values []*BigIntWithMeta) {
	sort.Sort(byBigIntWithMeta(values))
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
	var _hexValue string
	parts := strings.Split(hexValue, "0x")
	if len(parts) == 1 {
		_hexValue = parts[0]
	} else {
		_hexValue = parts[1]
	}
	v, err := hex.DecodeString(_hexValue)
	if err != nil {
		panic(err)
	}
	return v
}

// StructToMap returns a map containing fields from the s.
// Map fields are named after their json tags on the struct
func StructToMap(s interface{}) map[string]interface{} {
	_s := structs.New(s)
	_s.TagName = "json"
	return _s.Map()
}

// StrToDecimal returns decimal representation of a numeric string value
func StrToDecimal(v string) (decimal.Decimal, error) {
	if v == "" {
		v = "0"
	}
	return decimal.NewFromString(v)
}
