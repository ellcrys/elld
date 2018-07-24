package validators

import "fmt"

func fieldError(field, err string) error {
	return fmt.Errorf(fmt.Sprintf("field:%s, error:%s", field, err))
}

func fieldErrorWithIndex(index int, field, err string) error {
	return fmt.Errorf(fmt.Sprintf("index:%d, field:%s, error:%s", index, field, err))
}
