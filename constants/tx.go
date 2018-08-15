package constants

import (
	"github.com/shopspring/decimal"
)

var (
	// BalanceTxMinimumFee is the minimum transaction fee for balance transactions
	BalanceTxMinimumFee = decimal.NewFromFloat(0.01)
)
