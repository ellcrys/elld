package util

import (
	"encoding/json"
	"math/big"
	r "math/rand"
	"sort"
	"time"

	"github.com/ellcrys/druid/configdir"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func init() {
	r.Seed(time.Now().UnixNano())
}

// StructToBytes returns json encoded representation of a struct
func StructToBytes(s interface{}) []byte {
	b, _ := json.Marshal(s)
	return b
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

// LoadCfg loads the config file
func LoadCfg(cfgDirPath string) (*configdir.Config, error) {

	cfgDir, err := configdir.NewConfigDir(cfgDirPath)
	if err != nil {
		return nil, err
	}

	if err := cfgDir.Init(); err != nil {
		if err != nil {
			return nil, err
		}

		return nil, err
	}

	cfg, err := cfgDir.Load()
	if err != nil {
		return nil, err
	}

	cfg.SetConfigDir(cfgDir.Path())

	return cfg, nil
}
