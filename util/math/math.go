package math

import "math/big"

// BigMax returns the larger of x or y.
func BigMax(x, y *big.Int) *big.Int {
	if x.Cmp(y) < 0 {
		return y
	}
	return x
}

// SetBigInt is like big.Set. It sets
// x to the y and returns x. Except,
// x will be returned if y is nil
func SetBigInt(x, y *big.Int) *big.Int {
	if y == nil {
		return x
	}
	return x.Set(y)
}
