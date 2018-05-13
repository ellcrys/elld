package miner

import (
	"math/big"

	"github.com/ellcrys/druid/params"
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

	//next := new(big.Int).Add(parentBlockNumber, big1)
	//switch {
	// case config.IsByzantium(next):
	// 	return calcDifficultyByzantium(time, parent)
	// case config.IsHomestead(next):
	// 	return calcDifficultyHomestead(time, parent)
	// default:
	// 	return calcDifficultyFrontier(time, parent)
	// }

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

	//BigIntParentDifficulty, _ := new(big.Int).SetString(parent.Difficulty, 64)

	bigTime := new(big.Int).SetUint64(time)
	// bigParentTime := new(big.Int).Set(parent.Time)
	//bigParentTime, _ := new(big.Int).SetString(parent.Time, 10)

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
	y.Div(ParentDifficulty, params.DifficultyBoundDivisor)
	x.Mul(y, x)
	x.Add(ParentDifficulty, x)

	// minimum difficulty can ever be (before exponential factor)
	if x.Cmp(params.MinimumDifficulty) < 0 {
		x.Set(params.MinimumDifficulty)
	}
	// for the exponential factor

	// fmt.Println("<<>>", parent.Number)
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

// calcDifficultyByzantium is the difficulty adjustment algorithm. It returns
// the difficulty that a new block should have when created at time given the
// parent block's time and difficulty. The calculation uses the Byzantium rules.
// func calcDifficultyByzantium(time uint64, parent *ellBlock.FullBlock) *big.Int {
// 	// https://github.com/ethereum/EIPs/issues/100.
// 	// algorithm:
// 	// diff = (parent_diff +
// 	//         (parent_diff / 2048 * max((2 if len(parent.uncles) else 1) - ((timestamp - parent.timestamp) // 9), -99))
// 	//        ) + 2^(periodCount - 2)

// 	bigTime := new(big.Int).SetUint64(time)
// 	bigParentTime := new(big.Int).Set(parent.Time)

// 	// holds intermediate values to make the algo easier to read & audit
// 	x := new(big.Int)
// 	y := new(big.Int)

// 	// (2 if len(parent_uncles) else 1) - (block_timestamp - parent_timestamp) // 9
// 	x.Sub(bigTime, bigParentTime)
// 	x.Div(x, big9)
// 	if parent.UncleHash == types.EmptyUncleHash {
// 		x.Sub(big1, x)
// 	} else {
// 		x.Sub(big2, x)
// 	}

// 	// max((2 if len(parent_uncles) else 1) - (block_timestamp - parent_timestamp) // 9, -99)
// 	if x.Cmp(bigMinus99) < 0 {
// 		x.Set(bigMinus99)
// 	}
// 	// parent_diff + (parent_diff / 2048 * max((2 if len(parent.uncles) else 1) - ((timestamp - parent.timestamp) // 9), -99))
// 	y.Div(parent.Difficulty, params.DifficultyBoundDivisor)
// 	x.Mul(y, x)
// 	x.Add(parent.Difficulty, x)

// 	// minimum difficulty can ever be (before exponential factor)
// 	if x.Cmp(params.MinimumDifficulty) < 0 {
// 		x.Set(params.MinimumDifficulty)
// 	}
// 	// calculate a fake block number for the ice-age delay:
// 	//   https://github.com/ethereum/EIPs/pull/669
// 	//   fake_block_number = min(0, block.number - 3_000_000
// 	fakeBlockNumber := new(big.Int)
// 	if parent.Number.Cmp(big2999999) >= 0 {
// 		fakeBlockNumber = fakeBlockNumber.Sub(parent.Number, big2999999) // Note, parent is 1 less than the actual block number
// 	}
// 	// for the exponential factor
// 	periodCount := fakeBlockNumber
// 	periodCount.Div(periodCount, expDiffPeriod)

// 	// the exponential factor, commonly referred to as "the bomb"
// 	// diff = diff + 2^(periodCount - 2)
// 	if periodCount.Cmp(big1) > 0 {
// 		y.Sub(periodCount, big2)
// 		y.Exp(big2, y, nil)
// 		x.Add(x, y)
// 	}
// 	return x
// }

// // calcDifficultyFrontier is the difficulty adjustment algorithm. It returns the
// // difficulty that a new block should have when created at time given the parent
// // block's time and difficulty. The calculation uses the Frontier rules.
// func calcDifficultyFrontier(time uint64, parent *ellBlock.FullBlock) *big.Int {
// 	diff := new(big.Int)
// 	adjust := new(big.Int).Div(parent.Difficulty, params.DifficultyBoundDivisor)
// 	bigTime := new(big.Int)
// 	bigParentTime := new(big.Int)

// 	bigTime.SetUint64(time)
// 	bigParentTime.Set(parent.Time)

// 	if bigTime.Sub(bigTime, bigParentTime).Cmp(params.DurationLimit) < 0 {
// 		diff.Add(parent.Difficulty, adjust)
// 	} else {
// 		diff.Sub(parent.Difficulty, adjust)
// 	}
// 	if diff.Cmp(params.MinimumDifficulty) < 0 {
// 		diff.Set(params.MinimumDifficulty)
// 	}

// 	periodCount := new(big.Int).Add(parent.Number, big1)
// 	periodCount.Div(periodCount, expDiffPeriod)
// 	if periodCount.Cmp(big1) > 0 {
// 		// diff = diff + 2^(periodCount - 2)
// 		expDiff := periodCount.Sub(periodCount, big2)
// 		expDiff.Exp(big2, expDiff, nil)
// 		diff.Add(diff, expDiff)
// 		diff = math.BigMax(diff, params.MinimumDifficulty)
// 	}
// 	return diff
// }
