// Copyright 2017 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package ethash

import (
	"bytes"
	"errors"
	"math/big"
	"runtime"
	"time"

	c "github.com/ellcrys/elld/blockchain/common"

	"github.com/ellcrys/elld/wire"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
)

// Ethash proof-of-work protocol constants.
var (
	allowedFutureBlockTime = 15 * time.Second // Max time from current time allowed for blocks, before they're considered future blocks
)

// Various error messages to mark blocks invalid. These should be private to
// prevent engine specific errors from being referenced in the remainder of the
// codebase, inherently breaking if the engine is swapped out. Please put common
// error types into the consensus package.
var (
	errLargeBlockTime    = errors.New("timestamp too big")
	errZeroBlockTime     = errors.New("timestamp equals parent's")
	errTooManyUncles     = errors.New("too many uncles")
	errDuplicateUncle    = errors.New("duplicate uncle")
	errUncleIsAncestor   = errors.New("uncle is ancestor")
	errDanglingUncle     = errors.New("uncle's parent is not ancestor")
	errInvalidDifficulty = errors.New("non-positive difficulty")
	errInvalidMixDigest  = errors.New("invalid mix digest")
	errInvalidPoW        = errors.New("invalid proof-of-work")
)

// verifyHeader checks whether a header conforms to the consensus rules of the
// stock Ethereum ethash engine.
// See YP section 4.3.4. "Block Header Validity"
func (ethash *Ethash) verifyHeader(chain consensus.ChainReader, header, parent *types.Header, uncle bool, seal bool) error {
	// Ensure that the header's extra-data section is of a reasonable size
	// if uint64(len(header.Extra)) > params.MaximumExtraDataSize {
	// 	return fmt.Errorf("extra-data too long: %d > %d", len(header.Extra), params.MaximumExtraDataSize)
	// }
	// // Verify the header's timestamp
	// if uncle {
	// 	if header.Time.Cmp(math.MaxBig256) > 0 {
	// 		return errLargeBlockTime
	// 	}
	// } else {
	// 	if header.Time.Cmp(big.NewInt(time.Now().Add(allowedFutureBlockTime).Unix())) > 0 {
	// 		return consensus.ErrFutureBlock
	// 	}
	// }
	// if header.Time.Cmp(parent.Time) <= 0 {
	// 	return errZeroBlockTime
	// }
	// // Verify the block's difficulty based in it's timestamp and parent's difficulty
	// expected := ethash.CalcDifficulty(chain, header.Time.Uint64(), parent)

	// if expected.Cmp(header.Difficulty) != 0 {
	// 	return fmt.Errorf("invalid difficulty: have %v, want %v", header.Difficulty, expected)
	// }
	// // Verify that the gas limit is <= 2^63-1
	// cap := uint64(0x7fffffffffffffff)
	// if header.GasLimit > cap {
	// 	return fmt.Errorf("invalid gasLimit: have %v, max %v", header.GasLimit, cap)
	// }
	// // Verify that the gasUsed is <= gasLimit
	// if header.GasUsed > header.GasLimit {
	// 	return fmt.Errorf("invalid gasUsed: have %d, gasLimit %d", header.GasUsed, header.GasLimit)
	// }

	// // Verify that the gas limit remains within allowed bounds
	// diff := int64(parent.GasLimit) - int64(header.GasLimit)
	// if diff < 0 {
	// 	diff *= -1
	// }
	// limit := parent.GasLimit / params.GasLimitBoundDivisor

	// if uint64(diff) >= limit || header.GasLimit < params.MinGasLimit {
	// 	return fmt.Errorf("invalid gas limit: have %d, want %d += %d", header.GasLimit, parent.GasLimit, limit)
	// }
	// // Verify that the block number is parent's +1
	// if diff := new(big.Int).Sub(header.Number, parent.Number); diff.Cmp(big.NewInt(1)) != 0 {
	// 	return consensus.ErrInvalidNumber
	// }
	// // Verify the engine specific seal securing the block
	// if seal {
	// 	if err := ethash.VerifySeal(chain, header); err != nil {
	// 		return err
	// 	}
	// }
	// // If all checks passed, validate any special fields for hard forks
	// if err := misc.VerifyDAOHeaderExtraData(chain.Config(), header); err != nil {
	// 	return err
	// }
	// if err := misc.VerifyForkHashes(chain.Config(), header, uncle); err != nil {
	// 	return err
	// }
	return nil
}

// CalcDifficulty is the difficulty adjustment algorithm. It returns
// the difficulty that a new block should have when created at time
// given the parent block's time and difficulty.
func (ethash *Ethash) CalcDifficulty(chain c.ChainReader, time uint64, parent *wire.Header) *big.Int {
	return CalcDifficulty(time, parent)
}

// CalcDifficulty is the difficulty adjustment algorithm. It returns
// the difficulty that a new block should have when created at time
// given the parent block's time and difficulty.
func CalcDifficulty(time uint64, parent *wire.Header) *big.Int {
	return calcDifficultyFrontier(time, parent)
	// case config.IsHomestead(next):
	// 	return calcDifficultyHomestead(time, parent)
	// default:
	// 	return calcDifficultyFrontier(time, parent)
	// }
}

// Some weird constants to avoid constant memory allocs for them.
var (
	// expDiffPeriod = big.NewInt(100000)
	expDiffPeriod = big.NewInt(100)
	big1          = big.NewInt(1)
	big2          = big.NewInt(2)
	big9          = big.NewInt(9)
	big10         = big.NewInt(10)
	bigMinus99    = big.NewInt(-99)
	big2999999    = big.NewInt(2999999)
)

// calcDifficultyByzantium is the difficulty adjustment algorithm. It returns
// the difficulty that a new block should have when created at time given the
// parent block's time and difficulty. The calculation uses the Byzantium rules.
func calcDifficultyByzantium(time uint64, parent *wire.Header) *big.Int {
	// https://github.com/ethereum/EIPs/issues/100.
	// algorithm:
	// diff = (parent_diff +
	//         (parent_diff / 2048 * max((2 if len(parent.uncles) else 1) - ((timestamp - parent.timestamp) // 9), -99))
	//        ) + 2^(periodCount - 2)

	bigTime := new(big.Int).SetUint64(time)
	bigParentTime := new(big.Int).SetInt64(parent.Timestamp)

	// holds intermediate values to make the algo easier to read & audit
	x := new(big.Int)
	y := new(big.Int)

	// (2 if len(parent_uncles) else 1) - (block_timestamp - parent_timestamp) // 9
	x.Sub(bigTime, bigParentTime)
	x.Div(x, big9)
	x.Sub(big1, x)

	// max((2 if len(parent_uncles) else 1) - (block_timestamp - parent_timestamp) // 9, -99)
	if x.Cmp(bigMinus99) < 0 {
		x.Set(bigMinus99)
	}
	// parent_diff + (parent_diff / 2048 * max((2 if len(parent.uncles) else 1) - ((timestamp - parent.timestamp) // 9), -99))
	y.Div(parent.Difficulty, params.DifficultyBoundDivisor)
	x.Mul(y, x)
	x.Add(parent.Difficulty, x)

	// minimum difficulty can ever be (before exponential factor)
	if x.Cmp(params.MinimumDifficulty) < 0 {
		x.Set(params.MinimumDifficulty)
	}

	// for the exponential factor
	periodCount := new(big.Int).Add(new(big.Int).SetUint64(parent.Number), big1)
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

// calcDifficultyHomestead is the difficulty adjustment algorithm. It returns
// the difficulty that a new block should have when created at time given the
// parent block's time and difficulty. The calculation uses the Homestead rules.
func calcDifficultyHomestead(time uint64, parent *wire.Header) *big.Int {
	// https://github.com/ethereum/EIPs/blob/master/EIPS/eip-2.md
	// algorithm:
	// diff = (parent_diff +
	//         (parent_diff / 2048 * max(1 - (block_timestamp - parent_timestamp) // 10, -99))
	//        ) + 2^(periodCount - 2)

	bigTime := new(big.Int).SetUint64(time)
	bigParentTime := new(big.Int).SetInt64(parent.Timestamp)

	// holds intermediate values to make the algo easier to read & audit
	x := new(big.Int)
	y := new(big.Int)

	// 1 - (block_timestamp - parent_timestamp) // 10
	x.Sub(bigTime, bigParentTime)
	x.Div(x, big10)
	x.Sub(big1, x)

	// max(1 - (block_timestamp - parent_timestamp) // 10, -99)
	if x.Cmp(bigMinus99) < 0 {
		x.Set(bigMinus99)
	}
	// (parent_diff + parent_diff // 2048 * max(1 - (block_timestamp - parent_timestamp) // 10, -99))
	y.Div(parent.Difficulty, params.DifficultyBoundDivisor)
	x.Mul(y, x)
	x.Add(parent.Difficulty, x)

	// minimum difficulty can ever be (before exponential factor)
	if x.Cmp(params.MinimumDifficulty) < 0 {
		x.Set(params.MinimumDifficulty)
	}
	// for the exponential factor
	periodCount := new(big.Int).Add(new(big.Int).SetUint64(parent.Number), big1)
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

// calcDifficultyFrontier is the difficulty adjustment algorithm. It returns the
// difficulty that a new block should have when created at time given the parent
// block's time and difficulty. The calculation uses the Frontier rules.
func calcDifficultyFrontier(time uint64, parent *wire.Header) *big.Int {
	diff := new(big.Int)
	adjust := new(big.Int).Div(parent.Difficulty, params.DifficultyBoundDivisor)
	bigTime := new(big.Int)
	bigParentTime := new(big.Int)

	bigTime.SetUint64(time)
	bigParentTime.Set(new(big.Int).SetInt64(parent.Timestamp))

	if bigTime.Sub(bigTime, bigParentTime).Cmp(params.DurationLimit) < 0 {
		diff.Add(parent.Difficulty, adjust)
	} else {
		diff.Sub(parent.Difficulty, adjust)
	}
	if diff.Cmp(params.MinimumDifficulty) < 0 {
		diff.Set(params.MinimumDifficulty)
	}

	periodCount := new(big.Int).Add(new(big.Int).SetUint64(parent.Number), big1)
	periodCount.Div(periodCount, expDiffPeriod)
	if periodCount.Cmp(big1) > 0 {
		// diff = diff + 2^(periodCount - 2)
		expDiff := periodCount.Sub(periodCount, big2)
		expDiff.Exp(big2, expDiff, nil)
		diff.Add(diff, expDiff)
		diff = math.BigMax(diff, params.MinimumDifficulty)
	}
	return diff
}

// VerifySeal checks whether the given block satisfies
// the PoW difficulty requirements.
func (ethash *Ethash) VerifySeal(chain c.ChainReader, header *wire.Header) error {
	// If we're running a fake PoW, accept any seal as valid
	if ethash.config.PowMode == ModeFake || ethash.config.PowMode == ModeFullFake {
		time.Sleep(ethash.fakeDelay)
		if ethash.fakeFail == header.Number {
			return errInvalidPoW
		}
		return nil
	}
	// If we're running a shared PoW, delegate verification to it
	if ethash.shared != nil {
		return ethash.shared.VerifySeal(chain, header)
	}
	// Ensure that we have a valid difficulty for the block
	if header.Difficulty.Sign() <= 0 {
		return errInvalidDifficulty
	}
	// Recompute the digest and PoW value and verify against the header
	number := header.Number

	cache := ethash.cache(number)
	size := datasetSize(number)
	if ethash.config.PowMode == ModeTest {
		size = 32 * 1024
	}
	digest, result := hashimotoLight(size, cache.cache, header.HashNoNonce().Bytes(), header.Nonce.Uint64())
	// Caches are unmapped in a finalizer. Ensure that the cache stays live
	// until after the call to hashimotoLight so it's not unmapped while being used.
	runtime.KeepAlive(cache)

	if !bytes.Equal(header.MixHash.Bytes(), digest) {
		return errInvalidMixDigest
	}
	target := new(big.Int).Div(maxUint256, header.Difficulty)
	if new(big.Int).SetBytes(result).Cmp(target) > 0 {
		return errInvalidPoW
	}
	return nil
}

// Prepare implements consensus.Engine, initializing the difficulty field of a
// header to conform to the ethash protocol. The changes are done inline.
func (ethash *Ethash) Prepare(chain c.ChainReader, header *wire.Header) error {

	// Get the header of the block's parent.
	parent, err := chain.GetHeaderByHash(header.ParentHash)
	if err != nil {
		if err != c.ErrBlockExists {
			return err
		}
		return ErrUnknownParent
	}

	header.Difficulty = ethash.CalcDifficulty(chain, uint64(header.Timestamp), parent)
	return nil
}

// Finalize accumulates rewards, computes the final state and assembling the block.
func (ethash *Ethash) Finalize(chain c.BlockMaker, block *wire.Block) (*wire.Block, error) {
	// TODO: accumulate rewards, recompute state and update block header
	return block, nil
}
