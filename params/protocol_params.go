package params

import "math/big"

var (
	DifficultyBoundDivisor = big.NewInt(2048) // The bound divisor of the difficulty, used in the update calculations.
	// GenesisDifficulty      = big.NewInt(131072) // Difficulty of the Genesis block.
	MinimumDifficulty = big.NewInt(131072) // The minimum that the difficulty may ever be.
	DurationLimit     = big.NewInt(13)     // The decision boundary on the blocktime duration used to determine whether difficulty should go up or not.
)
