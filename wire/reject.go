package wire

import (
	"fmt"
)

var (

	// InvalidVersion means a node does not agree with a remote node version
	InvalidVersion = 0x01

	// TooManyAddresses means and `addr` message contained too many addresses
	TooManyAddresses = 0x10
)

var (
	// ErrRejected represents a rejected message error
	ErrRejected = fmt.Errorf("rejected")
)

// NewRejectMsg creates a reject message
func NewRejectMsg(msg string, code int32, reason string, extra []byte) *Reject {
	return &Reject{
		Message:   msg,
		Code:      code,
		Reason:    reason,
		ExtraData: extra,
	}
}
