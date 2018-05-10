package crypto

import (
	"fmt"
)

var (
	// ErrVerifyFailed means signature verification failed
	ErrVerifyFailed = fmt.Errorf("verify failed")
)
