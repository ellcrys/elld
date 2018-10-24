package core

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

	// ErrChainNotFound means a chain does not exist
	ErrChainNotFound = fmt.Errorf("chain not found")

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

	// ErrAccountNotFound refers to a missing account
	ErrAccountNotFound = fmt.Errorf("account not found")

	// ErrBestChainUnknown means the best/main chain is yet to be determined
	ErrBestChainUnknown = fmt.Errorf("best chain unknown")

	// ErrTxNotFound means a transaction was not found
	ErrTxNotFound = fmt.Errorf("transaction not found")

	// ErrDecodeFailed means an attempt to decode data failed
	ErrDecodeFailed = func(msg string) error {
		if msg != "" {
			msg = ": " + msg
		}
		return fmt.Errorf("decode attempt failed%s", msg)
	}

	// ErrChainParentNotFound means a chain's parent was not found
	ErrChainParentNotFound = fmt.Errorf("chain parent not found")

	// ErrChainParentBlockNotFound means a chain's parent block was not found
	ErrChainParentBlockNotFound = fmt.Errorf("chain parent block not found")
)
