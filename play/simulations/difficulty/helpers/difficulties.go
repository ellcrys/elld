package helpers

import (
	gomath "math"
	"math/big"

	"github.com/ellcrys/elld/util/math"
)

var (
	big1          = big.NewInt(1)
	big0          = big.NewInt(0)
	expDiffPeriod = big.NewInt(3)
	big2          = big.NewInt(2)
	big100F       = big.NewFloat(100)

	// DifficultyBoundDivisor is the bound divisor of the difficulty,
	// used in the update calculations.
	DifficultyBoundDivisor = big.NewInt(2048)

	// DurationLimit is the decision boundary on the blocktime duration used to
	// determine whether difficulty should go up or not.
	DurationLimit = big.NewInt(120)

	// MinimumDifficulty is the minimum that the difficulty may ever be.
	MinimumDifficulty = big.NewInt(100000)

	// MinimumDurationIncrease is the minimum percent increase
	// of a new block time in relation to its parent block time.
	// Formula: (((DurationLimit+1)-DurationLimit)/DurationLimit) * 100
	MinimumDurationIncrease = big.NewFloat(0.8)
)

// RoundFloat rounds x
func RoundFloat(x *big.Float) {
	n, _ := x.Float64()
	nr := gomath.Round(n)
	x.SetFloat64(nr)
}

func CalcDifficultyInception(time uint64, parent *Block) *big.Int {

	diff := new(big.Int)
	bigTime := new(big.Int).SetUint64(time)
	bigParentTime := new(big.Int).Set(new(big.Int).SetInt64(parent.Timestamp.Unix()))

	// Define the value to adjust difficulty by
	// when the block time is within or above the duration limit.
	adjust := new(big.Int).Div(parent.Difficulty, DifficultyBoundDivisor)

	// Calculate the time difference between the
	// new block time and parent block time. We'll
	// use this to determine whether on not to increase
	// or decrease difficulty
	blockTimeDiff := bigTime.Sub(bigTime, bigParentTime)

	// When block time difference is within the expected
	// block duration limit, we increasse difficulty
	if blockTimeDiff.Cmp(DurationLimit) < 0 {
		diff.Add(parent.Difficulty, adjust)
	}

	// When block time difference is equal or greater than
	// the expected block duration limit, we decrease difficulty
	if blockTimeDiff.Cmp(DurationLimit) >= 0 {

		// We need to determine the percentage increase of the
		// new block time compared to the duration limit
		durLimitF := new(big.Float).SetInt(DurationLimit)
		blockTimeF := new(big.Float).SetInt(bigTime)
		timeDiff := new(big.Float).Sub(blockTimeF, durLimitF)
		pctDiff := new(big.Float).Mul(new(big.Float).Quo(timeDiff, durLimitF), big100F)

		// If the percentage difference is below the allowed mimimum
		// reset to the minimum
		if pctDiff.Cmp(MinimumDurationIncrease) < 0 {
			pctDiff = new(big.Float).Set(MinimumDurationIncrease)
		}

		// Calculate the new adjustment based on time difference percentage
		pctDiff = pctDiff.Quo(pctDiff, big100F)
		timeDiffAdjust, _ := new(big.Float).Mul(pctDiff, new(big.Float).SetInt(adjust)).Int(nil)
		diff.Sub(parent.Difficulty, timeDiffAdjust)
	}

	// Ensure difficulty does not go below the required minimum
	if diff.Cmp(MinimumDifficulty) < 0 {
		diff.Set(MinimumDifficulty)
	}

	// Here, we exponentially increase the difficulty
	// when the block time is within expected duration.
	// Otherwise, exponentially reduce the difficulty
	periodCount := new(big.Int).Add(new(big.Int).SetUint64(parent.Number), big1)
	periodCount.Div(periodCount, expDiffPeriod)
	if periodCount.Cmp(big1) > 0 {
		expDiff := periodCount.Sub(periodCount, big2)
		expDiff.Exp(big2, expDiff, nil)

		if blockTimeDiff.Cmp(DurationLimit) < 0 {
			diff.Add(diff, expDiff)
		} else {
			diff.Sub(diff, expDiff)
		}

		diff = math.BigMax(diff, MinimumDifficulty)
	}

	return diff
}
