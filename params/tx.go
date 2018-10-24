package params

import (
	"github.com/shopspring/decimal"
)

var (
	// FeePerByte is the amount to be paid
	// as fee for a single byte.
	FeePerByte = decimal.NewFromFloat(0.00001)

	// TxTTL is the number of days a transaction
	// can last for in the pool
	TxTTL = 7
)
