package util

import "encoding/json"

// StructToBytes returns json encoded representation of a struct
func StructToBytes(s interface{}) []byte {
	b, _ := json.Marshal(s)
	return b
}
