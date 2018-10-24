package blakimoto

import (
	"errors"
	"fmt"
)

var (
	// ErrUnknownParent indicates an unknown parent of a block
	ErrUnknownParent = fmt.Errorf("block's parent is unknown")

	// ErrFutureBlock is returned when a block's timestamp is in the future according
	// to the current node.
	ErrFutureBlock = errors.New("block in the future")

	// ErrInvalidNumber is returned if a block's number doesn't equal it's parent's
	// plus one.
	ErrInvalidNumber = errors.New("invalid block number")
)
