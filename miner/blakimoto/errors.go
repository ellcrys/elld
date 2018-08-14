package blakimoto

import (
	"fmt"
)

var (
	// ErrUnknownParent indicates an unknown parent of a block
	ErrUnknownParent = fmt.Errorf("block's parent is unknown")
)
