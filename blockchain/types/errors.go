package types

import (
	"fmt"
)

var (
	// ErrBlockNotFound means a block was not found
	ErrBlockNotFound = fmt.Errorf("block not found")

	// ErrMetadataNotFound means the blockchain metadata was not found
	ErrMetadataNotFound = fmt.Errorf("metadata not found")
)
