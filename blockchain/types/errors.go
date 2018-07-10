package types

import (
	"fmt"
)

var (
	// ErrBlockNotFound means a block was not found
	ErrBlockNotFound = fmt.Errorf("block not found")

	// ErrMetadataNotFound means the blockchain metadata was not found
	ErrMetadataNotFound = fmt.Errorf("metadata not found")

	// ErrChainAlreadyKnown means a chain is already known to the blockchain manager
	ErrChainAlreadyKnown = fmt.Errorf("chain already known")

	// ErrBlockExists means block exists
	ErrBlockExists = fmt.Errorf("block already exists")

	// ErrOrphanBlock means a block is orphaned
	ErrOrphanBlock = fmt.Errorf("orphan block")
)
