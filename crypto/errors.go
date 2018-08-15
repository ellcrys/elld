package crypto

import (
	"fmt"
)

var (
	// ErrTxVerificationFailed means signature verification failed
	ErrTxVerificationFailed = fmt.Errorf("transaction verification failed")

	// ErrBlockVerificationFailed means signature verification failed
	ErrBlockVerificationFailed = fmt.Errorf("block verification failed")
)
