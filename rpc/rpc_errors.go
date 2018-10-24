package rpc

import (
	"fmt"
)

var (
	// ErrMethodArgType creates an error about an invalid type
	// sent as an argument to a RPC method
	ErrMethodArgType = func(expectedType string) error {
		return fmt.Errorf("invalid argument type: expecting " + expectedType)
	}
)
