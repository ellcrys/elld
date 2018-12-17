package params

import (
	"math/big"
	"time"
)

// Gossip parameters
var (
	// MaxGetBlockHashes is the max number of block headers to request
	// from a remote peer per request.
	MaxGetBlockHashes = int64(5)
)

// Monetary parameters
var (
	// Decimals is the number of coin decimal places
	Decimals = int32(18)
)

// Block parameters
var (
	// AllowedFutureBlockTime is the number of seconds
	// a block's timestamp can have beyond the current timestamp
	AllowedFutureBlockTime = 15 * time.Second

	// MaxBlockNonTxsSize is the maximum size
	// of the non-transactional data a block
	// can have (headers, signature, hash).
	MaxBlockNonTxsSize = int64(1024)

	// MaxBlockTxsSize is the maximum size of
	// transactions that can fit in a block
	MaxBlockTxsSize = int64(9998976)

	// MaximumExtraDataSize is the size of extra data a block can contain.
	MaximumExtraDataSize uint64 = 32

	// DifficultyBoundDivisor is the bound divisor of the difficulty,
	// used in the update calculations.
	DifficultyBoundDivisor = big.NewInt(2048)

	// DurationLimit is the decision boundary on the blocktime duration used to
	// determine whether difficulty should go up or not.
	DurationLimit = big.NewInt(60)

	// GenesisDifficulty is the difficulty of the Genesis block.
	GenesisDifficulty = big.NewInt(50000000)

	// MinimumDifficulty is the minimum that the difficulty may ever be.
	MinimumDifficulty = big.NewInt(100000)

	// MinimumDurationIncrease is the minimum percent increase
	// a block's time can be when compared to its parent's
	MinimumDurationIncrease = big.NewFloat(2)

	// SyncMinHeightDiff is the number of blocks
	// a remote peer can have more than the local
	// peer to trigger block sync with the peer
	SyncMinHeightDiff = uint64(3)
)

// Transaction parameters
var (
	// PoolCapacity is the max. number of transaction
	// that can be added to the transaction pool at
	// any given time.
	PoolCapacity = int64(10000)
)

// Engine parameters
var (
	// QueueProcessorInterval is the duration between each
	// attempt to process a job from the queues managed by
	// the block manager
	QueueProcessorInterval = 1 * time.Second
)
