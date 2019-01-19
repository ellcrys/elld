package types

// Account package error codes
const (
	// ErrCodeListAccountFailed for failure to list account
	ErrCodeListAccountFailed = 30000
	// ErrCodeAccountNotFound for missing account
	ErrCodeAccountNotFound = 30001
)

// RPC package error codes
const (
	// ErrCodeInvalidAuthParams for invalid authorization parameter
	ErrCodeInvalidAuthParams = 40000
	// ErrCodeInvalidAuthCredentials for invalid authorization credentials
	ErrCodeInvalidAuthCredentials = 40001
)

// Blockchain package error codes
const (
	// ErrCodeQueryFailed for when a query fails
	ErrCodeQueryFailed = 50000
	// ErrCodeBlockNotFound for when a block is not found
	ErrCodeBlockNotFound = 50001
	// ErrCodeTransactionNotFound for when a transaction is not found
	ErrCodeTransactionNotFound = 50001
	// ErrCodeBlockQuery for non-specific block query errors
	ErrCodeBlockQuery = 50002
)

// General error codes
const (
	// ErrCodeUnexpectedArgType for when an argument type is invalid
	ErrCodeUnexpectedArgType = 60000
	// ErrCodeQueryParamError for when a query parameter is invalid
	ErrCodeQueryParamError = 60001
	// ErrValueDecodeFailed for when decoding a value failed
	ErrValueDecodeFailed = 600
)

// Engine package error codes
const (
	// ErrCodeAddress for when the is an issue with an address
	ErrCodeAddress = 70000
	// ErrCodeNodeConnectFailure for when there is an issue connecting to a node
	ErrCodeNodeConnectFailure = 70001
	// ErrCodeTxFailed when transaction failed
	ErrCodeTxFailed = 70002
)
