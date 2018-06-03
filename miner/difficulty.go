package miner

import (
	"math/big"

	"github.com/ellcrys/druid/constants"
)

// Some weird constants to avoid constant memory allocs for them.
var (
	expDiffPeriod = big.NewInt(100000)
	big1          = big.NewInt(1)
	big2          = big.NewInt(2)
	big9          = big.NewInt(9)
	big10         = big.NewInt(10)
	bigMinus99    = big.NewInt(-99)
	big2999999    = big.NewInt(2999999)
)

// CalcDifficulty is the difficulty adjustment algorithm. It returns
// the difficulty that a new block should have when created at time
// given the parent block's time and difficulty.
func (ethash *Ethash) CalcDifficulty(DifficultyVersion string, time uint64, parentBlockTime *big.Int, ParentDifficulty *big.Int, parentBlockNumber *big.Int) *big.Int {
	return CalcDifficulty(DifficultyVersion, time, parentBlockTime, ParentDifficulty, parentBlockNumber)
}

// CalcDifficulty is the difficulty adjustment algorithm. It returns
// the difficulty that a new block should have when created at time
// given the parent block's time and difficulty.
func CalcDifficulty(DifficultyVersion string, time uint64, parentBlockTime *big.Int, ParentDifficulty *big.Int, parentBlockNumber *big.Int) *big.Int {
	return calcDifficultyHomestead(time, parentBlockTime, ParentDifficulty)
}

// calcDifficultyHomestead is the difficulty adjustment algorithm. It returns
// the difficulty that a new block should have when created at time given the
// parent block's time and difficulty. The calculation uses the Homestead rules.
func calcDifficultyHomestead(time uint64, parentBlockTime *big.Int, ParentDifficulty *big.Int) *big.Int {
	// https://github.com/ethereum/EIPs/blob/master/EIPS/eip-2.md
	// algorithm:
	// diff = (parent_diff +
	//         (parent_diff / 2048 * max(1 - (block_timestamp - parent_timestamp) // 10, -99))
	//        ) + 2^(periodCount - 2)

	// convert time to *big.Int
	bigTime := new(big.Int).SetUint64(time)

	// holds intermediate values to make the algo easier to read & audit
	x := new(big.Int)
	y := new(big.Int)

	// 1 - (block_timestamp - parent_timestamp) // 10
	x.Sub(bigTime, parentBlockTime)
	x.Div(x, big10)
	x.Sub(big1, x)

	// max(1 - (block_timestamp - parent_timestamp) // 10, -99)
	if x.Cmp(bigMinus99) < 0 {
		x.Set(bigMinus99)
	}
	// (parent_diff + parent_diff // 2048 * max(1 - (block_timestamp - parent_timestamp) // 10, -99))
	y.Div(ParentDifficulty, constants.DifficultyBoundDivisor)
	x.Mul(y, x)
	x.Add(ParentDifficulty, x)

	// minimum difficulty can ever be (before exponential factor)
	if x.Cmp(constants.MinimumDifficulty) < 0 {
		x.Set(constants.MinimumDifficulty)
	}

	// for the exponential factor
	periodCount := new(big.Int).Add(new(big.Int).Set(ParentDifficulty), big1)
	periodCount.Div(periodCount, expDiffPeriod)

	// the exponential factor, commonly referred to as "the bomb"
	// diff = diff + 2^(periodCount - 2)
	if periodCount.Cmp(big1) > 0 {
		y.Sub(periodCount, big2)
		y.Exp(big2, y, nil)
		x.Add(x, y)
	}
	return x
}
