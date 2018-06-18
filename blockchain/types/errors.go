package types

import (
	"fmt"
)

var (
	// ErrBlockNotFound means a block was not found
	ErrBlockNotFound = fmt.Errorf("block not found")
)
