package miner

import (
	"errors"
	"fmt"
	"math/big"
	"runtime"

	ellBlock "github.com/ellcrys/elld/wire"
)

var (
	errLargeBlockTime         = errors.New("timestamp too big")
	errZeroBlockTime          = errors.New("timestamp equals parent's")
	errTooManyUncles          = errors.New("too many uncles")
	errDuplicateUncle         = errors.New("duplicate uncle")
	errUncleIsAncestor        = errors.New("uncle is ancestor")
	errDanglingUncle          = errors.New("uncle's parent is not ancestor")
	errInvalidDifficulty      = errors.New("non-positive difficulty")
	errInvalidMixDigest       = errors.New("invalid mix digest")
	errInvalidPoW             = errors.New("invalid proof-of-work")
	errNonPositiveBlockNumber = errors.New("non Positive Block Number")
)

// VerifyPOW checks whether the given block satisfies
// the PoW difficulty requirements.
func (ethash *Ethash) VerifyPOW(block *ellBlock.Block) error {

	// Ensure that we have a valid difficulty for the block
	blockDiffuclty, _ := new(big.Int).SetString(block.Difficulty, 10)
	if blockDiffuclty.Sign() <= 0 {
		return errInvalidDifficulty
	}

	// Recompute the digest and PoW value and verify against the header
	number := block.Number

	//block number must be a positive number
	if number <= 0 {
		return errNonPositiveBlockNumber
	}

	size := datasetSize(number)
	if ethash.config.PowMode == ModeTest {
		size = 32 * 1024
	}

	cache := ethash.cache(number)

	//get Digest and result for POW verification
	digest, result := hashimotoLight(size, cache.cache, block.HashNoNonce().Bytes(), block.Nounce)

	// Caches are unmapped in a finalizer. Ensure that the cache stays live
	// until after the call to hashimotoLight so it's not unmapped while being used.
	runtime.KeepAlive(cache)

	// convert digest to string
	outputDigest := fmt.Sprintf("%x", digest)

	// check if the mix digest is equivalent to the block Mix Digest
	if outputDigest != block.PowHash {
		return errInvalidMixDigest
	}

	target := new(big.Int).Div(maxUint256, blockDiffuclty)
	if new(big.Int).SetBytes(result).Cmp(target) > 0 {
		return errInvalidPoW
	}
	return nil
}
