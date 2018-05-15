package rpc

var (
	errCodeAccountStoredAccount   = 0x1
	errCodeTransaction            = 0x100
	errCodeUnknownTransactionType = 0x101
)

// NewErrorResult creates a result describing an error
func NewErrorResult(result *Result, err string, errCode, status int) error {
	r := Result{}
	r.Error = err
	r.ErrCode = errCode
	r.Status = status
	*result = r
	return nil
}

// NewOKResult creates a result describing an error
func NewOKResult(result *Result, status int, data map[string]interface{}) error {
	r := Result{}
	r.Data = data
	r.Status = status
	*result = r
	return nil
}
