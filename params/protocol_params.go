package params

import "math/big"

var (
	// MaximumExtraDataSize is the size of extra data a block can contain.
	MaximumExtraDataSize uint64 = 32

	// DifficultyBoundDivisor is the bound divisor of the difficulty,
	// used in the update calculations.
	DifficultyBoundDivisor = big.NewInt(2048)

	// DurationLimit is the decision boundary on the blocktime duration used to
	// determine whether difficulty should go up or not.
	DurationLimit = big.NewInt(13)

	// GenesisDifficulty is the difficulty of the Genesis block.
	GenesisDifficulty = big.NewInt(131072)

	// MinimumDifficulty is the minimum that the difficulty may ever be.
	MinimumDifficulty = big.NewInt(100000)
)
