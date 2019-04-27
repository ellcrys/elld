package params

import "fmt"

var (
	// ErrRejected represents a rejected error
	ErrRejected = fmt.Errorf("rejected")

	// ErrMiningWithEphemeralKey represent an error about
	// mining with an ephemeral node key
	ErrMiningWithEphemeralKey = fmt.Errorf("Cannot mine with an ephemeral key. Please Provide an " +
		"account using '--account' flag.")

	// ErrBranchParentNotInMainChain means a branch's
	// parent block is not on the main chain
	ErrBranchParentNotInMainChain = fmt.Errorf("parent block does not exist on the main chain")
)
