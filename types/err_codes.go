package types

const (
	// Account package errors from 30000
	ErrCodeListAccountFailed = 30000
	ErrCodeAccountNotFound   = 30001

	// RPC package errors from 40000
	ErrCodeInvalidAuthParams      = 40000
	ErrCodeInvalidAuthCredentials = 40001

	// Blochchain package errors from 50000
	ErrCodeQueryFailed         = 50000
	ErrCodeBlockNotFound       = 50001
	ErrCodeTransactionNotFound = 50001

	// General
	ErrCodeUnexpectedArgType = 60000
	ErrCodeQueryParamError   = 60001

	// Engine package errors from 70000
	ErrCodeAddress     = 70000
	ErrCodeNodeConnect = 70001
	ErrCodeTxFailed    = 70002
)
