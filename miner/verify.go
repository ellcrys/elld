package miner

import (
	"errors"
	"fmt"
	"math/big"

	ellBlock "github.com/ellcrys/druid/wire"
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

	blockNumber := block.Number

	epoch := blockNumber / epochLength
	currentI, _ := ethash.datasets.get(epoch)

	current := currentI.(*dataset)

	// Wait for generation to finish if need be.
	// cache and Dag file
	current.generate(ethash.config.DatasetDir, ethash.config.DatasetsOnDisk, ethash.config.PowMode == ModeTest)

	var (
		Mhash              = block.HashNoNonce().Bytes()
		blockDifficulty, _ = new(big.Int).SetString(block.Difficulty, 10)
		Mtarget            = new(big.Int).Div(maxUint256, blockDifficulty)
		Mdataset           = current
	)

	// Ensure that we have a valid difficulty for the block
	if blockDifficulty.Sign() <= 0 {
		return errInvalidDifficulty
	}

	// Recompute the digest and PoW value and verify against the header
	number := block.Number

	//block number must be a positive number
	if number <= 0 {
		return errNonPositiveBlockNumber
	}

	_, result := hashimotoFull(Mdataset.dataset, Mhash, block.Nounce)
	if new(big.Int).SetBytes(result).Cmp(Mtarget) <= 0 {
		return nil
	}

	return fmt.Errorf("invalid block")
}
