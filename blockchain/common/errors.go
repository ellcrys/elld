package common

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

	// ErrBlockExists means a block exists
	ErrBlockExists = fmt.Errorf("block already exists")

	// ErrBlockRejected means a block has been rejected
	ErrBlockRejected = fmt.Errorf("rejected block")

	// ErrOrphanBlock means a block is orphaned
	ErrOrphanBlock = fmt.Errorf("orphan block")

	// ErrVeryStaleBlock means a block is stable and has a height not equal to the current tip
	ErrVeryStaleBlock = fmt.Errorf("very stale block")

	// ErrBlockStateRootInvalid means the state root on a block header does is not valid after
	// transactions are executed.
	ErrBlockStateRootInvalid = fmt.Errorf("block state root is not valid")

	// ErrBlockFailedValidation means a block failed validation
	ErrBlockFailedValidation = fmt.Errorf("block failed validation")
)
