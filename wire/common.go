package wire

// NewRejectMsg creates a reject message
func NewRejectMsg(msg string, code int32, reason string, extra []byte) *Reject {
	return &Reject{
		Message:   msg,
		Code:      code,
		Reason:    reason,
		ExtraData: extra,
	}
}
