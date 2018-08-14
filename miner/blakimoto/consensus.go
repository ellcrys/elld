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

package blakimoto

import (
	"errors"
	"fmt"
	"math/big"
	"time"

	c "github.com/ellcrys/elld/blockchain/common"

	"github.com/ellcrys/elld/wire"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/params"
)

var (
	allowedFutureBlockTime = 15 * time.Second // Max time from current time allowed for blocks, before they're considered future blocks
)

// Various error messages to mark blocks invalid. These should be private to
// prevent engine specific errors from being referenced in the remainder of the
// codebase, inherently breaking if the engine is swapped out. Please put common
// error types into the consensus package.
var (
	errZeroBlockTime     = errors.New("timestamp equals parent's")
	errInvalidDifficulty = errors.New("non-positive difficulty")
	errInvalidMixDigest  = errors.New("invalid mix digest")
	errInvalidPoW        = errors.New("invalid proof-of-work")
)

// VerifyHeader checks whether a header conforms to the consensus rules
func (b *Blakimoto) VerifyHeader(chain c.ChainReader, header, parent *wire.Header, seal bool) error {

	// Ensure that the header's extra-data section is of a reasonable size
	if uint64(len(header.Extra)) > params.MaximumExtraDataSize {
		return fmt.Errorf("extra-data too long: %d > %d", len(header.Extra), params.MaximumExtraDataSize)
	}

	// Verify the header's timestamp
	if time.Unix(header.Timestamp, 0).After(time.Now().Add(allowedFutureBlockTime)) {
		return consensus.ErrFutureBlock
	}

	if header.Timestamp <= parent.Timestamp {
		return errZeroBlockTime
	}

	// Verify the block's difficulty based in it's timestamp and parent's difficulty
	expected := b.CalcDifficulty(chain, uint64(header.Timestamp), parent)
	if expected.Cmp(header.Difficulty) != 0 {
		return fmt.Errorf("invalid difficulty: have %v, want %v", header.Difficulty, expected)
	}

	// Verify that the block number is parent's +1
	if diff := header.Number - parent.Number; diff != 1 {
		return consensus.ErrInvalidNumber
	}

	// Verify the engine specific seal securing the block
	if seal {
		if err := b.VerifySeal(chain, header); err != nil {
			return err
		}
	}

	return nil
}

// CalcDifficulty is the difficulty adjustment algorithm. It returns
// the difficulty that a new block should have when created at time
// given the parent block's time and difficulty.
func (b *Blakimoto) CalcDifficulty(chain c.ChainReader, time uint64, parent *wire.Header) *big.Int {
	return CalcDifficulty(time, parent)
}

// CalcDifficulty is the difficulty adjustment algorithm. It returns
// the difficulty that a new block should have when created at time
// given the parent block's time and difficulty.
func CalcDifficulty(time uint64, parent *wire.Header) *big.Int {
	return calcDifficultyFrontier(time, parent)
}

// Some weird constants to avoid constant memory allocs for them.
var (
	// expDiffPeriod = big.NewInt(100000)
	expDiffPeriod = big.NewInt(3)
	big1          = big.NewInt(1)
	big2          = big.NewInt(2)
)

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
func (b *Blakimoto) VerifySeal(chain c.ChainReader, header *wire.Header) error {
	// If we're running a fake PoW, accept any seal as valid
	if b.config.PowMode == ModeTest {
		time.Sleep(b.fakeDelay)
		// if ethash.fakeFail == header.Number {
		// 	return errInvalidPoW
		// }
		return nil
	}

	// Ensure that we have a valid difficulty for the block
	if header.Difficulty.Sign() <= 0 {
		return errInvalidDifficulty
	}

	// Recompute the digest and PoW value and verify against the header
	result := blakimoto(header.HashNoNonce().Bytes(), header.Nonce.Uint64())

	target := new(big.Int).Div(maxUint256, header.Difficulty)
	if new(big.Int).SetBytes(result).Cmp(target) > 0 {
		return errInvalidPoW
	}
	return nil
}

// Prepare initializes the difficulty field of a
// header to conform to the protocol
func (b *Blakimoto) Prepare(chain c.ChainReader, header *wire.Header) error {

	// Get the header of the block's parent.
	parent, err := chain.GetHeaderByHash(header.ParentHash)
	if err != nil {
		if err != c.ErrBlockExists {
			return err
		}
		return ErrUnknownParent
	}

	header.Difficulty = b.CalcDifficulty(chain, uint64(header.Timestamp), parent)
	return nil
}

// Finalize accumulates rewards, computes the final state and assembling the block.
func (b *Blakimoto) Finalize(chain c.BlockMaker, block *wire.Block) (*wire.Block, error) {
	// TODO: accumulate rewards, recompute state and update block header
	return block, nil
}
