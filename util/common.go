package util

import (
	"encoding/json"
	r "math/rand"
	"time"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func init() {
	r.Seed(time.Now().UnixNano())
}

// StructToBytes returns json encoded representation of a struct
func StructToBytes(s interface{}) []byte {
	b, _ := json.Marshal(s)
	return b
}

// RandString gets random string of fixed length
func RandString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[r.Intn(len(letterBytes))]
	}
	return string(b)
}
