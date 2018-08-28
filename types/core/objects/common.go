package objects

import (
	"fmt"

	"github.com/vmihailenco/msgpack"
)

func fieldError(field, err string) error {
	return fmt.Errorf(fmt.Sprintf("field:%s, error:%s", field, err))
}

// getBytes return the serialized representation of fields
func getBytes(fields []interface{}) []byte {

	result, err := msgpack.Marshal(fields)
	if err != nil {
		panic(err)
	}

	return result
}
